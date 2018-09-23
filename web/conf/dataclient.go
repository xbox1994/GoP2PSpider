package conf

import (
	"GoP2PSpider/rpcsupport"
	"github.com/astaxie/beego"
	"net/rpc"
)

func CreateDataClient() *rpc.Client {
	dataServerHost := beego.AppConfig.String("dataserverhost")
	client, e := rpcsupport.NewClient(dataServerHost)
	if e != nil {
		panic(e)
	}
	return client
}
