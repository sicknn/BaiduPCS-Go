package downloader

import (
	"github.com/qjfoidnh/BaiduPCS-Go/requester/transfer"
)

type (
	//WorkerStatuser 状态
	WorkerStatuser interface {
		StatusCode() StatusCode //状态码
		StatusText() string
	}

	//StatusCode 状态码
	StatusCode int

	//WorkerStatus worker状态
	WorkerStatus struct {
		statusCode StatusCode
	}

	// DownloadStatusFunc 下载状态处理函数
	DownloadStatusFunc func(status transfer.DownloadStatuser, workersCallback func(RangeWorkerFunc))
)

const (
	//StatusCodeInit 初始化
	StatusCodeInit StatusCode = iota
	//StatusCodeSucceeded 成功
	StatusCodeSucceeded
	//StatusCodePending 等待响应
	StatusCodePending
	//StatusCodeDownloading 下载中
	StatusCodeDownloading
	//StatusCodeWaitToWrite 等待写入数据
	StatusCodeWaitToWrite
	//StatusCodeInternalError 内部错误
	StatusCodeInternalError
	//StatusCodeTooManyConnections 连接数太多
	StatusCodeTooManyConnections
	//StatusCodeNetError 网络错误
	StatusCodeNetError
	//StatusCodeFailed 下载失败
	StatusCodeFailed
	//StatusCodePaused 已暂停
	StatusCodePaused
	//StatusCodeReseted 已重设连接
	StatusCodeReseted
	//StatusCodeCanceled 已取消
	StatusCodeCanceled
)

//GetStatusText 根据状态码获取状态信息
func GetStatusText(sc StatusCode) string {
	switch sc {
	case StatusCodeInit:
		return "Initialized"
	case StatusCodeSucceeded:
		return "Succeeded"
	case StatusCodePending:
		return "Waiting for response"
	case StatusCodeDownloading:
		return "Downloading"
	case StatusCodeWaitToWrite:
		return "Waiting to write data"
	case StatusCodeInternalError:
		return "Internal error"
	case StatusCodeTooManyConnections:
		return "Too many connections"
	case StatusCodeNetError:
		return "Network error"
	case StatusCodeFailed:
		return "Failed"
	case StatusCodePaused:
		return "Paused"
	case StatusCodeReseted:
		return "Connection reset"
	case StatusCodeCanceled:
		return "Canceled"
	default:
		return "Unknown status code"
	}
}

//NewWorkerStatus 初始化WorkerStatus
func NewWorkerStatus() *WorkerStatus {
	return &WorkerStatus{
		statusCode: StatusCodeInit,
	}
}

//SetStatusCode 设置worker状态码
func (ws *WorkerStatus) SetStatusCode(sc StatusCode) {
	ws.statusCode = sc
}

//StatusCode 返回状态码
func (ws *WorkerStatus) StatusCode() StatusCode {
	return ws.statusCode
}

//StatusText 返回状态信息
func (ws *WorkerStatus) StatusText() string {
	return GetStatusText(ws.statusCode)
}
