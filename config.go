package main

import (
	yaml "gopkg.in/yaml.v2"
	"io/ioutil"

	"go.uber.org/zap"
)

//配置文件地址
var configFilePath = "/app/config/config.yaml"

var ruleMap = make(map[string]string)

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

		ruleMap[service.Name] = service.ServiceHost
	}

	_, ok := ruleMap[Config.DefaultService]

	if ok == false {
		//不存在
		logger.Fatal("yamlFile services occur error, you should set default servie's host", zap.String("default service name", Config.DefaultService))
	}

	for index, rule := range Config.Rules {
		if rule.ServiceName == "" {
			logger.Fatal("yamlFile rules occur error, you should set service for rule", zap.Int("index", index))
		}
	}
}

func ruleError(index int, service Service) {
	logger.Fatal("yamlFile services occur error", zap.Int("index", index), zap.String("name", service.Name))
}
