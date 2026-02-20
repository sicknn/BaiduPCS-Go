package pcscommand

import (
	"fmt"
	"github.com/qjfoidnh/BaiduPCS-Go/baidupcs"
	"github.com/qjfoidnh/BaiduPCS-Go/internal/pcsconfig"
)

// RunChangeDirectory 执行更改工作目录
func RunChangeDirectory(targetPath string, isList bool) {
	pcs := GetBaiduPCS()
	err := matchPathByShellPatternOnce(&targetPath)
	if err != nil {
		fmt.Println(err)
		return
	}

	data, err := pcs.FilesDirectoriesMeta(targetPath)
	if err != nil {
		fmt.Println(err)
		return
	}

	if !data.Isdir {
		fmt.Printf("Error: %s is not a directory\n", targetPath)
		return
	}

	GetActiveUser().Workdir = targetPath
	pcsconfig.Config.Save()

	fmt.Printf("Changed working directory: %s\n", targetPath)

	if isList {
		RunLs(".", nil, baidupcs.DefaultOrderOptions)
	}
}
