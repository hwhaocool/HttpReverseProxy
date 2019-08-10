package main

import (
	yaml "gopkg.in/yaml.v2"
	"io/ioutil"

	"regexp"
	"fmt"

	"go.uber.org/zap"
	"github.com/gin-gonic/gin"
)

//配置文件地址
var configFilePath = "/app/config/config.yaml"

//serviceMap key是缩写，value 是 host
var serviceMap = make(map[string]string)

//[]RuleSet
var  ruleList []RuleSet

// config 配置文件
var config GreyConfig

// InitConfigFile 初始化配置
func InitConfigFile() {
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

	checkAndAnalysisRule()
}

//checkAndAnalysisRule 校验规则 解析规则
func checkAndAnalysisRule() {
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

	b, err := yaml.Marshal(ruleList)
	if err != nil {
		fmt.Println(string(b))

		Logger.Info(string(b))
	}
	
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

	//是否有不合法的规则
	ok := false

	//多个 result 之间是 并且 的关系
	for _, result := range reg.FindAllSubmatch(ruleByte, -1) {

		ok = true

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

	if ! ok {
		Logger.Fatal("current rule is inavlid", zap.Int("index", index), zap.String("rule", rule.Rule))
	}

	ruleList[index] = *currentRule
	Logger.Info("current rule set", zap.Any("rule", currentRule))
	
}

//GetDestination 得到当前 请求将要发往的目的地
func GetDestination(ctx *gin.Context) string {
	for index, rule := range ruleList {
		if rule.isMatch(ctx) {

			Logger.Info("match", zap.Int("index", index))
			return rule.ServiceHost
		}
	}

	Logger.Info("match nothing, use default")
	return serviceMap[config.DefaultService]
}