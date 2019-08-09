package main

import (
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	// "github.com/jinzhu/configor"
	// "./config"
	// "./httpMgr"
	// "./jwtAuth"
	// "./myLog"
	// "./service"
)

var simpleHostProxy = httputil.ReverseProxy{
	Director: func(req *http.Request) {
		req.URL.Scheme = "http"
		// req.URL.Host = HOST
		// req.Host = HOST
	},
}

var logger *zap.Logger

func main() {

	//初始化日志
	logger := InitLogger()
	logger.Info("proxy b begin to start")

	// 读取配置文件
	InitConfigFile()

	//打印ip
	logger.Info("local ip", zap.String("ip", getLocalIP()))

	//新建gin 实例
	router := gin.New()

	//返回欢迎信息
	router.GET("/", welcome)
	router.HEAD("/", nginxHealthCheck)

	//其它 -> 根据经纪人类型来
	router.NoRoute(fanggeekNoRoute)

	//启动 gin 并监听端口
	err := router.Run(":8080")
	if err != nil {
		logger.Fatal("proxy start failed,", zap.Error(err))
	}
}

// fanggeekNoRoute 自定义的 NoRoute 函数
// 有些类型的接口，企业号已全部梳理出来了，剩下的没有必要走 分发，可以直接走个人号
func fanggeekNoRoute(c *gin.Context) {
	path := c.Request.URL.Path
	method := c.Request.Method

	c.Request.Header.Get()

	logger.Info("fanggeekNoRoute", zap.String("method", method), zap.String("path", path))

	distributeReq(c)

	// if strings.HasPrefix(path, "/v1/customer") {
	//     //客户相关的，企业号签名已梳理过，剩下没有写的，都走到个人号
	//     logger.Info("fanggeekNoRoute direct 2 agent", zap.String("method", method), zap.String("path", path))
	//     agentProxy(c)
	// } else {
	//     distributeReq(c)
	// }
}

//分发
func distributeReq(ctx *gin.Context) {
	logger.Info("distributeReq", zap.String("method", ctx.Request.Method), zap.String("url", ctx.Request.RequestURI))

	//取出token
	authHeader := ctx.Request.Header.Get("Authorization")

	if 0 == len(authHeader) {
		logger.Warn("auth is empty")
		//没有的话，直接到个人号
		// agentProxy(ctx)

		//没有的话，直接到企业号
		teamProxy(ctx)
		return
	}

	jwtToken := strings.Replace(authHeader, "Bearer ", "", 1)

	//校验token
	_, err := jwtAuth.VerifyJWT(jwtToken)
	if err != nil {
		logger.Error("check token failed", zap.Error(err), zap.String("token", authHeader))

		//校验不过，输出错误信息
		errorRespone(ctx)
		return
	}

	//调接口，得到 agent 信息
	//重试两次
	agentInfo, err := service.GetAgentInfo(jwtToken, 2)
	if err != nil {
		logger.Error("get agent info failed", zap.Error(err))

		errorRespone(ctx)
		return
	}

	ReverseProxy(ctx, agentInfo.Type)
}

// welcome 健康检查接口
func welcome(ctx *gin.Context) {
	logger.Info("now is welcome", zap.String("addr", ctx.Request.RemoteAddr))

	ctx.JSON(200, gin.H{
		"type":    "ok",
		"message": "proxy is ok",
		"ip":      getLocalIP(),
	})
}

// slb 健康检查接口使用 head 方法
func nginxHealthCheck(ctx *gin.Context) {
	ctx.JSON(200, gin.H{
		"type":    "ok",
		"message": "proxy is ok",
	})
}

// errorRespone a gin.HandlerFunc
// 错误提示
func errorRespone(ctx *gin.Context) {
	ctx.JSON(200, gin.H{
		"type": "error",
		"message": gin.H{
			"code":   201,
			"errmsg": "帐号验证失败，请重新登录",
		},
	})
}

// ReverseProxy 反向代理逻辑
func ReverseProxy(c *gin.Context, atype int) {
	logger.Info("select service", zap.Int("type", atype))

	if 1 == atype {
		//企业号
		teamProxy(c)
	} else {
		agentProxy(c)
	}
}

// agentProxy 个人号接口的反向代理
func agentProxy(c *gin.Context) {
	logger.Info("agent request", zap.String("method", c.Request.Method), zap.String("url", c.Request.RequestURI))

	target := config.Config.Servicehost.AgentAPI
	url, _ := url.Parse(target)

	proxy := httputil.NewSingleHostReverseProxy(url)
	proxy.ErrorHandler = myErrorHandler

	proxy.ServeHTTP(c.Writer, c.Request)
}

// teamProxy 企业号接口的反向代理
func teamProxy(c *gin.Context) {
	logger.Info("team request", zap.String("method", c.Request.Method), zap.String("url", c.Request.RequestURI))

	target := config.Config.Servicehost.TeamAPI
	url, _ := url.Parse(target)

	proxy := httputil.NewSingleHostReverseProxy(url)
	proxy.ErrorHandler = myErrorHandler

	proxy.ServeHTTP(c.Writer, c.Request)
}

// myErrorHandler 代理服务器的错误处理，只是打印日志
func myErrorHandler(rw http.ResponseWriter, req *http.Request, err error) {
	logger.Error("http proxy error", zap.Error(err), zap.Any("request 2", *req))
	logger.Error("http proxy error", zap.Error(err), zap.String("host", req.Host), zap.String("url", req.RequestURI))
	rw.WriteHeader(http.StatusBadGateway)
}

func withHeader(ctx *gin.Context) {
	ctx.Request.Header.Add("request-uid", "id")
	simpleHostProxy.ServeHTTP(ctx.Writer, ctx.Request)
}

// getLocalIP 得到local ip
func getLocalIP() string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return ""
	}
	for _, address := range addrs {
		// check the address type and if it is not a loopback the display it
		if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP.String()
			}
		}
	}
	return ""
}
