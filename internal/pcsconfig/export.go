package pcsconfig

import (
	"fmt"
	"github.com/olekukonko/tablewriter"
	"github.com/qjfoidnh/BaiduPCS-Go/baidupcs"
	"github.com/qjfoidnh/BaiduPCS-Go/pcstable"
	"github.com/qjfoidnh/BaiduPCS-Go/pcsutil/converter"
	"github.com/qjfoidnh/BaiduPCS-Go/requester"
	"os"
	"strconv"
)

// ActiveUser 获取当前登录的用户
func (c *PCSConfig) ActiveUser() *Baidu {
	if c.activeUser == nil {
		return &Baidu{}
	}
	return c.activeUser
}

// ActiveUserBaiduPCS 获取当前登录的用户的baidupcs.BaiduPCS
func (c *PCSConfig) ActiveUserBaiduPCS() *baidupcs.BaiduPCS {
	if c.pcs == nil {
		c.pcs = c.ActiveUser().BaiduPCS()
	}
	return c.pcs
}

func (c *PCSConfig) httpClientWithUA(ua string) *requester.HTTPClient {
	client := requester.NewHTTPClient()
	client.SetHTTPSecure(c.EnableHTTPS)
	client.SetUserAgent(ua)
	return client
}

// HTTPClient 返回设置好的 HTTPClient
func (c *PCSConfig) HTTPClient() *requester.HTTPClient {
	return c.httpClientWithUA(c.UserAgent)
}

// PCSHTTPClient 返回设置好的 PCS HTTPClient
func (c *PCSConfig) PCSHTTPClient() *requester.HTTPClient {
	return c.httpClientWithUA(c.PCSUA)
}

// PanHTTPClient 返回设置好的 Pan HTTPClient
func (c *PCSConfig) PanHTTPClient() *requester.HTTPClient {
	return c.httpClientWithUA(c.PanUA)
}

// NumLogins 获取登录的用户数量
func (c *PCSConfig) NumLogins() int {
	return len(c.BaiduUserList)
}

// AverageParallel 返回平均的下载最大并发量
func (c *PCSConfig) AverageParallel() int {
	return AverageParallel(c.MaxParallel, c.MaxDownloadLoad)
}

// PrintTable 输出表格
func (c *PCSConfig) PrintTable() {
	tb := pcstable.NewTable(os.Stdout)
	tb.SetHeader([]string{"Name", "Value", "Suggested", "Description"})
	tb.SetHeaderAlignment(tablewriter.ALIGN_LEFT)
	tb.SetColumnAlignment([]int{tablewriter.ALIGN_DEFAULT, tablewriter.ALIGN_LEFT, tablewriter.ALIGN_LEFT, tablewriter.ALIGN_LEFT})
	tb.AppendBulk([][]string{
		[]string{"appid", fmt.Sprint(c.AppID), "", "Baidu PCS app ID"},
		[]string{"cache_size", converter.ConvertFileSize(int64(c.CacheSize), 2), "1KB ~ 256KB", "Download cache size. Increase this if disk usage is high or download speed is slow"},
		[]string{"max_parallel", strconv.Itoa(c.MaxParallel), "1 ~ 20", "Maximum total download concurrency (non-SVIP users cannot exceed 1)"},
		[]string{"max_upload_parallel", strconv.Itoa(c.MaxUploadParallel), "1 ~ 100", "Maximum upload concurrency per file"},
		[]string{"max_download_load", strconv.Itoa(c.MaxDownloadLoad), "1 ~ 5", "Maximum number of files downloading simultaneously"},
		[]string{"max_download_rate", showMaxRate(c.MaxDownloadRate), "", "Maximum download speed limit, 0 means unlimited"},
		[]string{"max_upload_rate", showMaxRate(c.MaxUploadRate), "", "Maximum upload speed limit, 0 means unlimited"},
		[]string{"max_upload_load", strconv.Itoa(c.MaxUploadLoad), "1 ~ 4", "Maximum number of files uploading simultaneously"},
		[]string{"savedir", c.SaveDir, "", "Directory to save downloaded files"},
		[]string{"enable_https", fmt.Sprint(c.EnableHTTPS), "true", "Enable HTTPS"},
		[]string{"force_login_username", fmt.Sprint(c.ForceLogin), "empty", "Force login username; use only when tieba user-info API is unavailable"},
		[]string{"ignore_illegal", fmt.Sprint(c.IgnoreIllegal), "false", "Disable illegal-character check on upload filenames"},
		[]string{"upload_policy", fmt.Sprint(c.UPolicy), baidupcs.SkipPolicy, fmt.Sprintf("Policy for duplicate filenames: %s (default, skip), %s (overwrite), %s (skip unchanged-size files, overwrite others)",
			baidupcs.SkipPolicy, baidupcs.OverWritePolicy, baidupcs.RsyncPolicy)},
		[]string{"user_agent", c.UserAgent, requester.DefaultUserAgent, "Browser user-agent"},
		[]string{"pcs_ua", c.PCSUA, "", "PCS user-agent"},
		[]string{"pcs_addr", c.PCSAddr, "pcs.baidu.com", "PCS server address"},
		[]string{"fix_pcs_addr", fmt.Sprint(c.FixPCSAddr), "false", "Use static PCS server address instead of dynamic selection"},
		[]string{"pan_ua", c.PanUA, baidupcs.NetdiskUA, "Pan user-agent"},
		[]string{"proxy", c.Proxy, "", "Proxy setting, supports http/socks5"},
		[]string{"proxy_hostnames", c.ProxyHostnames, "", "Hostnames to proxy, comma-separated; empty means proxy all hostnames"},
		[]string{"local_addrs", c.LocalAddrs, "", "Local network interface addresses, comma-separated"},
	})
	tb.Render()
}
