package pcscommand

import (
	"fmt"
	"github.com/qjfoidnh/BaiduPCS-Go/baidupcs"
	"github.com/qjfoidnh/BaiduPCS-Go/internal/pcsconfig"
	"github.com/qjfoidnh/BaiduPCS-Go/internal/pcsfunctions/pcsupload"
	"github.com/qjfoidnh/BaiduPCS-Go/pcstable"
	"github.com/qjfoidnh/BaiduPCS-Go/pcsutil"
	"github.com/qjfoidnh/BaiduPCS-Go/pcsutil/checksum"
	"github.com/qjfoidnh/BaiduPCS-Go/pcsutil/converter"
	"github.com/qjfoidnh/BaiduPCS-Go/pcsutil/taskframework"
	"os"
	"path"
	"path/filepath"
	"strings"
)

const (
	// DefaultUploadMaxRetry 默认上传失败最大重试次数
	DefaultUploadMaxRetry = 3
)

type (
	// UploadOptions 上传可选项
	UploadOptions struct {
		Parallel        int
		MaxRetry        int
		Load            int
		NoRapidUpload   bool
		NoSplitFile     bool   // 禁用分片上传
		Policy          string // 同名文件处理策略
		NoFilenameCheck bool   // 禁用文件名合法性检查
	}
)

func uploadPrintFormat(load int) string {
	if load <= 1 {
		return pcsupload.DefaultPrintFormat
	}
	return "[%s] ↑ %s/%s %s/s in %s ...\n"
}

// RunUpload 执行文件上传
func RunUpload(localPaths []string, savePath string, opt *UploadOptions) {
	if opt == nil {
		opt = &UploadOptions{}
	}

	// 检测opt
	if opt.Parallel <= 0 {
		opt.Parallel = pcsconfig.Config.MaxUploadParallel
	}

	opt.NoFilenameCheck = pcsconfig.Config.IgnoreIllegal

	if opt.MaxRetry < 0 {
		opt.MaxRetry = DefaultUploadMaxRetry
	}

	if opt.Load <= 0 {
		opt.Load = pcsconfig.Config.MaxUploadLoad
	}

	if opt.Policy != baidupcs.SkipPolicy && opt.Policy != baidupcs.OverWritePolicy && opt.Policy != baidupcs.RsyncPolicy {
		opt.Policy = pcsconfig.Config.UPolicy
	}

	err := matchPathByShellPatternOnce(&savePath)
	if err != nil {
		fmt.Printf("Warning: failed to resolve remote path %s for upload, %s\n", savePath, err)
	}

	switch len(localPaths) {
	case 0:
		fmt.Printf("Local path is empty\n")
		return
	}

	// 打开上传状态
	uploadDatabase, err := pcsupload.NewUploadingDatabase()
	if err != nil {
		fmt.Printf("Failed to open unfinished upload database: %s\n", err)
		return
	}
	defer uploadDatabase.Close()

	var (
		pcs = GetBaiduPCS()
		// 使用 task framework
		executor = &taskframework.TaskExecutor{
			IsFailedDeque: true, // 失败统计
		}
		subSavePath string
		// 统计
		statistic = &pcsupload.UploadStatistic{}
	)
	fmt.Print("\n")
	fmt.Printf("[0] Tip: max concurrency per file: %d, max concurrent uploading files: %d\n", opt.Parallel, opt.Load)

	statistic.StartTimer() // 开始计时

	LoadCount := 0

	for k := range localPaths {
		walkedFiles, err := pcsutil.WalkDir(localPaths[k], "")
		if err != nil {
			fmt.Printf("Warning: traversal error: %s\n", err)
			continue
		}

		for k3 := range walkedFiles {
			var localPathDir string
			// 针对 windows 的目录处理
			if os.PathSeparator == '\\' {
				walkedFiles[k3] = pcsutil.ConvertToUnixPathSeparator(walkedFiles[k3])
				localPathDir = pcsutil.ConvertToUnixPathSeparator(filepath.Dir(localPaths[k]))
			} else {
				localPathDir = filepath.Dir(localPaths[k])
			}

			// 避免去除文件名开头的"."
			if localPathDir == "." {
				localPathDir = ""
			}
			if len(localPaths) == 1 && len(walkedFiles) == 1 {
				opt.Load = 1
			}
			subSavePath = strings.TrimPrefix(walkedFiles[k3], localPathDir)
			if !opt.NoFilenameCheck && !pcsutil.ChPathLegal(walkedFiles[k3]) {
				fmt.Printf("[0] %s contains illegal path characters, skipped\n", walkedFiles[k3])
				continue
			}
			LoadCount++
			info := executor.Append(&pcsupload.UploadTaskUnit{
				LocalFileChecksum: checksum.NewLocalFileChecksum(walkedFiles[k3], int(baidupcs.SliceMD5Size)),
				SavePath:          path.Clean(savePath + baidupcs.PathSeparator + subSavePath),
				PCS:               pcs,
				UploadingDatabase: uploadDatabase,
				Parallel:          opt.Parallel,
				PrintFormat:       uploadPrintFormat(opt.Load),
				NoRapidUpload:     opt.NoRapidUpload,
				NoSplitFile:       opt.NoSplitFile,
				UploadStatistic:   statistic,
				Policy:            opt.Policy,
			}, opt.MaxRetry)
			if LoadCount >= opt.Load {
				LoadCount = opt.Load
			}
			fmt.Printf("[%s] Added to upload queue: %s\n", info.Id(), walkedFiles[k3])
		}
	}

	// 没有添加任何任务
	if executor.Count() == 0 {
		fmt.Printf("No files detected for upload.\n")
		return
	}

	// 设置上传文件并发数
	executor.SetParallel(LoadCount)
	// 执行上传任务
	executor.Execute()

	fmt.Printf("\n")
	fmt.Printf("Upload finished, elapsed: %s, total size: %s\n", statistic.Elapsed()/1e6*1e6, converter.ConvertFileSize(statistic.TotalSize()))

	// 输出上传失败的文件列表
	failedList := executor.FailedDeque()
	if failedList.Size() != 0 {
		fmt.Printf("The following files failed to upload:\n")
		tb := pcstable.NewTable(os.Stdout)
		for e := failedList.Shift(); e != nil; e = failedList.Shift() {
			item := e.(*taskframework.TaskInfoItem)
			tb.Append([]string{item.Info.Id(), item.Unit.(*pcsupload.UploadTaskUnit).LocalFileChecksum.Path})
		}
		tb.Render()
	}
}
