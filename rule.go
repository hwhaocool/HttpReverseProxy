package main

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// GreyConfig 配置
type GreyConfig struct {
	Services       []Service
	DefaultService string `yaml:"defaultService"`
	Rules          []Rule ``
}

// Service 服务信息
type Service struct {
	Name        string `yaml:"name"`
	ServiceHost string `yaml:"serviceHost"`
}

// Rule 文件里的规则信息
type Rule struct {
	Rule        string `yaml:"rule"`
	ServiceName string `yaml:"serviceName"`
}

// RuleSet 解析出来的规则
type RuleSet struct {
	Headers []HeaderRule
	Cookies []CookieRule
	ServiceHost string
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