package main

import (
	"encoding/json"
	"testing"

	"github.com/qbox/net-deftones/cmd/cdnlogetl/sinker"
)

func Test_main(t *testing.T) {
	var bytes = `{"client_ip":"183.221.46.163","content_type":"video/mp4","domain":"v5-hl-bd-qn-tt-lite.toutiaovod.com","url":"https://v5-hl-bd-qn-tt-lite.toutiaovod.com/ea22d0bc870c951e4d5d7a017f2eb8ef/65fe219d/video/tos/cn/tos-cn-vd-0026/4ef98948fea54a5d8652feb02f0aba62/media-video-hvc1/?a=35&ch=0&cr=7&dr=0&er=1&lr=default&cd=0%7C0%7C0%7C1&cv=1&br=702&bt=702&cs=4&ds=3&mime_type=video_mp4&qs=0&rc=ZDw2OWk3NGU8PDY6PGVpaEBpMzo7M3h2OzxmbjMzNDYzM0BeYGFgYy8zNTUxLjVgNTM0YSMvNTEzZS5xZGFfLS0vMS9zcw%3D%3D&btag=e00038008&dy_q=1711064381&l=2024032207394195BE77167061935E4859","request_time":1711066338,"response_time":"14","server_ip":"39.135.214.201","request_method":"GET","scheme":"http","server_protocol":"HTTP/1.1","status_code":"206","http_range":"bytes=133681437-134487857","bytes_sent":807925,"body_bytes_sent":"806421","hitmiss":"HIT","http_referer":"-","ua":"ttplayer(version:2.10.158.55-toutiao,appId:35,os:Android,traceId:69593771694T1711064382785T53607,appSessionId:Njk1OTM3NzE2OTQ1OTQ0ODA0MjQxNzExMDM2ODA2NzA4,tag:longvideo),AVDML_2.1.158.44-tt-net4_ANDROID,longvideo,MDLTaskPlay,MDLGroup(35)","server_port":"443","first_byte_time":"5","http_x_forward_for":"183.221.46.163,39.135.214.201","request_length":1042,"request_id":"587de9557b24b45cebde6daf253bb119","sent_http_content_length":"806421","request_body_length":0,"upstream_response_time":"0","http_cookie":"-","tcpinfo_rtt":"-","tcpinfo_rttvar":"-","layer":"0","bd_hitmiss":"edge_hit","TT-Request-TraceId":"-","Error_Reason":"-","x-tt-trace-id":"-","ServerTiming":"-","x-tt-trace-host":"-","x-tt-trace-tag":"-","via":"CHN-SCchengdu-CMCCZJ7-CACHE42[5],CHN-SCchengdu-CMCCZJ7-CACHE44[0,TCP_HIT,1],CHN-GDfoshan-GLOBALZJ1-CACHE154[102],CHN-GDfoshan-GLOBALZJ1-CACHE86[99,TCP_MISS,102],CHN-HAzhengzhou-GLOBALZJ1-CACHE152[28],CHN-HAzhengzhou-GLOBALZJ1-CACHE86[0,TCP_HIT,26],CHN-HAzhengzhou-GLOBAL3-CACHE111[46],CHN-HAzhengzhou-GLOBAL3-CACHE86[42,TCP_MISS,45],CHN-HElangfang-GLOBAL6-CACHE30[27],CHN-HElangfang-GLOBAL6-CACHE114[0,TCP_HIT,25]","sent_http_content_range":"bytes 133681437-134487857/216965399"}`
	var x sinker.CdnEdgeLog

	if err := json.Unmarshal([]byte(bytes), &x); err != nil {
		t.Fatal(err)
	}
}
