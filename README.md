# GenerateProjectStruct

生成较为通用的项目结构代码，适用于web项目和grpc服务等。

参考的[gin-vue-admin](https://github.com/flipped-aurora/gin-vue-admin)项目。

项目分层架构设计，业务逻辑减低耦合度。

1. 克隆项目
```shell
git clone https://github.com/Huashusu/GenerateProjectStruct.git
```
2. 构建运行
```shell
go build .
```
3. 生成项目
```shell
./GenerateProjectStruct
```
4. 进入项目，检查依赖
```shell
cd <project name>
go mod tidy
```

生成的目录结构
```shell
├── api
├── cmd
├── go.mod
└── main.go
├── config
│   ├── config.go
│   ├── log.go
│   └── system.go
├── config.yaml
├── core
│   ├── log_file_write.go
│   ├── server.go
│   ├── viper.go
│   └── zap.go
├── global
│   └── global.go
├── go.mod
├── initialiaze
│   └── router.go
├── log
├── main.go
├── middleware
│   ├── logger.go
│   └── recovery.go
├── model
├── router
├── service
└── utils
    └── local_ip.go
```
生成样例项目结构演示地址
[gin-web](http://github.com/Huashusu/gin-web)

### 二开必看
1. router中开发例子：
```go
type _RouteGroup struct {
	Other otherRouter
}

var RouteGroup = new(_RouteGroup)
```
需要更好的隔离性和减低耦合可以采用以下方案：
```go
type routerInterface interface {
    InjectRouter(*gin.RouterGroup)
}

type _RouteGroup struct {
    Other routerInterface
}

var _ routerInterface = &otherRouter{}

var RouteGroup = _RouteGroup{
    Other: new(otherRouter),
}
```
2. model和api目录结构应一一对应，使用总结构包装一起，开放接口实现和面向接口开发，参考router开发例子
3. service推荐面向接口开发，类似于SpringBoot的Service，调用DB等，在global中添加全局DB连接，读写分离等在core中实现
> 这个仅靠描述，价值有限，有空更新[gin-web](http://github.com/Huashusu/gin-web)的样例代码