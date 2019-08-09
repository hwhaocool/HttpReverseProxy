package main

import (
	yaml "gopkg.in/yaml.v2"
	"io/ioutil"

	"regexp"

	"go.uber.org/zap"
)

//配置文件地址
// var configFilePath = "/app/config/config.yaml"
var configFilePath = "./config/config.yaml"

//serviceMap key是缩写，value 是 host
var serviceMap = make(map[string]string)

//[]RuleSet
var  ruleList []RuleSet

// Config 配置文件
var Config GreyConfig

// GreyConfig 配置
type GreyConfig struct {
	Services       []Service
	DefaultService string `yaml:"defaultService"`
	Rules          []Rule ``
}

type Service struct {
	Name        string `yaml:"name"`
	ServiceHost string `yaml:"serviceHost"`
}

type Rule struct {
	Rule        string `yaml:"rule"`
	ServiceName string `yaml:"serviceName"`
}

type RuleSet struct {
	Headers []HeaderRule
	Cookies []CookieRule
	ServiceName string
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

	// 解析yaml 内容到 Config
	err = yaml.Unmarshal(yamlFile, &Config)

	if err != nil {
		Logger.Fatal("yamlFile Unmarshal error", zap.Error(err))
	}

	Logger.Info("", zap.Any("config", Config))

	checkRule()
}

//checkRule 校验规则
func checkRule() {
	if Config.DefaultService == "" {
		Logger.Fatal("yamlFile DefaultService is required, you should set it")
	}

	for index, service := range Config.Services {
		defer serviceError(index, service)

		serviceMap[service.Name] = service.ServiceHost
	}

	_, ok := serviceMap[Config.DefaultService]

	if ok == false {
		//不存在
		Logger.Fatal("yamlFile services occur error, you should set default servie's host", zap.String("default service name", Config.DefaultService))
	}

	ruleList = make([]RuleSet, len(Config.Rules))

	for index, rule := range Config.Rules {
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

	currentRule := ruleList[index]

	currentRule.ServiceName = rule.ServiceName
	currentRule.Headers = make([]HeaderRule, 10)
	currentRule.Cookies = make([]CookieRule, 10)

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

			// k := HeaderRule{
			// 	Key: ruleKey,
			// 	Value : ruleValue,
			// }

			currentRule.Headers = append(currentRule.Headers, *h)
			// append(currentRule.Headers, k)
		case "cookie":
			h := new(CookieRule)
			h.Key = ruleKey
			h.Value = ruleValue

			currentRule.Cookies = append(currentRule.Cookies, *h)
		}
	}

	Logger.Info("current rule set", zap.Any("rule", currentRule))
	
}