// Package pcsupdate 更新包
package pcsupdate

import (
	"archive/zip"
	"bytes"
	"fmt"
	"github.com/qjfoidnh/BaiduPCS-Go/internal/pcsconfig"
	"github.com/qjfoidnh/BaiduPCS-Go/pcsliner"
	"github.com/qjfoidnh/BaiduPCS-Go/pcsutil"
	"github.com/qjfoidnh/BaiduPCS-Go/pcsutil/cachepool"
	"github.com/qjfoidnh/BaiduPCS-Go/pcsutil/checkaccess"
	"github.com/qjfoidnh/BaiduPCS-Go/pcsutil/converter"
	"github.com/qjfoidnh/BaiduPCS-Go/pcsutil/jsonhelper"
	"github.com/qjfoidnh/BaiduPCS-Go/requester/downloader"
	"github.com/qjfoidnh/BaiduPCS-Go/requester/rio"
	"github.com/qjfoidnh/BaiduPCS-Go/requester/transfer"
	"net/http"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"
)

const (
	// ReleaseName 分享根目录名称
	ReleaseName = "BaiduPCS-Go-releases"
)

type info struct {
	filename    string
	size        int64
	downloadURL string
}

// CheckUpdate 检测更新
func CheckUpdate(version string, yes bool) {
	if !checkaccess.AccessRDWR(pcsutil.ExecutablePath()) {
		fmt.Printf("Program directory is not writable, update cannot continue.\n")
		return
	}
	fmt.Println("Checking for updates, please wait...")
	c := pcsconfig.Config.HTTPClient()
	resp, err := c.Req(http.MethodGet, "https://api.github.com/repos/qjfoidnh/BaiduPCS-Go/releases/latest", nil, nil)
	if resp != nil {
		defer resp.Body.Close()
	}
	if err != nil {
		fmt.Printf("Failed to fetch data: %s\n", err)
		return
	}

	releaseInfo := ReleaseInfo{}
	err = jsonhelper.UnmarshalData(resp.Body, &releaseInfo)
	if err != nil {
		fmt.Printf("Failed to parse JSON data: %s\n", err)
		return
	}

	// 没有更新, 或忽略 Beta 版本, 和版本前缀不符的
	if strings.Contains(releaseInfo.TagName, "Beta") || !strings.HasPrefix(releaseInfo.TagName, "v") || version >= releaseInfo.TagName {
		fmt.Printf("No updates found.\n")
		return
	}

	fmt.Printf("New version found: %s\n", releaseInfo.TagName)

	line := pcsliner.NewLiner()
	defer line.Close()

	if !yes {
		y, err := line.State.Prompt("Proceed with update? (y/n): ")
		if err != nil {
			fmt.Printf("Input error: %s\n", err)
			return
		}

		if y != "y" && y != "Y" {
			fmt.Printf("Update cancelled.\n")
			return
		}
	}

	builder := &strings.Builder{}
	builder.WriteString("BaiduPCS-Go-" + releaseInfo.TagName + "-" + runtime.GOOS + "-.*?")

	switch runtime.GOARCH {
	case "amd64":
		builder.WriteString("(amd64|x86_64|x64)")
	case "386":
		builder.WriteString("(386|x86)")
	case "arm":
		builder.WriteString("(armv5|armv7|arm)")
	case "arm64":
		builder.WriteString("arm64")
	case "mips":
		builder.WriteString("mips")
	case "mips64":
		builder.WriteString("mips64")
	case "mipsle":
		builder.WriteString("(mipsle|mipsel)")
	case "mips64le":
		builder.WriteString("(mips64le|mips64el)")
	default:
		builder.WriteString(runtime.GOARCH)
	}

	builder.WriteString("\\.zip")

	exp := regexp.MustCompile(builder.String())

	var targetList []*info
	for _, asset := range releaseInfo.Assets {
		if asset == nil || asset.State != "uploaded" {
			continue
		}

		if exp.MatchString(asset.Name) {
			targetList = append(targetList, &info{
				filename:    asset.Name,
				size:        asset.Size,
				downloadURL: asset.BrowserDownloadURL,
			})
		}
	}

	var target info
	switch len(targetList) {
	case 0:
		fmt.Printf("No update file matches this system, GOOS: %s, GOARCH: %s\n", runtime.GOOS, runtime.GOARCH)
		return
	case 1:
		target = *targetList[0]
	default:
		fmt.Println()
		for k := range targetList {
			fmt.Printf("%d: %s\n", k, targetList[k].filename)
		}

		fmt.Println()
		t, err := line.State.Prompt("Enter index to download update: ")
		if err != nil {
			fmt.Printf("%s\n", err)
			return
		}

		i, err := strconv.Atoi(t)
		if err != nil {
			fmt.Printf("Input error: %s\n", err)
			return
		}

		if i < 0 || i >= len(targetList) {
			fmt.Printf("Input error: index out of range\n")
			return
		}

		target = *targetList[i]
	}

	if target.size > 0x7fffffff {
		fmt.Printf("file size too large: %d\n", target.size)
		return
	}

	fmt.Printf("Preparing to download update: %s\n", target.filename)

	// 开始下载
	buf := rio.NewBuffer(cachepool.RawMallocByteSlice(int(target.size)))
	der := downloader.NewDownloader(target.downloadURL, buf, &downloader.Config{
		MaxParallel: 20,
		CacheSize:   10000,
	})
	der.SetClient(c)

	der.OnDownloadStatusEvent(func(status transfer.DownloadStatuser, workersCallback func(downloader.RangeWorkerFunc)) {
		var leftStr string
		left := status.TimeLeft()
		if left < 0 {
			leftStr = "-"
		} else {
			leftStr = left.String()
		}

		fmt.Printf("\r ↓ %s/%s %s/s in %s, left %s ............",
			converter.ConvertFileSize(status.Downloaded(), 2),
			converter.ConvertFileSize(status.TotalSize(), 2),
			converter.ConvertFileSize(status.SpeedsPerSecond(), 2),
			status.TimeElapsed()/1e7*1e7, leftStr,
		)
	})
	der.OnFinish(func() {
		fmt.Println()
	})
	der.OnSuccess(func() {
		fmt.Printf("Download completed\n")
	})

	err = der.Execute()
	if err != nil {
		fmt.Printf("Download error: %s\n", err)
		return
	}

	// 读取文件
	reader, err := zip.NewReader(bytes.NewReader(buf.Bytes()), target.size)
	if err != nil {
		fmt.Printf("Failed to read update file: %s\n", err)
		return
	}

	execPath := pcsutil.ExecutablePath()

	var fileNum, errTimes int
	for _, zipFile := range reader.File {
		if zipFile == nil {
			continue
		}

		info := zipFile.FileInfo()

		if info.IsDir() {
			continue
		}

		rc, err := zipFile.Open()
		if err != nil {
			fmt.Printf("Failed to open zip entry: %s\n", err)
			continue
		}

		fileNum++

		name := zipFile.Name[strings.Index(zipFile.Name, "/")+1:]
		if name == "BaiduPCS-Go" {
			err = update(pcsutil.Executable(), rc)
		} else {
			err = update(filepath.Join(execPath, name), rc)
		}

		if err != nil {
			errTimes++
			fmt.Printf("Error processing zip path %s: %s\n", zipFile.Name, err)
			continue
		}
	}

	if errTimes == fileNum {
		fmt.Printf("Update failed\n")
		return
	}

	fmt.Printf("Update completed, please restart the program\n")
}
