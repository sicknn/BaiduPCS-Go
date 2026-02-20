package pcscommand

import (
	"fmt"
	"os"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/qjfoidnh/BaiduPCS-Go/baidupcs"
	"github.com/qjfoidnh/BaiduPCS-Go/pcstable"
)

// RunShareSet 执行分享
func RunShareSet(paths []string, option *baidupcs.ShareOption) {
	pcspaths, err := matchPathByShellPattern(paths...)
	if err != nil {
		fmt.Println(err)
		return
	}

	shared, err := GetBaiduPCS().ShareSet(pcspaths, option)
	if err != nil {
		fmt.Printf("%s failed: %s\n", baidupcs.OperationShareSet, err)
		return
	}
	if option.IsCombined {
		fmt.Printf("shareID: %d, Link: %s?pwd=%s\n", shared.ShareID, shared.Link, shared.Pwd)
	} else {
		fmt.Printf("shareID: %d, Link: %s, Password: %s\n", shared.ShareID, shared.Link, shared.Pwd)
	}
}

// RunShareCancel 执行取消分享
func RunShareCancel(shareIDs []int64) {
	if len(shareIDs) == 0 {
		fmt.Printf("%s failed, no shareid provided\n", baidupcs.OperationShareCancel)
		return
	}

	err := GetBaiduPCS().ShareCancel(shareIDs)
	if err != nil {
		fmt.Printf("%s failed: %s\n", baidupcs.OperationShareCancel, err)
		return
	}

	fmt.Printf("%s succeeded\n", baidupcs.OperationShareCancel)
}

// RunShareList 执行列出分享列表
func RunShareList(page int) {
	if page < 1 {
		page = 1
	}

	pcs := GetBaiduPCS()
	records, err := pcs.ShareList(page)
	if err != nil {
		fmt.Printf("%s failed: %s\n", baidupcs.OperationShareList, err)
		return
	}

	tb := pcstable.NewTable(os.Stdout)
	tb.SetHeader([]string{"#", "ShareID", "Share Link", "Extraction Code", "Feature Dir", "Feature Path", "Expires", "Views"})
	for k, record := range records {
		if record.ExpireType == -1 {
			record.Valid = "Expired" // expired share
		} else {
			if record.ExpireTime == 0 {
				record.Valid = "Permanent"
			} else {
				tm := time.Unix(time.Now().Unix()+record.ExpireTime, 0)
				record.Valid = tm.Format("2006/01/02 15:04:05")

			}

		}
		// 获取Passwd
		if record.Public == 0 && record.ExpireType != -1 {
			// 私密分享
			info, pcsError := pcs.ShareSURLInfo(record.ShareID)
			if pcsError != nil {
				// 获取错误
				fmt.Printf("[%d] Failed to get share password: %s\n", k, pcsError)
			} else {
				record.Passwd = strings.TrimSpace(info.Pwd)
			}
		}

		tb.Append([]string{strconv.Itoa(k), strconv.FormatInt(record.ShareID, 10), record.Shortlink, record.Passwd, path.Clean(path.Dir(record.TypicalPath)), record.TypicalPath, record.Valid, strconv.Itoa(record.ViewCount)})
	}
	tb.Render()
}
