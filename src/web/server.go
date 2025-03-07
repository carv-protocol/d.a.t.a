package web

import (
	"context"
	"net"
	"net/http"
	"net/http/httputil"
	"os"
	"runtime/debug"
	"strconv"
	"strings"
	"time"

	"github.com/carv-protocol/d.a.t.a/src/pkg/logger"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

var (
	server *http.Server
)

func Start(port int) {
	server = newServer(port)
	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.GetLogger().Fatalf("listen err: %v", err)
		}
	}()
}

func Stop() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		logger.GetLogger().Errorf("[web] api svr shutdown err: %v", err)
	}
}

func newServer(port int) *http.Server {

	gin.SetMode(gin.ReleaseMode)

	r := gin.New()
	r.Use(GinRecovery(true), ZapLogger(logger.GetLogger()))

	r.Any("/talk", Talk)
	r.GET("/healthy", Healthy)
	r.GET("/are/you/ready", AreYouReady)

	return &http.Server{
		Addr:    ":" + strconv.Itoa(port),
		Handler: r,
	}
}

func GinRecovery(stack bool) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				var brokenPipe bool
				if ne, ok := err.(*net.OpError); ok {
					if se, ok := ne.Err.(*os.SyscallError); ok {
						if strings.Contains(strings.ToLower(se.Error()), "broken pipe") || strings.Contains(strings.ToLower(se.Error()), "connection reset by peer") {
							brokenPipe = true
						}
					}
				}

				httpRequest, _ := httputil.DumpRequest(c.Request, false)
				if brokenPipe {
					logger.GetLogger().Errorf("%v %v %v %v %v (%v)", c.Writer.Status(), c.Request.Method, c.Request.URL.RequestURI(), "", c.Request.UserAgent(), c.ClientIP())
					c.Error(err.(error))
					c.Abort()
					return
				}

				if stack {
					logger.GetLogger().Errorf("[Recovery from panic]\n%v%v\n%v", string(httpRequest), err, string(debug.Stack()))
				} else {
					logger.GetLogger().Errorf("[Recovery from panic]\n%v%v", string(httpRequest), err)
				}
				c.AbortWithStatus(http.StatusInternalServerError)
			}
		}()
		c.Next()
	}
}

func ZapLogger(logger *zap.SugaredLogger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()
		logger.Infof(
			"Request handled, method: %s | path: %s | status: %d | duration: %s",
			c.Request.Method, c.Request.URL.Path, c.Writer.Status(), time.Since(start),
		)
	}
}
