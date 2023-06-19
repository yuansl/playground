package main

import (
	"encoding/json"
	"fmt"
)

type DownloadLog struct {
	Schema     string `json:"_schema"`     // schema名称，对于这个就是DownloadLog
	Local      string `json:"_loc"`        // 用户本地IP
	OS         string `json:"os"`          // 操作系统 名称+版本
	App        string `json:"app"`         // App 名称+版本；Web端使用当前站点的域名，Native端传bundleId
	SDK        string `json:"sdk"`         // SDK 名称+版本
	DevModel   string `json:"dev_model"`   // 设备型号
	DevID      string `json:"dev_id"`      // 设备唯一ID; 设备唯一ID，Web端暂定传空，Native端传设备ID
	ReqID      string `json:"r_id"`        // 请求ID
	XLocal     string `json:"loc" `        // 用户本地IP
	IP         string `json:"ip"`          // 节点IP
	Domain     string `json:"domain"`      // 节点域名
	StatusCode int64  `json:"status_code"` // 状态码
	ErrorMsg   string `json:"err_msg"`     // 错误信息，如TimeoutException、UnknownHostException等，以及不可恢复的error
	ErrorDesc  string `json:"err_desc"`    // 错误信息描述，错误信息的更多内容描述
	RangeStart int64  `json:"range_st"`    // Range Start
	RangeEnd   int64  `json:"range_end"`   // Range End
	RetryTimes int64  `json:"retry"`       // 重试次数
	FileTaskID string `json:"ftask_id"`    // 同一个文件请求的批次ID
	// 耗时情况：域名解析->TCP连接建立->TLS链接建立->多次重定向->准备发送请求消息->返回第一个字节->请求操作完成
	// DNS域名解析类型：HTTPDNS,HTTPDNS_CACHE,LOCALDNS,LOCALDNS_CACHE；目前DNS失效后是否会走LOCALDNS待定
	// 第一次请求成功是HTTPDNS_CACHE
	// 第一次请求失败走LOCALDNS
	// 有HTTPDNS_CACHE且HTTPDNS_CACHE不过期走HTTPDNS_CACHE
	// HTTPDNS_CACHE过期走过期的HTTPDNS_CACHE然后协程请求HTTPDNS
	NameLookupType      bool    `json:"lookup_type"`      // DNS域名解析类型,true为HTTPDNS,false为LOCALDNS
	TimeNameLookup      float64 `json:"t_lookup"`         // DNS域名解析耗时,0为缓存使用
	TimeConnect         float64 `json:"t_conn"`           // 连接建立耗时
	TimeTlsConnect      float64 `json:"t_tls"`            // TLS链接建立耗时
	TimeStartTransfer   float64 `json:"t_st_trans"`       // 返回第一个字节耗时
	TimeContentTransfer float64 `json:"t_content_trans" ` // 请求数据Content耗时
	TimeTotal           float64 `json:"t_total"`          // 请求总耗时
	RespBodySize        int     `json:"resp_size"`        // 请求到的数据总量，用于计算下载速度
}

func main() {
	var download DownloadLog

	data, _ := json.Marshal(download)

	fmt.Printf("%s\n", data)
}
