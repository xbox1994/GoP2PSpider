package controllers

import (
	"GoP2PSpider/config"
	"GoP2PSpider/types"
	"GoP2PSpider/web/conf"
	"github.com/astaxie/beego"
	"log"
	"net/rpc"
	"strconv"
)

var (
	dataClient *rpc.Client
)

func init() {
	dataClient = conf.CreateDataClient()
}

type SearchController struct {
	beego.Controller
}

func (this *SearchController) Get() {
	q := this.GetString("q")
	if q == "" {
		this.TplName = "index.tpl"
	} else {
		s := this.Input().Get("start")
		var start int
		if s != "" {
			var e error
			start, e = strconv.Atoi(s)
			if e != nil {
				log.Printf("start param %s is invalid: %v", s, e)
			}
		}

		result := types.QueryResult{}
		dataClient.Call(config.DataServiceQuery, types.QueryParam{
			Q:     q,
			Start: start,
		}, &result)
		log.Printf("receive data:%v", result)
		this.TplName = "search.tpl"
	}
}
