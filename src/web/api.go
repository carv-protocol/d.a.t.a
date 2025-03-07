package web

import (
	"net/http"

	"github.com/carv-protocol/d.a.t.a/src/web/proto"

	"github.com/gin-gonic/gin"
)

func Healthy(c *gin.Context) {
	SetOrigin(c)

	c.JSON(http.StatusOK, proto.HealthyRsp{})
}

func AreYouReady(c *gin.Context) {
	SetOrigin(c)

	c.JSON(http.StatusOK, proto.AreYouReadyRsp{
		Status: "success",
	})
}

func Talk(c *gin.Context) {
	SetOrigin(c)

	var req proto.TalkReq
	if err := ParamsCheck(c, &req); err != nil {
		c.JSON(http.StatusOK, *CommErr(http.StatusBadRequest, err.Error()))
		return
	}

	c.JSON(http.StatusOK, proto.TalkRsp{
		Error:   *NilErr(),
		Content: "",
	})
}
