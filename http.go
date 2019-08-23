package main

import (
    "net"
    "net/http"
    "net/http/httputil"
    "net/url"
    "time"
    "strings"
    "fmt"
    "go.uber.org/zap"
)

// MyURL MyURL
type MyURL struct {
    url.URL
}

// RequestURI RequestURI
func (u *MyURL) RequestURI() string {
    // return u.URL.RequestURI()

    result := u.Opaque
    if result == "" {
        result = u.EscapedPath()
        if result == "" {
            result = "/"
        }
    } else {
        if strings.HasPrefix(result, "//") {
            result = u.Scheme + ":" + result
        }
    }
    if u.ForceQuery || u.RawQuery != "" {
        result += "?" + u.RawQuery
    }
    return result
}

// getTransport 得到自定义的 Transport
func getTransport() http.RoundTripper {

    return &http.Transport{
        Proxy: http.ProxyFromEnvironment,
        DialContext: (&net.Dialer{
            Timeout:   30 * time.Second,
            KeepAlive: 30 * time.Second,
            DualStack: true,
        }).DialContext,

        IdleConnTimeout:       90 * time.Second,
        TLSHandshakeTimeout:   10 * time.Second,
        ExpectContinueTimeout: 1 * time.Second,

        //上面的都是系统默认配置

        //连接池 最大连接数
        MaxIdleConns: 400,

        //每个host的默认连接数，默认为2
        MaxIdleConnsPerHost: 50,
    }
}

// MyReverseProxy 自定义反向代理
func MyReverseProxy(target *url.URL) *httputil.ReverseProxy {
    proxy := httputil.NewSingleHostReverseProxy(target)
    proxy.ErrorHandler = myErrorHandler
    proxy.Transport = getTransport()

    return proxy
}

// myErrorHandler 代理服务器的错误处理，只是打印日志
func myErrorHandler(rw http.ResponseWriter, req *http.Request, err error) {
    Logger.Error("http proxy error", zap.Error(err), zap.String("request 2", getRequestString(req)))
    Logger.Error("http proxy error", zap.Error(err), zap.String("host", req.Host), zap.String("url", req.RequestURI))
    rw.WriteHeader(http.StatusBadGateway)
}

// GetRequestString 得到 request 的字符串
func getRequestString(req *http.Request) string {
    return fmt.Sprintf("[req] method = %s, proto = %s, host = %s, RemoteAddr = %s, RequestURI = %s, url = %s", 
        req.Method, req.Proto, req.Host, req.RemoteAddr, req.RequestURI, req.URL.String())
}