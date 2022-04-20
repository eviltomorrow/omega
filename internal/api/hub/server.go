package hub

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/eviltomorrow/omega/internal/api/hub/pb"
	"github.com/eviltomorrow/omega/pkg/file"
	"github.com/eviltomorrow/omega/pkg/lock"
	"github.com/eviltomorrow/omega/pkg/zlog"
	"github.com/hashicorp/go-version"
	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

var (
	BinDir        = "../bin"
	ImageDir      = "../var/images"
	ImageLockFile = "../var/images/.lock"
)

type Server struct {
	pb.UnimplementedHubServer
}

// Pull(*wrapperspb.StringValue, Hub_PullServer) error
// List(context.Context, *emptypb.Empty) (*ImageDesc, error)
// Del(context.Context, *wrapperspb.StringValue) (*wrapperspb.StringValue, error)
// Add(Hub_AddServer) error

func (s *Server) Pull(req *wrapperspb.StringValue, ps pb.Hub_PullServer) error {
	var path string
	if req.Value == "latest" {
		fis, err := ioutil.ReadDir(ImageDir)
		if err != nil {
			return err
		}

		versions := make([]*version.Version, 0, len(fis))
		for _, fi := range fis {
			if !fi.IsDir() {
				continue
			}

			name := fi.Name()
			if strings.HasPrefix(name, "v") {
				version, err := version.NewVersion(name)
				if err != nil {
					zlog.Error("Load version from dir failure", zap.String("name", name), zap.Error(err))
				} else {
					versions = append(versions, version)
				}
			}
		}
		if len(versions) == 0 {
			return fmt.Errorf("image repository is empty")
		}

		sort.Sort(version.Collection(versions))
		path = filepath.Join(ImageDir, versions[len(versions)-1].Original(), "omega")
	} else {
		version, err := genVersion(req.Value)
		if err != nil {
			return err
		}
		path = filepath.Join(ImageDir, version.Original(), "omega")
	}

	_, pipe, signal, err := file.Read(path)
	if err != nil {
		return err
	}
loop:
	for {
		select {
		case buf, ok := <-pipe:
			if !ok {
				break loop
			}

			if err := ps.Send(&pb.Image{Buf: buf}); err != nil {
				return err
			}
		case err := <-signal:
			return err
		}
	}

	return nil
}

func (s *Server) List(ctx context.Context, _ *emptypb.Empty) (*pb.ImageDesc, error) {
	fl, err := lock.CreateFileLock(ImageLockFile)
	if err != nil {
		return nil, err
	}
	defer lock.DestroyFileLock(fl)

	fis, err := ioutil.ReadDir(ImageDir)
	if err != nil {
		return nil, err
	}

	versions := make([]*version.Version, 0, len(fis))
	for _, fi := range fis {
		if !fi.IsDir() {
			continue
		}

		name := fi.Name()
		if strings.HasPrefix(name, "v") {
			version, err := version.NewVersion(name)
			if err != nil {
				zlog.Error("Load version from dir failure", zap.String("name", name), zap.Error(err))
			} else {
				versions = append(versions, version)
			}
		}
	}
	if len(versions) == 0 {
		return nil, fmt.Errorf("image repository is empty")
	}
	sort.Sort(version.Collection(versions))

	var id = &pb.ImageDesc{
		Images: make([]*pb.Image, 0, len(versions)),
	}
	for _, version := range versions {
		buf, _ := ioutil.ReadFile(filepath.Join(ImageDir, version.Original(), "release.md"))
		var d = &desc{}
		_ = json.Unmarshal(buf, d)

		var image = &pb.Image{
			ReleaseNotes: d.ReleaseNote,
			Tag:          version.Original(),
			Md5:          d.Md5,
			CreateTime:   d.CreateTime,
		}
		id.Images = append(id.Images, image)
	}
	return id, nil
}

func (s *Server) Del(ctx context.Context, req *wrapperspb.StringValue) (*wrapperspb.StringValue, error) {
	version, err := genVersion(req.Value)
	if err != nil {
		return nil, err
	}

	fl, err := lock.CreateFileLock(ImageLockFile)
	if err != nil {
		return nil, err
	}
	defer lock.DestroyFileLock(fl)

	path := filepath.Join(ImageDir, version.Original())
	_, err = os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("image not exist")
		}
		return nil, err
	}

	buf, _ := ioutil.ReadFile(filepath.Join(path, "release.md"))
	var d = &desc{}
	_ = json.Unmarshal(buf, d)

	if err := os.RemoveAll(path); err != nil {
		return nil, err
	}
	return &wrapperspb.StringValue{Value: d.Md5}, nil
}

func (s *Server) Push(as pb.Hub_PushServer) error {
	fi, err := os.Stat(ImageDir)
	if err != nil {
		return err
	}
	if !fi.IsDir() {
		return fmt.Errorf("%s is a file", ImageDir)
	}

	image, err := as.Recv()
	if err != nil {
		return err
	}

	fl, err := lock.CreateFileLock(ImageLockFile)
	if err != nil {
		return err
	}
	defer lock.DestroyFileLock(fl)

	version, err := genVersion(image.Tag)
	if err != nil {
		return err
	}
	var (
		releaseNote = image.ReleaseNotes
		createTime  = image.CreateTime
		tag         = version.Original()
		base        = filepath.Join(ImageDir, tag)
		name        = filepath.Join(base, "omega")
		ok          bool
	)

	if err := os.MkdirAll(base, 0700); err != nil {
		return err
	}
	defer func() {
		if !ok {
			os.RemoveAll(base)
		}
	}()

	var (
		data = make(chan []byte, 128)
		sig  = make(chan error)
	)
	go func() {
		for {
			req, err := as.Recv()
			if err == io.EOF {
				break
			}
			if err != nil {
				sig <- err
				break
			}
			if len(req.Buf) != 0 {
				data <- req.Buf
			}
		}
		close(data)
	}()
	if err := file.Write(name, 0744, data, sig); err != nil {
		return err
	}

	md5, err := file.CalculateMD5(name)
	if err != nil {
		return err
	}
	var (
		d      = &desc{ReleaseNote: releaseNote, Md5: md5, CreateTime: createTime}
		buf, _ = json.Marshal(d)
	)
	if err := ioutil.WriteFile(filepath.Join(base, "release.md"), buf, 0644); err != nil {
		return err
	}

	ok = true
	return as.SendAndClose(&wrapperspb.StringValue{Value: md5})
}

type desc struct {
	ReleaseNote string `json:"release_note"`
	Md5         string `json:"md5"`
	CreateTime  string `json:"create_time"`
}

func genVersion(v string) (*version.Version, error) {
	if v == "" || !strings.HasPrefix(v, "v") || strings.Count(v, ".") != 2 {
		return nil, fmt.Errorf("verify tag failure, expect: v3.x.x, actual: %v", v)
	}
	version, err := version.NewVersion(v)
	if err != nil {
		return nil, err
	}
	return version, nil
}
