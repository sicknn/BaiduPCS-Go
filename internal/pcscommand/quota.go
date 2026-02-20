package pcscommand

import (
	"fmt"
	"github.com/qjfoidnh/BaiduPCS-Go/pcsutil/converter"
)

// RunGetQuota 执行 获取当前用户空间配额信息, 并输出
func RunGetQuota() {
	quota, used, err := GetBaiduPCS().QuotaInfo()
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Printf("Username: %s, Total: %s, Used: %s, Usage: %f%%\n",
		GetActiveUser().Name,
		converter.ConvertFileSize(quota),
		converter.ConvertFileSize(used),
		100*float64(used)/float64(quota),
	)
}
