package pcscommand

import (
	"fmt"
	"github.com/qjfoidnh/BaiduPCS-Go/pcstable"
	"os"
	"strconv"
)

// RunRemove 执行 批量删除文件/目录
func RunRemove(paths ...string) {
	paths, err := matchPathByShellPattern(paths...)
	if err != nil {
		fmt.Println(err)
		return
	}

	pnt := func() {
		tb := pcstable.NewTable(os.Stdout)
		tb.SetHeader([]string{"#", "File/Directory"})
		for k := range paths {
			tb.Append([]string{strconv.Itoa(k), paths[k]})
		}
		tb.Render()
	}

	err = GetBaiduPCS().Remove(paths...)
	if err != nil {
		fmt.Println(err)
		fmt.Println("Operation failed, the following files/directories failed to delete:")
		pnt()
		return
	}

	fmt.Println("Operation succeeded, the following files/directories were deleted (recoverable from recycle bin):")
	pnt()
}

// RunMkdir 执行 创建目录
func RunMkdir(path string) {
	activeUser := GetActiveUser()
	err := GetBaiduPCS().Mkdir(activeUser.PathJoin(path))
	if err != nil {
		fmt.Printf("Failed to create directory %s, %s\n", path, err)
		return
	}

	fmt.Println("Directory created successfully:", path)
}
