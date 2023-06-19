package main

import (
	"flag"
	"sync"

	"github.com/gin-gonic/gin"
)

var name2key map[string]string

var entries = []map[string]interface{}{
	{"name": "file.yalla.live", "key": "file-yalla-live-lb"},
	{"name": "q8h2kkzau.bkt.clouddn.com", "key": "q8h2kkzau-bkt-clouddn-com-lb"},
	{"name": "qa0eiorut.bkt.clouddn.com", "key": "qa0eiorut-bkt-clouddn-com-lb"},
	{"name": "www.iseetech.com.cn", "key": "www-iseetech-com-cn-lb"},
	{"name": "dl6.ztems.com", "key": "dl6-ztems-com-lb"},
	{"name": "wdl1.pcfg.cache.wpscdn.com", "key": "wdl1-pcfg-cache-wpscdn-com-lb"},
	{"name": "img-origin.ml.moonlian.com", "key": "img-1-origin-ml-moonlian-com-lb"},
	{"name": "hwwordqn.tuyoorock.com", "key": "hwwordqn-tuyoorock-com-lb"},
	{"name": "fga0.market.xiaomi.com", "key": "fga0-market-xiaomi-com-lb"},
	{"name": "autopatchhk2.yuanshen.com", "key": "autopatchhk2-yuanshen-com-lb"},
	{"name": "dc.xhscdn.com", "key": "dc-xhscdn-com-lb"},
	{"name": "file.yalla.games", "key": "file-yalla-games-lb"},
	{"name": "file.yallaludo.com", "key": "file-yallaludo-com-lb"},
	{"name": "ggglobal.qbox.net", "key": "ggglobal-qbox-net-lb"},
	{"name": "sg.xiaohongshu.com", "key": "sg-xiaohongshu-com-lb"},
	{"name": "sns-img-anim-hw.xhscdn.com", "key": "sns-img-anim-hw-xhscdn-com-lb"},
	{"name": "sns-img-anim-qn.xhscdn.com", "key": "sns-img-anim-qn-xhscdn-com-lb"},
	{"name": "sns-img-hw.xhscdn.com", "key": "sns-img-hw-xhscdn-com-lb"},
	{"name": "sns-img-qc.xhscdn.com", "key": "sns-img-qc-xhscdn-com-lb"},
	{"name": "sns-img-qn.xhscdn.com", "key": "sns-img-qn-xhscdn-com-lb"},
	{"name": "sns-video-hw.xhscdn.com", "key": "sns-video-hw-xhscdn-com-lb"},
	{"name": "sns-video-qc.xhscdn.com", "key": "sns-video-qc-xhscdn-com-lb"},
	{"name": "sns-video-qn.xhscdn.com", "key": "sns-video-qn-xhscdn-com-lb"},
	{"name": "v.xiaohongshu.com", "key": "v-xiaohongshu-com-lb"},
	{"name": "autopatchhk.yuanshen.com", "key": "autopatchhk-yuanshen-com-lb"},
	{"name": "genshinimpact.mihoyo.com", "key": "genshinimpact-mihoyo-com-lb"},
	{"name": "cdn.poger.xyz", "key": "cdn-poger-xyz"},
}
var initOnce sync.Once

func init() {
	initOnce.Do(func() {
		name2key = make(map[string]string)
		for _, pair := range entries {
			name2key[pair["name"].(string)] = pair["key"].(string)
		}
	})
}

var addr string

func parseCmdArgs() {
	flag.StringVar(&addr, "addr", ":18120", "specify address")
	flag.Parse()
}

func main() {
	parseCmdArgs()

	r := gin.Default()

	r.GET("/v1/reference/all", func(c *gin.Context) {
		c.JSON(200, entries)
	})

	r.GET("/v1/reference/domain/:domain", func(c *gin.Context) {
		domain := c.Param("domain")

		if v, ok := name2key[domain]; ok {
			c.JSON(200, gin.H{
				"name": domain,
				"key":  v,
			})
		} else {
			c.JSON(200, gin.H{
				"name": domain,
			})
		}

	})
	r.Run(addr) // listen and serve on 0.0.0.0:8080

}
