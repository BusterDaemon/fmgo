package mounts

import (
	"github.com/fntlnz/mountinfo"
)

type mInfoZeroErr struct{}

func (e mInfoZeroErr) Error() string {
	return "No mount points are found."
}

func GetMounts() ([]string, error) {
	var (
		mInfo  []mountinfo.Mountinfo
		mounts []string
		err    error
	)
	if mInfo, err = mountinfo.GetMountInfo("/proc/self/mountinfo"); err != nil {
		return nil, err
	}

	for _, m := range mInfo {
		if m.FilesystemType != "tmpfs" &&
			m.FilesystemType != "fuse.portal" &&
			m.FilesystemType != "cgroup" &&
			m.FilesystemType != "sysfs" &&
			m.FilesystemType != "proc" &&
			m.FilesystemType != "devtmpfs" &&
			m.FilesystemType != "devpts" &&
			m.FilesystemType != "securityfs" &&
			m.FilesystemType != "efivarfs" &&
			m.FilesystemType != "cgroup2" {
			mounts = append(mounts, m.MountPoint)
		}
	}

	if len(mounts) == 0 {
		return nil, &mInfoZeroErr{}
	}

	return mounts, nil
}
