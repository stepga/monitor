package node

import (
	"encoding/json"
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
	FileSystems []FileSystem `json:"file_systems"`
}

func (nodeinfo NodeInfo) Report() string {
	return nodeinfo.String()
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

	info.HostName, err = HostName()
	if err != nil {
		return nil, err
	}

	info.OperatingSystemName = OperatingSystemName()

	info.OperatingSystemVersion, err = OperatingSystemVersion()
	if err != nil {
		return nil, err
	}

	info.FileSystems, err = FileSystems()
	if err != nil {
		return nil, err
	}

	info.RebootRequired, err = RebootRequired()
	if err != nil {
		return nil, err
	}

	return info, nil
}
