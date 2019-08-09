
FROM registry.cn-shenzhen.aliyuncs.com/cuishiwen/geek-go:v1 AS builder
RUN mkdir /build 

#把当前目录的文件全部复制到 build里面去
ADD . /build/
WORKDIR /build 

#在docker里面编译
RUN go version
RUN go build -o main .

# 编译完成之后，把生成的文件复制到 alpine 里运行
FROM alpine:3.9
COPY --from=builder /build/main /app/
COPY --from=builder /build/config /app/config/

WORKDIR /app

#指定端口为 8080
EXPOSE 8080

ENTRYPOINT ["./main"]