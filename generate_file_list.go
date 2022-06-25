package main

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"path"
	"strings"
)

type fileItem struct {
	Path     string
	Filename string
	Text     string
}

var (
	fileList = make([]fileItem, 0, 10)
)

func init() {
	fileList = append(fileList,
		fileItem{
			Path:     "global",
			Filename: "global.go",
			Text: `package global

import (
	"{{projectName}}/config"

	"go.uber.org/zap"

	"github.com/spf13/viper"
)

var (
	VIPER  *viper.Viper
	CONFIG *config.Server
	LOG    *zap.Logger
)
`},
		fileItem{
			Path:     "core",
			Filename: "zap.go",
			Text: `package core

import (
	"{{projectName}}/global"
	"fmt"
	"strings"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func Zap() (logger *zap.Logger) {
	cores := make([]zapcore.Core, 0, 7)
	d := getCore(zap.DebugLevel)
	i := getCore(zap.InfoLevel)
	w := getCore(zap.WarnLevel)
	e := getCore(zap.ErrorLevel)
	p := getCore(zap.PanicLevel)
	f := getCore(zap.FatalLevel)
	switch strings.ToLower(global.CONFIG.Log.Level) {
	case "debug":
		cores = append(cores, d, i, w, e, p, f)
	case "info":
		cores = append(cores, i, w, e, p, f)
	case "warn":
		cores = append(cores, w, e, p, f)
	case "error":
		cores = append(cores, e, p, f)
	case "panic":
		cores = append(cores, p, f)
	case "fatal":
		cores = append(cores, f)
	default:
		cores = append(cores, d, i, w, e, p, f)
	}
	logger = zap.New(zapcore.NewTee(cores...), zap.AddCaller())
	if global.CONFIG.Log.ShowLine {
		logger = logger.WithOptions(zap.AddCaller())
	}
	return logger
}

func getEncodeConfig() (config zapcore.EncoderConfig) {
	config = zapcore.EncoderConfig{
		MessageKey:     "msg",
		LevelKey:       "level",
		TimeKey:        "time",
		NameKey:        "logger",
		CallerKey:      "caller",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.CapitalColorLevelEncoder,
		EncodeTime:     CustomEncodeTime,
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.FullCallerEncoder,
	}
	if global.CONFIG.Log.Format == "json" {
		config.EncodeLevel = zapcore.CapitalLevelEncoder
	}
	return config
}

func getEncoder() zapcore.Encoder {
	if global.CONFIG.Log.Format == "json" {
		return zapcore.NewJSONEncoder(getEncodeConfig())
	}
	return zapcore.NewConsoleEncoder(getEncodeConfig())
}

func getCore(level zapcore.Level) (core zapcore.Core) {
	writer, err := FileRotateLog.GetWriterSyncer(level.String())
	if err != nil {
		fmt.Printf("get writer syncer failed err:%v\n", err)
		return
	}
	return zapcore.NewCore(getEncoder(), writer, level)
}

func CustomEncodeTime(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
	enc.AppendString(t.Format(time.RFC3339))
}
`},
		fileItem{
			Path:     "core",
			Filename: "viper.go",
			Text: `package core

import (
	"{{projectName}}/global"
	"fmt"

	"github.com/spf13/viper"
)

func Viper() *viper.Viper {
	v := viper.New()
	v.SetConfigFile("./config.yaml")
	v.SetConfigName("config")
	v.SetConfigType("yaml")
	v.AddConfigPath(".")
	err := v.ReadInConfig()
	if err != nil {
		panic(fmt.Sprintf("read config err:%+v\n", err))
	}
	if err = v.Unmarshal(&global.CONFIG); err != nil {
		fmt.Println(err)
	}
	return v
}
`},
		fileItem{
			Path:     "core",
			Filename: "server.go",
			Text: `package core

import (
	"{{projectName}}/global"
	"{{projectName}}/initialiaze"
	"{{projectName}}/utils"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type server interface {
	ListenAndServe() error
	ListenAndServeTLS(certFile string, keyFile string) error
}

func RunServer() {
	Router := initialiaze.Routers()
	address := fmt.Sprintf(":%d", global.CONFIG.System.Port)
	s := initServer(address, Router)
	time.Sleep(time.Millisecond * 100)
	protocol := "http://"
	fmt.Printf("############################################################################\n")
	fmt.Printf("\n        Project Start Success")
	fmt.Printf("\n        Server Listen at:")
	for _, ip := range utils.GetLocalIP() {
		fmt.Printf("\n        %s%s%s", protocol, ip, address)
	}
	fmt.Printf("\n\n############################################################################\n")
	err := s.ListenAndServe()
	if err != nil {
		fmt.Println(err)
	}
}

func initServer(address string, router *gin.Engine) server {
	return &http.Server{
		Addr:           address,
		Handler:        router,
		ReadTimeout:    20 * time.Second,
		WriteTimeout:   20 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}
}
`},
		fileItem{
			Path:     "core",
			Filename: "log_file_write.go",
			Text: `package core

import (
	"{{projectName}}/global"
	"os"
	"path"
	"time"

	rotatelogs "github.com/lestrrat-go/file-rotatelogs"
	"go.uber.org/zap/zapcore"
)

var FileRotateLog = new(fileWriteLog)

const (
	Day = time.Hour * 24
	_   = iota
	KB  = 1 << (10 * iota)
	MB  = 1 << (10 * iota)
	GB  = 1 << (10 * iota)
)

type fileWriteLog struct{}

// GetWriterSyncer 根据日志不同级别获取写入的文件流
func (f *fileWriteLog) GetWriterSyncer(level string) (zapcore.WriteSyncer, error) {
	fileWrite, err := rotatelogs.New(
		path.Join(global.CONFIG.Log.Dir, "%Y-%m-%d", level+".log"),
		rotatelogs.ForceNewFile(),
		rotatelogs.WithClock(rotatelogs.Local),
		rotatelogs.WithMaxAge(time.Duration(global.CONFIG.Log.MaxRetentionDays)*Day),
		rotatelogs.WithRotationTime(Day*1),
		rotatelogs.WithRotationSize(MB*100),
	)
	if global.CONFIG.Log.ConsoleOutput {
		return zapcore.NewMultiWriteSyncer(zapcore.AddSync(os.Stdout), zapcore.AddSync(fileWrite)), err
	}
	return zapcore.AddSync(fileWrite), err
}
`},
		fileItem{
			Path:     "initialiaze",
			Filename: "router.go",
			Text: `package initialiaze

import (
	"{{projectName}}/middleware"
	"net/http"

	"github.com/gin-gonic/gin"
)

func Routers() *gin.Engine {
	Router := gin.New()
	Router.Use(middleware.ZapRecovery(true))
	Router.Use(middleware.ZapLogger())

	PublicGroup := Router.Group("")
	{
		PublicGroup.GET("/health", func(c *gin.Context) { c.String(http.StatusOK, "ok") })
	}
	//{
		//otherRouter.InitOtherRouter(PublicGroup)
	//}
	//PrivateGroup := Router.Group("")
	return Router
}
`},
		fileItem{
			Path:     "middleware",
			Filename: "logger.go",
			Text: `package middleware

import (
	"{{projectName}}/global"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func ZapLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery
		c.Next()
		cost := time.Since(start)
		global.LOG.Info(path,
			zap.String("path", path),
			zap.Int("status", c.Writer.Status()),
			zap.String("method", c.Request.Method),
			zap.String("query", query),
			zap.String("ip", c.ClientIP()),
			zap.String("errors", c.Errors.ByType(gin.ErrorTypePrivate).String()),
			zap.String("cost", cost.String()),
		)
	}
}
`},
		fileItem{
			Path:     "middleware",
			Filename: "recovery.go",
			Text: `package middleware

import (
	"{{projectName}}/global"
	"net"
	"net/http"
	"net/http/httputil"
	"os"
	"runtime/debug"
	"strings"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func ZapRecovery(stack bool) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				// Check for a broken connection, as it is not really a
				// condition that warrants a panic stack trace.
				var brokenPipe bool
				if ne, ok := err.(*net.OpError); ok {
					if se, ok := ne.Err.(*os.SyscallError); ok {
						if strings.Contains(strings.ToLower(se.Error()), "broken pipe") || strings.Contains(strings.ToLower(se.Error()), "connection reset by peer") {
							brokenPipe = true
						}
					}
				}

				httpRequest, _ := httputil.DumpRequest(c.Request, false)
				if brokenPipe {
					global.LOG.Error(c.Request.URL.Path,
						zap.Any("error", err),
						zap.String("request", string(httpRequest)),
					)
					// If the connection is dead, we can't write a status to it.
					_ = c.Error(err.(error)) // nolint: errcheck
					c.Abort()
					return
				}

				if stack {
					global.LOG.Error("[Recovery from panic]",
						zap.Any("error", err),
						zap.String("request", string(httpRequest)),
						zap.String("stack", string(debug.Stack())),
					)
				} else {
					global.LOG.Error("[Recovery from panic]",
						zap.Any("error", err),
						zap.String("request", string(httpRequest)),
					)
				}
				c.AbortWithStatus(http.StatusInternalServerError)
			}
		}()
		c.Next()
	}
}
`},
		fileItem{
			Path:     "config",
			Filename: "config.go",
			Text:     "package config\n\ntype Server struct {\n\tSystem System `mapstructure:\"system\" json:\"system\" yaml:\"system\"`\n\tLog    Log    `mapstructure:\"log\" json:\"log\" yaml:\"log\"`\n}\n"},
		fileItem{
			Path:     "config",
			Filename: "log.go",
			Text:     "package config\n\n// Log 日志配置\ntype Log struct {\n\tLevel            string `mapstructure:\"level\" json:\"level\" yaml:\"level\"`\n\tFormat           string `mapstructure:\"format\" json:\"format\" yaml:\"format\"`\n\tDir              string `mapstructure:\"dir\" json:\"dir\" yaml:\"dir\"`\n\tMaxRetentionDays int    `mapstructure:\"max_retention_days\" json:\"max_retention_days\" yaml:\"max_retention_days\"`\n\tShowLine         bool   `mapstructure:\"show_line\" json:\"show_line\" yaml:\"show_line\"`\n\tConsoleOutput    bool   `mapstructure:\"console_output\" json:\"console_output\" yaml:\"console_output\"`\n}\n"},
		fileItem{
			Path:     "config",
			Filename: "system.go",
			Text:     "package config\n\n// System 系统配置\ntype System struct {\n\t// 端口\n\tPort int `mapstructure:\"port\" json:\"port\" yaml:\"port\"`\n}\n"},
		fileItem{
			Path:     "utils",
			Filename: "local_ip.go",
			Text: `package utils

import (
	"fmt"
	"net"
)

func GetLocalIP() []string {
	ips := make([]string, 0, 4)
	netInterfaces, err := net.Interfaces()
	if err != nil {
		fmt.Println("net.Interfaces failed, err:", err.Error())
	}
	for i := 0; i < len(netInterfaces); i++ {
		if (netInterfaces[i].Flags & net.FlagUp) != 0 {
			addrList, _ := netInterfaces[i].Addrs()
			for _, address := range addrList {
				if ipNet, ok := address.(*net.IPNet); ok && !ipNet.IP.IsLoopback() {
					if ipNet.IP.To4() != nil {
						ips = append(ips, ipNet.IP.String())
					}
				}
			}
		}
	}
	return ips
}
`},
		fileItem{
			Path:     "",
			Filename: "config.yaml",
			Text: `# System configuration
system:
  port: 8000

#log configuration
log:
  level: "debug"          # 日志级别 值：
  format: "console"       # 日志输出
  dir: "log"              # 存放日志目录
  max_retention_days: 30  # 日志保留最长时间
  show_line: true         # 显示打印日志的行数
  console_output: true    # 控制台输出`},
		fileItem{
			Path:     "",
			Filename: "main.go",
			Text: `package main

import (
	"{{projectName}}/core"
	"{{projectName}}/global"

	"go.uber.org/zap"
)

func main() {

	global.VIPER = core.Viper()

	global.LOG = core.Zap()
	zap.ReplaceGlobals(global.LOG)

	core.RunServer()
}
`},
		fileItem{
			Path:     "cmd",
			Filename: "main.go",
			Text: `package main

import (
	"fmt"
	"os"
	"os/exec"
	"time"
)

func main() {
	if len(os.Args) <= 1 {
		fmt.Println("no executable file path")
		return
	}
	cmdPath := os.Args[1]
	params := make([]string, 0, 4)
	if len(os.Args) >= 2 {
		params = os.Args[2:]
	}
	for true {
		cmd := exec.Command(cmdPath, params...)
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		err := cmd.Start()
		if err != nil {
			fmt.Printf("[%s] %s 参数:%+v", time.Now().Format("2006-01-02 15:04:05.000"), cmdPath, params)
			fmt.Printf(" 启动错误:%+v\n", err)
			return
		}
		fmt.Printf("[%s] 启动成功：PID:%d\n", time.Now().Format("2006-01-02 15:04:05.000"), cmd.Process.Pid)
		err = cmd.Wait()
		if err == nil {
			return
		}
		fmt.Printf("[%s] %s 参数:%+v", time.Now().Format("2006-01-02 15:04:05.000"), cmdPath, params)
		fmt.Printf(" 意外终止:%+v\n", err)
		interval := time.Second * 5
		fmt.Printf("[%s] %s后重启命令\n", time.Now().Format("2006-01-02 15:04:05.000"), interval)
		time.Sleep(interval)
	}
}
`})
}

var (
	maxLength = 0
)

// GenerateCode 生成默认代码结构
func GenerateCode(projectName string) {
	rootPath := path.Join("./", projectName)
	for _, item := range fileList {
		length := len(path.Join(rootPath, item.Path, item.Filename))
		if maxLength < length {
			maxLength = length
		}
	}
	for _, item := range fileList {
		item.Text = strings.ReplaceAll(item.Text, "{{projectName}}", projectName)
		item.Path = path.Join(rootPath, item.Path, item.Filename)
		err := os.WriteFile(item.Path, bytes.NewBufferString(item.Text).Bytes(), 0744)
		if err != nil {
			log.Printf(FormatString("error"), item.Path, err)
		} else {
			log.Printf(FormatString("success"), item.Path)
		}
	}
}

func FormatString(status string) string {
	if status == "error" {
		return fmt.Sprintf("Generate path file:%%-%ds error:%%+v\n", maxLength+1)
	}
	if status == "success" {
		return fmt.Sprintf("Generate path file:%%-%ds success\n", maxLength+1)
	}
	return ""
}
