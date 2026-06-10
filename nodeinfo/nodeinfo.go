package nodeinfo

import (
	"encoding/json"

	"github.com/stepga/monitor/fs"
	"github.com/stepga/monitor/uname"
)

type NodeInfo struct {
	// as reported by `uname -n`
	HostName string `json:"host_name"`
	// as reported by `uname -s`
	OperatingSystemName string `json:"operating_system_name"`
	// as reported by `uname -r`
	OperatingSystemVersion string `json:"operating_system_version"`
	// e.g. due to linux system upgrade and indication via
	// /var/run/reboot-required
	RebootRequired bool `json:"reboot_required"`
	// an array of the mounted filesystems and their respective
	// used and total sizes in bytes
	FileSystems []fs.FileSystem `json:"file_systems"`
}

func (nodeinfo *NodeInfo) Marshal() ([]byte, error) {
	return json.Marshal(nodeinfo)
}

func (nodeinfo *NodeInfo) String() string {
	data, err := json.MarshalIndent(nodeinfo, "", "  ")
	if err != nil {
		return ""
	}
	return string(data)
}

func CreateInfo() (*NodeInfo, error) {
	var err error
	info := &NodeInfo{}

	info.HostName, err = uname.Hostname()
	if err != nil {
		return nil, err
	}

	info.OperatingSystemName = uname.OperatingSystemName()

	info.OperatingSystemVersion, err = uname.OperatingSystemVersion()
	if err != nil {
		return nil, err
	}

	info.FileSystems, err = fs.FileSystems()
	if err != nil {
		return nil, err
	}

	return info, nil
}
