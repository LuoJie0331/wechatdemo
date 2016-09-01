package main

import (
	"github.com/gin-gonic/gin"
	"gitlab.appdao.com/luojie/wechat/controllers"
)

func registerTestAPI(ginIns *gin.Engine) {
	testAPI := ginIns.Group("/test/")
	{
		testAPI.GET("/ping", controllers.Ping)
		testAPI.GET("/wechat", controllers.Wechat)
	}
}
