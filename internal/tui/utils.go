package tui

import (
	"fmt"
	"io/fs"
	"syscall"
)

func GetFileOwner(fInfo fs.FileInfo) string {
	data := fInfo.Sys().(*syscall.Stat_t)
	if data == nil {
		return "Undefined"
	}

	return fmt.Sprintf(
		"%d:%d",
		int(data.Gid),
		int(data.Uid),
	)
}
