package pcsdownload

import "errors"

var (
	// ErrDownloadNotSupportChecksum 文件不支持校验
	ErrDownloadNotSupportChecksum = errors.New("this file does not support checksum verification")
	// ErrDownloadChecksumFailed 文件校验失败
	ErrDownloadChecksumFailed = errors.New("checksum verification failed, file md5 does not match server record")
	// ErrDownloadFileBanned 违规文件
	ErrDownloadFileBanned = errors.New("this file may violate policy and does not support verification")
	// ErrDlinkNotFound 未取得下载链接
	ErrDlinkNotFound = errors.New("download link not found")
	// ErrShareInfoNotFound 未在已分享列表中找到分享信息
	ErrShareInfoNotFound = errors.New("share info not found in share list")
)
