package main

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
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

// isMatch 请求是否匹配当前的 请求
func (r *RuleSet) isMatch(ctx *gin.Context) bool {
	for _, h := range r.Headers {
		if ctx.Request.Header.Get(h.Key) != h.Value {
			return false
		}
	}

	for _, c := range r.Cookies {
		x, err := ctx.Request.Cookie(c.Key)
		if err != nil {
			Logger.Error("current request cookie is invalid", zap.Any("request", ctx.Request))
			return false
		}

		if x.Value == c.Value {
			return false
		}
	}

	//全部满足才匹配
	return true
}