package main

import (
	"flag"
	"fmt"
	"net"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/go-redis/redis"
	api "github.com/hpifu/go-godtoken/api"
	"github.com/hpifu/go-godtoken/internal/service"
	"github.com/hpifu/go-kit/logger"
	rotatelogs "github.com/lestrrat-go/file-rotatelogs"
	"github.com/spf13/viper"
	"google.golang.org/grpc"
)

// AppVersion name
var AppVersion = "unknown"

func main() {
	version := flag.Bool("v", false, "print current version")
	configfile := flag.String("c", "configs/godtoken.json", "config file path")
	flag.Parse()
	if *version {
		fmt.Println(AppVersion)
		os.Exit(0)
	}

	// load config
	config := viper.New()
	config.SetEnvPrefix("account")
	config.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	config.AutomaticEnv()
	config.SetConfigType("json")
	fp, err := os.Open(*configfile)
	if err != nil {
		panic(err)
	}
	err = config.ReadConfig(fp)
	if err != nil {
		panic(err)
	}

	// init logger
	infoLog, err := logger.NewTextLoggerWithViper(config.Sub("logger.infoLog"))
	if err != nil {
		panic(err)
	}
	warnLog, err := logger.NewTextLoggerWithViper(config.Sub("logger.warnLog"))
	if err != nil {
		panic(err)
	}
	accessLog, err := logger.NewJsonLoggerWithViper(config.Sub("logger.accessLog"))
	if err != nil {
		panic(err)
	}

	service.InfoLog = infoLog
	service.WarnLog = warnLog
	service.AccessLog = accessLog

	// init redis
	option := &redis.Options{
		Addr:         config.GetString("redis.addr"),
		DialTimeout:  config.GetDuration("redis.dialTimeout"),
		ReadTimeout:  config.GetDuration("redis.readTimeout"),
		WriteTimeout: config.GetDuration("redis.writeTimeout"),
		MaxRetries:   config.GetInt("redis.maxRetries"),
		PoolSize:     config.GetInt("redis.poolSize"),
		Password:     config.GetString("redis.password"),
		DB:           config.GetInt("redis.db"),
	}
	rc := redis.NewClient(option)
	if err := rc.Ping().Err(); err != nil {
		panic(err)
	}
	infoLog.Infof("init redis success. option [%#v]", option)

	infoLog.Infof("%v init success, port[%v]", os.Args[0], config.GetString("service.port"))

	// run server
	server := grpc.NewServer(
		grpc.UnaryInterceptor(service.Interceptor),
	)
	go func() {
		api.RegisterServiceServer(server, service.NewService(rc))
		address, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%v", config.GetInt("service.port")))
		if err != nil {
			panic(err)
		}

		if err := server.Serve(address); err != nil {
			panic(err)
		}
	}()

	// graceful quit
	quit := make(chan os.Signal)
	signal.Notify(quit, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	infoLog.Infof("%v shutdown ...", os.Args[0])
	server.GracefulStop()

	// close loggers
	warnLog.Out.(*rotatelogs.RotateLogs).Close()
	accessLog.Out.(*rotatelogs.RotateLogs).Close()
	infoLog.Errorf("%v shutdown success", os.Args[0])
	infoLog.Out.(*rotatelogs.RotateLogs).Close()
}
