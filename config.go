package main

import (
	yaml "gopkg.in/yaml.v2"
	"io/ioutil"

	"regexp"

	"go.uber.org/zap"
)

//配置文件地址
var configFilePath = "/app/config/config.yaml"

//serviceMap key是缩写，value 是 host
var serviceMap = make(map[string]string)

// var ruleList = []RuleSet

// Config 配置文件
var Config GreyConfig

// GreyConfig 配置
type GreyConfig struct {
	Services       []Service
	DefaultService string ``
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

type HeaderRule struct {
	Key   string
	Value string
}

type CookieRule struct {
	Key   string
	Value string
}

// func (s *Rule) 

// InitConfigFile 初始化配置
func InitConfigFile() {

	// var logger = myLog.Logger

	Logger.Info("xxx")

	//读文件
	yamlFile, err := ioutil.ReadFile(configFilePath)
	if err != nil {

		//出错，直接退出
		logger.Fatal("yamlFile read error", zap.Error(err))
	}

	// 解析yaml 内容到 Config
	err = yaml.Unmarshal(yamlFile, &Config)

	if err != nil {
		logger.Fatal("yamlFile Unmarshal error", zap.Error(err))
	}

	logger.Info("", zap.Any("config", Config))
}

//checkRule 校验规则
func checkRule() {
	if Config.DefaultService == "" {
		logger.Fatal("yamlFile DefaultService is required, you should set it")
	}

	for index, service := range Config.Services {
		defer ruleError(index, service)

		serviceMap[service.Name] = service.ServiceHost
	}

	_, ok := serviceMap[Config.DefaultService]

	if ok == false {
		//不存在
		logger.Fatal("yamlFile services occur error, you should set default servie's host", zap.String("default service name", Config.DefaultService))
	}

	for index, rule := range Config.Rules {
		if rule.ServiceName == "" {
			logger.Fatal("yamlFile rules occur error, you should set service for rule", zap.Int("index", index))
		}

		analysisRule(rule)
	}

	
}

func ruleError(index int, service Service) {
	logger.Fatal("yamlFile services occur error", zap.Int("index", index), zap.String("name", service.Name))
}

func analysisRule(rule Rule) {
	ruleByte := []byte(rule.Rule)

	reg := regexp.MustCompile(`(header|cookie)\(\s*\"([^"]+)\"\s*,\s*\"([^"]+)\"\s*\)`)

	result := reg.FindSubmatch(ruleByte)
	
	logger.Info("", zap.ByteStrings("result", result))
}