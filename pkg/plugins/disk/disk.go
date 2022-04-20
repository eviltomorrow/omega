package disk

import (
	"fmt"
	"strings"
	"time"

	"github.com/eviltomorrow/omega"
	"github.com/eviltomorrow/omega/metric"
	"github.com/eviltomorrow/omega/pkg/plugins"
)

type DiskStats struct {
	ps plugins.PS

	// Legacy support
	LegacyMountPoints []string `json:"-"`

	MountPoints []string `json:"mount_points"`
	IgnoreFS    []string `json:"ignore_fs"`
}

func (ds *DiskStats) Description() string {
	return "Read metrics about disk usage by mount point"
}

var diskSampleConfig = `
  ## By default stats will be gathered for all mount points.
  ## Set mount_points will restrict the stats to only the specified mount points.
  # mount_points = ["/"]

  ## Ignore mount points by filesystem type.
  ignore_fs = ["tmpfs", "devtmpfs", "devfs", "iso9660", "overlay", "aufs", "squashfs"]
`

func (ds *DiskStats) SampleConfig() string {
	return diskSampleConfig
}

func (ds *DiskStats) Init() error {
	// Legacy support:
	if len(ds.LegacyMountPoints) != 0 {
		ds.MountPoints = ds.LegacyMountPoints
	}

	ps := plugins.NewSystemPS()
	ds.ps = ps

	return nil
}

func (ds *DiskStats) Gather() ([]omega.Metric, error) {
	disks, partitions, err := ds.ps.DiskUsage(ds.MountPoints, ds.IgnoreFS)
	if err != nil {
		return nil, fmt.Errorf("getting disk usage info failure, nest error: %v", err)
	}

	var (
		metrics = make([]omega.Metric, 0, 64)
		now     = time.Now()
	)
	for i, du := range disks {
		if du.Total == 0 {
			// Skip dummy filesystem (procfs, cgroupfs, ...)
			continue
		}
		mountOpts := MountOptions(partitions[i].Opts)
		tags := map[string]string{
			"path":   du.Path,
			"device": strings.Replace(partitions[i].Device, "/dev/", "", -1),
			"fstype": du.Fstype,
			"mode":   mountOpts.Mode(),
		}
		var usedPercent float64
		if du.Used+du.Free > 0 {
			usedPercent = float64(du.Used) /
				(float64(du.Used) + float64(du.Free)) * 100
		}

		fields := map[string]interface{}{
			"total":        du.Total,
			"free":         du.Free,
			"used":         du.Used,
			"used_percent": usedPercent,
			"inodes_total": du.InodesTotal,
			"inodes_free":  du.InodesFree,
			"inodes_used":  du.InodesUsed,
		}
		metrics = append(metrics, metric.New("disk", tags, fields, now, omega.Gauge))
	}

	return metrics, nil
}

type MountOptions []string

func (opts MountOptions) Mode() string {
	if opts.exists("rw") {
		return "rw"
	} else if opts.exists("ro") {
		return "ro"
	} else {
		return "unknown"
	}
}

func (opts MountOptions) exists(opt string) bool {
	for _, o := range opts {
		if o == opt {
			return true
		}
	}
	return false
}

func (d *DiskStats) Config(conf map[string]interface{}) error {
	return nil
}

func init() {
	plugins.Register("disk", &DiskStats{
		ps: plugins.NewSystemPS(),
	})
}
