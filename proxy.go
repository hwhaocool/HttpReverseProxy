package main

import (
    "strings"
    "net"
    "net/http"
    "net/http/httputil"
    "net/url"

    "github.com/gin-gonic/gin"
    "go.uber.org/zap"
    "time"
    "math/rand"
)

func main() {
    //初始化日志
    InitLogger()
    Logger.Info("proxy b begin to start")

    // 读取配置文件
    InitConfigFile()

    //打印ip
    Logger.Info("local ip", zap.String("ip", getLocalIP()))

    //新建gin 实例
    router := gin.New()

    //返回欢迎信息
    router.GET("/grey", welcome)
    router.HEAD("/grey", welcomeSlb)

    // slb 健康检查接口使用 head 方法
    router.HEAD("/", welcome)

    //其它 -> 根据经纪人类型来
    router.NoRoute(reverseProxy)

    //启动 gin 并监听端口
    err := router.Run(":8080")
    if err != nil {
        Logger.Fatal("proxy start failed,", zap.Error(err))
    }
}

// reverseProxy 反向代理逻辑
func reverseProxy(ctx *gin.Context) {
    start := time.Now().UnixNano() / 1e5

    rand.Seed(time.Now().Unix())
    randomID := rand.Intn(1000)
        
    target := GetDestination(ctx, randomID)

    if strings.HasPrefix(target, "http") == false {
        target = "http://" + target
    }

    url, _ := url.Parse(target)
    
    Logger.Info("scheme", zap.String("request", ctx.Request.URL.Scheme), 
        zap.String("proxy", url.Scheme),
        zap.Int("randomId", randomID),
        )

    proxy := httputil.NewSingleHostReverseProxy(url)
    proxy.ErrorHandler = myErrorHandler

    end := time.Now().UnixNano() / 1e5

    //记录处理rule的耗时
    Logger.Info("reverseProxy", zap.String("method", ctx.Request.Method), 
        zap.String("url", ctx.Request.RequestURI), 
        zap.String("target", target),
        zap.Int64("cost(1/10 ms)", end - start),
        zap.Int("randomId", randomID))

    start = time.Now().UnixNano() / 1e5
    proxy.ServeHTTP(ctx.Writer, ctx.Request)
    end = time.Now().UnixNano() / 1e5

    Logger.Info("reverseProxy",
        zap.Int64("request cost(1/10 ms)", end - start),
        zap.Int("randomId", randomID))

}

// welcome 健康检查接口
func welcome(ctx *gin.Context) {

    ctx.JSON(200, gin.H{
        "type":    "ok",
        "message": "grey proxy is ok",
        "ip":      getLocalIP(),
    })
}

// welcomeSlb slb 健康检查接口
func welcomeSlb(ctx *gin.Context) {
    Logger.Info("welcomeSlb",
        zap.String("remote addr", ctx.Request.RemoteAddr))

    ctx.JSON(200, gin.H{
        "type":    "ok",
        "message": "grey proxy is ok",
        "ip":      getLocalIP(),
    })
}

// myErrorHandler 代理服务器的错误处理，只是打印日志
func myErrorHandler(rw http.ResponseWriter, req *http.Request, err error) {
    Logger.Error("http proxy error", zap.Error(err), zap.Any("request 2", *req))
    Logger.Error("http proxy error", zap.Error(err), zap.String("host", req.Host), zap.String("url", req.RequestURI))
    rw.WriteHeader(http.StatusBadGateway)
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
