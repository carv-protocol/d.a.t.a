package web

import (
	"net/http"

	"github.com/carv-protocol/d.a.t.a/src/web/proto"

	"github.com/gin-gonic/gin"
)

func CommErr(errCode int64, errMsg string) *proto.Error {
	return &proto.Error{
		ErrCode: errCode,
		ErrMsg:  errMsg,
	}
}

func NilErr() *proto.Error {
	return &proto.Error{
		ErrCode: http.StatusOK,
		ErrMsg:  "OK",
	}
}

func SetOrigin(c *gin.Context) {
	origin := c.Request.Header.Get("Origin")
	c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
	c.Header("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE,UPDATE")
	c.Header("Access-Control-Allow-Headers", "Authorization, Content-Length, X-CSRF-Token, Token,session")
	c.Header("Access-Control-Expose-Headers", "Content-Length, Access-Control-Allow-Origin, Access-Control-Allow-Headers")
	c.Header("Access-Control-Allow-Credentials", "true")
}

func ParamsCheck(c *gin.Context, req interface{}) error {
	if c.Request.Method == "GET" {
		return c.ShouldBind(req)
	} else {
		return c.ShouldBindJSON(req)
	}
}
