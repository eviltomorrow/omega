package cmd

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/eviltomorrow/omega/pkg/file"
	"github.com/eviltomorrow/omega/pkg/tools"
	"github.com/spf13/cobra"
	"github.com/tidwall/gjson"
)

var omega_install = &cobra.Command{
	Use:   "install",
	Short: "install omega-watchdog tool",
	Long:  "  \r\nomega-ctl install omega api support",
	Run: func(cmd *cobra.Command, args []string) {
		if path, err := apiWatchdogInstall(goroutines); err != nil {
			log.Printf("[E] Install watchdog failure, nest error: %v", err)
		} else {
			log.Printf("[I] Install complete, report: cat %s", path)
		}
	},
}

var omega_uninstall = &cobra.Command{
	Use:   "uninstall",
	Short: "uninstall omega-watchdog tool",
	Long:  "  \r\nomega-ctl uninstall omega api support",
	Run: func(cmd *cobra.Command, args []string) {

	},
}

var (
	imageDir   string
	goroutines = 3
)

func init() {
	root.AddCommand(omega_install)
	omega_install.Flags().StringVar(&imageDir, "image_dir", "image", "image location to install")
	omega_install.MarkFlagRequired("image_dir")

	// 未实现
	_ = omega_uninstall
}

func apiWatchdogInstall(num int) (string, error) {
	var (
		resourceFile = filepath.Join(imageDir, "resource.json")
		imageFile    = filepath.Join(imageDir, "omega.tar.gz")
		shellFile    = filepath.Join(imageDir, "omega-install.sh")
	)

	for _, path := range []string{imageFile, resourceFile, shellFile} {
		fi, err := os.Stat(path)
		if err != nil {
			return "", err
		}
		if fi.IsDir() {
			return "", fmt.Errorf("path is a dir, nest path: %v", path)
		}
	}

	f, err := os.Open(resourceFile)
	if err != nil {
		return "", err
	}
	defer f.Close()

	var (
		now     = time.Now()
		logPath = fmt.Sprintf("install-report-%s.log", now.Format("2006-01-02 15:04:05"))
	)
	logFile, err := os.OpenFile(logPath, os.O_CREATE|os.O_TRUNC|os.O_RDWR, 0644)
	if err != nil {
		return "", err
	}
	defer logFile.Close()

	var scanner = bufio.NewScanner(f)
	scanner.Split(bufio.ScanLines)

	var resources = make([]*resource, 0, 128)
loop:
	for scanner.Scan() {
		var line = strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		var r = &resource{}
		if err := json.Unmarshal([]byte(line), r); err != nil {
			log.Printf("[E] Unmarshal line to resource failure, nest error: %v", err)
			logFile.WriteString(fmt.Sprintf("[E] Unmarshal line to resource failure, nest error: %v\r\n", err))
			continue
		}
		for _, data := range []string{r.InnerIP, r.Username, r.HomeDir, r.Endpoints, r.GroupName} {
			if data == "" {
				log.Printf("[E] Invalid resource, nest resource: %v", r.String())
				logFile.WriteString(fmt.Sprintf("[E] Invalid resource, nest resource: %v\r\n", r.String()))
				continue loop
			}
		}
		if r.Password == "" && r.PrivateKeyPath == "" {
			log.Printf("[E] Invalid resource, nest resource: %v", r.String())
			logFile.WriteString(fmt.Sprintf("[E] Invalid resource, nest resource: %v\r\n", r.String()))
			continue
		}
		resources = append(resources, r)
	}

	var (
		p = make(chan *resource, 32)
		s = make(chan string, 32)
	)
	go func() {
		for _, r := range resources {
			p <- r
		}
		close(p)
	}()

	var signal = make(chan struct{}, 1)
	go func() {
		for r := range s {
			logFile.WriteString(r)
		}
		signal <- struct{}{}
	}()

	var wait sync.WaitGroup
	for i := 0; i < num; i++ {
		wait.Add(1)
		go func() {
			for r := range p {
				info, err := install(r, imageFile, shellFile)
				if err != nil {
					var msg = fmt.Sprintf("[E] Install image failure, resource: %v, nest error: %v", r.InnerIP, err)
					s <- msg
					log.Print(msg)
				} else {
					var msg = fmt.Sprintf("[I] Install image success, resource: %v, info: %v", r.InnerIP, info)
					s <- msg
					log.Print(msg)
				}
			}
			wait.Done()
		}()
	}
	wait.Wait()
	close(s)
	<-signal

	return logPath, nil
}

func install(r *resource, imageFile string, shellFile string) (string, error) {
	var host = r.InnerIP
	if r.OuterIP != "" {
		host = r.OuterIP
	}

	var pk []byte
	if r.PrivateKeyPath != "" {
		buf, err := ioutil.ReadFile(r.PrivateKeyPath)
		if err != nil {
			return "", err
		}
		pk = buf
	}
	scp, err := file.NewSCP(r.Username, r.Password, host, r.Port, pk, timeout)
	if err != nil {
		return "", err
	}
	defer scp.Close()

	var (
		imageName = filepath.Base(imageFile)
		shellName = filepath.Base(shellFile)
		imagePath = filepath.Join(r.HomeDir, imageName)
		shellPath = filepath.Join(r.HomeDir, shellName)
	)

	if err := scp.Upload(imageFile, imagePath); err != nil {
		return "", err
	}
	if err := scp.Upload(shellFile, shellPath); err != nil {
		return "", err
	}

	var c = fmt.Sprintf("cd %s; chmod a+x %s; ./%s -i %s -o %s -e %s -g %s", r.HomeDir, shellPath, shellName, r.InnerIP, r.OuterIP, r.Endpoints, r.GroupName)
	stdout, stderr, err := scp.Run(c, timeout)
	var buf bytes.Buffer
	buf.WriteString(tools.BytesToStringSlow(stdout))
	buf.WriteString(tools.BytesToStringSlow(stderr))
	if err != nil {
		return "", fmt.Errorf("error: %v, output: %v", err, buf.String())
	}
	if len(buf.Bytes()) != 0 {
		isJson := gjson.Valid(buf.String())
		if isJson {
			return buf.String(), nil
		}
		return "", fmt.Errorf("%s", buf.String())
	}
	return "", fmt.Errorf("[panic] stdout: nil")
}

type resource struct {
	OuterIP        string `json:"outer_ip"`
	InnerIP        string `json:"inner_ip"`
	Port           int    `json:"port"`
	Username       string `json:"username"`
	Password       string `json:"password"`
	PrivateKeyPath string `json:"private_key_path"`
	HomeDir        string `json:"home_dir"`
	Endpoints      string `json:"endpoints"`
	GroupName      string `json:"group_name"`
}

func (r *resource) String() string {
	buf, _ := json.Marshal(r)
	return string(buf)
}
