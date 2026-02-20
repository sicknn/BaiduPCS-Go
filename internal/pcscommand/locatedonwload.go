package pcscommand

import (
	"fmt"
	"github.com/qjfoidnh/BaiduPCS-Go/internal/pcsconfig"
	"github.com/qjfoidnh/BaiduPCS-Go/pcstable"
	"github.com/qjfoidnh/BaiduPCS-Go/pcsutil/converter"
	"os"
	"strconv"
)

// RunLocateDownload 执行获取直链
func RunLocateDownload(pcspaths []string, opt *LocateDownloadOption) {
	if opt == nil {
		opt = &LocateDownloadOption{}
	}

	absPaths, err := matchPathByShellPattern(pcspaths...)
	if err != nil {
		fmt.Println(err)
		return
	}

	pcs := GetBaiduPCS()

	if opt.FromPan {
		fds, err := pcs.FilesDirectoriesBatchMeta(absPaths...)
		if err != nil {
			fmt.Printf("%s\n", err)
			return
		}

		fidList := make([]int64, 0, len(fds))
		for i := range fds {
			fidList = append(fidList, fds[i].FsID)
		}

		list, err := pcs.LocatePanAPIDownload(fidList...)
		if err != nil {
			fmt.Printf("%s\n", err)
			return
		}

		tb := pcstable.NewTable(os.Stdout)
		tb.SetHeader([]string{"#", "fs_id", "Path", "Link"})

		var (
			i          int
			fidStrList = converter.SliceInt64ToString(fidList)
		)
		for k := range fidStrList {
			for i = range list {
				if fidStrList[k] == list[i].FsID {
					tb.Append([]string{strconv.Itoa(k), list[i].FsID, fds[k].Path, list[i].Dlink})
					list = append(list[:i], list[i+1:]...)
					break
				}
			}
		}
		tb.Render()
		fmt.Printf("\nNote: the links above are not directly accessible; you must be logged in to a Baidu account to download\n")
		return
	}

	for i, pcspath := range absPaths {
		info, err := pcs.LocateDownload(pcspath)
		if err != nil {
			fmt.Printf("[%d] %s, path: %s\n", i, err, pcspath)
			continue
		}

		fmt.Printf("[%d] %s: \n", i, pcspath)
		tb := pcstable.NewTable(os.Stdout)
		tb.SetHeader([]string{"#", "Link"})
		for k, u := range info.URLStrings(pcsconfig.Config.EnableHTTPS) {
			tb.Append([]string{strconv.Itoa(k), u.String()})
		}
		tb.Render()
		fmt.Println()
	}
	fmt.Printf("Tip: to access download links, set your downloader User-Agent to: %s\n", pcsconfig.Config.PanUA)
}
