package routers

import (
	"GoP2PSpider/web/controllers"
	"github.com/astaxie/beego"
)

func init() {
	beego.Router("/", &controllers.SearchController{})
}
