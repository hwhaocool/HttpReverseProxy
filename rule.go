package main

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
    Rule        string `yaml:"rule"`                 //规则表达式集合， header cookie host
    ServiceName string `yaml:"serviceName"`          //service name， 必填
    Weight      int    `yaml:"weight"`               //权重，非必填， 默认50， 范围是 1-100
    Name        string `yaml:"name"`                 //规则名称，必填
}

// getWeight 得到权重
func (r *Rule) getWeight() int {
    if r.Weight == 0 {
        return 50
    } else if r.Weight > 100 {
        return 100
    }

    return r.Weight
}

