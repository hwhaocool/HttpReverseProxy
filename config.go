package main

import (
	yaml "gopkg.in/yaml.v2"
	"io/ioutil"

	"regexp"

	"go.uber.org/zap"
	"github.com/gin-gonic/gin"
)

//配置文件地址
// var configFilePath = "/app/config/config.yaml"
var configFilePath = "./config/config.yaml"

//serviceMap key是缩写，value 是 host
var serviceMap = make(map[string]string)

//[]RuleSet
var  ruleList []RuleSet

// config 配置文件
var config GreyConfig

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

func (r *RuleSet) isMatch(ctx *gin.Context) bool {
	return true
}

// func (s *Rule) 

// InitConfigFile 初始化配置
func InitConfigFile() {

	Logger.Info("xxx")

	//读文件
	yamlFile, err := ioutil.ReadFile(configFilePath)
	if err != nil {

		//出错，直接退出
		Logger.Fatal("yamlFile read error", zap.Error(err))
	}

	// 解析yaml 内容到 config
	err = yaml.Unmarshal(yamlFile, &config)

	if err != nil {
		Logger.Fatal("yamlFile Unmarshal error", zap.Error(err))
	}

	Logger.Info("", zap.Any("config", config))

	checkRule()
}

//checkRule 校验规则
func checkRule() {
	if config.DefaultService == "" {
		Logger.Fatal("yamlFile DefaultService is required, you should set it")
	}

	for _, service := range config.Services {
		// defer serviceError(index, service)

		serviceMap[service.Name] = service.ServiceHost
	}

	_, ok := serviceMap[config.DefaultService]

	if ok == false {
		//不存在
		Logger.Fatal("yamlFile services occur error, you should set default servie's host", zap.String("default service name", config.DefaultService))
	}

	ruleList = make([]RuleSet, len(config.Rules))

	for index, rule := range config.Rules {
		if rule.ServiceName == "" {
			Logger.Fatal("yamlFile rules occur error, you should set service for rule", zap.Int("index", index))
		}

		analysisRule(index, rule)
	}

	Logger.Info("rule is ", zap.Any("ruleList", ruleList))
	
}

func serviceError(index int, service Service) {
	Logger.Error("yamlFile services occur error", zap.Int("index", index), zap.String("name", service.Name))
}

func analysisRule(index int, rule Rule) {
	ruleByte := []byte(rule.Rule)

	reg := regexp.MustCompile(`(header|cookie)\(\s*\"([^"]+)\"\s*,\s*\"([^"]+)\"\s*\)`)

	currentRule := new(RuleSet)

	//换成服务地址
	currentRule.ServiceHost =  serviceMap[rule.ServiceName]

	currentRule.Headers = make([]HeaderRule, 0)
	currentRule.Cookies = make([]CookieRule, 0)

	//多个 result 之间是 并且 的关系
	for _, result := range reg.FindAllSubmatch(ruleByte, -1) {

		Logger.Info("", zap.ByteStrings("result", result))

		ruleType := string(result[1])
		ruleKey := string(result[2])
		ruleValue := string(result[3])

		switch ruleType {
		case "header":
			h := new(HeaderRule)
			h.Key = ruleKey
			h.Value = ruleValue

			currentRule.Headers = append(currentRule.Headers, *h)
		case "cookie":
			h := new(CookieRule)
			h.Key = ruleKey
			h.Value = ruleValue

			currentRule.Cookies = append(currentRule.Cookies, *h)
		}
	}

	ruleList[index] = *currentRule
	Logger.Info("current rule set", zap.Any("rule", currentRule))
	
}

//GetDestination 得到当前 请求将要发往的目的地
func GetDestination(ctx *gin.Context) string {
	for _, rule := range ruleList {
		if rule.isMatch(ctx) {
			return rule.ServiceHost
		}
	}

	return serviceMap[config.DefaultService]
}