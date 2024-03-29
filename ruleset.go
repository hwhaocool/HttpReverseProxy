package main

import (
    "go.uber.org/zap"
    "net/http"
)

// RuleSet 解析出来的规则
type RuleSet struct {
    Headers []HeaderRule
    Cookies []CookieRule
    ServiceHost string
    Weight      int 
    RuleName    string 
}

// HeaderRule HeaderRule
type HeaderRule struct {
    Key   string
    Value string
}

//CookieRule cookie
type CookieRule struct {
    Key   string
    Value string
}

// RuleSets RuleSet的切片，主要是为了实现排序接口
type RuleSets []RuleSet

//Len()
func (s RuleSets) Len() int {
    return len(s)
}

//Less():权重将由高到低排序
func (s RuleSets) Less(i, j int) bool {
    return s[i].Weight > s[j].Weight
}

//Swap()
func (s RuleSets) Swap(i, j int) {
    s[i], s[j] = s[j], s[i]
}

// isMatch 请求是否匹配当前的 请求
func (r *RuleSet) isMatch(req *http.Request) bool {
    for _, h := range r.Headers {
        if req.Header.Get(h.Key) != h.Value {
            return false
        }
    }

    for _, c := range r.Cookies {
        x, err := req.Cookie(c.Key)
        if err != nil {
            Logger.Debug("current request cookie is invalid", zap.String("request", getRequestString(req)), zap.Any("cookie", req.Cookies()))
            return false
        }

        if x.Value == c.Value {
            return false
        }
    }

    //全部满足才匹配
    return true
}