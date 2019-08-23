package main

import (
    "net"
    "net/http"
    "net/http/httputil"
    "net/url"
    "time"
    "strings"
    // "crypto/tls"
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

// GetTransport 得到自定义的 Transport
func GetTransport() http.RoundTripper {

    return &http.Transport{
        Proxy: MyProxy,
        // Proxy: ProxyFromEnvironment,
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

// func (t *MyTransport) RoundTrip(req *http.Request) (*http.Response, error) {
//     return t.RoundTrip(req)

//     // t.
// }

// func  MyDirector(req *http.Request) {
// 	req.URL.Scheme = target.Scheme
// 	req.URL.Host = target.Host
// 	req.URL.Path = singleJoiningSlash(target.Path, req.URL.Path)
// 	if targetQuery == "" || req.URL.RawQuery == "" {
// 		req.URL.RawQuery = targetQuery + req.URL.RawQuery
// 	} else {
// 		req.URL.RawQuery = targetQuery + "&" + req.URL.RawQuery
// 	}
// 	if _, ok := req.Header["User-Agent"]; !ok {
// 		// explicitly disable User-Agent so it's not set to default value
// 		req.Header.Set("User-Agent", "")
// 	}
// }

// MyReverseProxy 我的反向代理，主要是定制了 director的 scheme，其它代码都是照抄
func MyReverseProxy(target *url.URL) *httputil.ReverseProxy {
    targetQuery := target.RawQuery
    director := func(req *http.Request) {
        //req.URL.Scheme = target.Scheme
        req.URL.Host = target.Host
        req.URL.Path = singleJoiningSlash(target.Path, req.URL.Path)
        if targetQuery == "" || req.URL.RawQuery == "" {
            req.URL.RawQuery = targetQuery + req.URL.RawQuery
        } else {
            req.URL.RawQuery = targetQuery + "&" + req.URL.RawQuery
        }
        if _, ok := req.Header["User-Agent"]; !ok {
            // explicitly disable User-Agent so it's not set to default value
            req.Header.Set("User-Agent", "")
        }
    }
    return &httputil.ReverseProxy{Director: director}
}

// singleJoiningSlash 照抄 revserseproxy 的代码
func singleJoiningSlash(a, b string) string {
    aslash := strings.HasSuffix(a, "/")
    bslash := strings.HasPrefix(b, "/")
    switch {
    case aslash && bslash:
        return a + b[1:]
    case !aslash && !bslash:
        return a + "/" + b
    }
    return a + b
}

// MyProxy MyProxy
func MyProxy(req *http.Request) (*url.URL, error) {
    target := GetDestination(req, 2)

    if strings.HasPrefix(target, "http") == false {
        target = "http://" + target
    }

    u, _ := url.Parse(target)
    return u, nil
}