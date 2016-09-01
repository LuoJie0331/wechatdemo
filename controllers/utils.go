package controllers

import (
	"github.com/gin-gonic/gin"
	"gitlab.appdao.com/luojie/wechat/logger"
)

const (
	SERVER_ERROR = iota
	BAD_REQUEST
	BAD_POST_DATA
	LOGIN_NEEDED
	LOGIN_FAILED
	PASSWORD_WRONG
	NOT_PERMITTED
	REGISTER_FAILED
)

type RequestLogData struct {
	Status bool
	Error  string
	Msg    string
}

var (
	errorStr = map[int][2]string{
		SERVER_ERROR:    [2]string{"sever_error", "服务器错误"},
		BAD_REQUEST:     [2]string{"bad_request", "客户端请求错误"},
		BAD_POST_DATA:   [2]string{"bad_post_data", "客户端请求体错误"},
		LOGIN_NEEDED:    [2]string{"login_needed", "未登录"},
		LOGIN_FAILED:    [2]string{"login_failed", "登录失败"},
		PASSWORD_WRONG:  [2]string{"password_wrong", "账号或密码错误"},
		NOT_PERMITTED:   [2]string{"not_permitted", "无权进行此次操作"},
		REGISTER_FAILED: [2]string{"register_failed", "注册失败, 登录名或邮箱重复"},
	}
)
var adminTokenEncryptKey = []byte("appwillgoogle@chaoyang")

//func DirectSuccess(c *gin.Context, data interface{}) {
//	res := data
//
//	logger.CommonLogger.Debug(map[string]interface{}{
//		"type": "api_request",
//		"url":  c.Request.RequestURI,
//		"resp": res,
//	})
//
//	c.Set("request_log", &RequestLogData{Status: true})
//	c.JSON(200, res)
//}

func Success(c *gin.Context, data interface{}) {
	res := gin.H{"status": true}
	if data != nil {
		res["data"] = data
	}

	logger.CommonLogger.Debug(map[string]interface{}{
		"type": "api_request",
		"url":  c.Request.RequestURI,
		"resp": res,
	})

	c.Set("request_log", &RequestLogData{Status: true})
	c.JSON(200, res)
}

func Error(c *gin.Context, errorCode int, data ...interface{}) {
	var (
		errCodeStr = errorStr[errorCode][0]
		errMsg     = errorStr[errorCode][1]
		errMsgLog  = errMsg
	)

	if len(data) >= 1 {
		if data[0] != nil {
			errMsg = data[0].(string)
		}
		if len(data) >= 2 {
			if data[1] != nil {
				errMsgLog = data[1].(string)
			} else {
				errMsgLog = errMsg
			}
		}
	}

	logger.CommonLogger.Error(map[string]interface{}{
		"type":    "api_request",
		"code":    errCodeStr,
		"url":     c.Request.RequestURI,
		"err_msg": errMsgLog,
		"err":     errMsg,
	})

	res := gin.H{"status": false, "code": errCodeStr, "msg": errMsg}
	c.Set("request_log", &RequestLogData{Status: false, Error: errCodeStr, Msg: errMsgLog})
	c.JSON(200, res)
}

func ClientError(c *gin.Context, data interface{}) {
	Error(c, BAD_POST_DATA, data, nil)
}

func ServerError(c *gin.Context, data interface{}) {
	Error(c, SERVER_ERROR, nil, data)
}
