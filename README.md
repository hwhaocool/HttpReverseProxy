ms-grey-proxy 房极客 灰度 转发系统
----

## 部署

部署文件 请看 `example/proxy-grey.yaml`  
需要有 `nginx` namespace

## 配置文件
配置文件是采用 `ConfigMap` 挂载的形式

需要新建一个 `grey-conf` 的 `ConfigMap`  
它由一个`key` 为 `config.yaml`  
`value` 就是配置文件的内容

请参考 默认配置文件 `example/config.yaml`

