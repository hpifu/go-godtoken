package main

import (
	"flag"
	"fmt"
	"github.com/hpifu/go-kit/hgrpc"
	"github.com/sirupsen/logrus"
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
	"github.com/olivere/elastic/v7"
	"github.com/spf13/viper"
	"google.golang.org/grpc"
	"gopkg.in/sohlich/elogrus.v7"
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
	config.SetEnvPrefix("godtoken")
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
	infoLog, warnLog, accessLog, err := logger.NewLoggerGroupWithViper(config.Sub("logger"))
	if err != nil {
		panic(err)
	}
	client, err := elastic.NewClient(
		elastic.SetURL(config.GetString("es.uri")),
		elastic.SetSniff(false),
	)
	if err != nil {
		panic(err)
	}
	hook, err := elogrus.NewAsyncElasticHook(client, "go-godtoken", logrus.InfoLevel, "go-godtoken-log")
	if err != nil {
		panic(err)
	}
	accessLog.Hooks.Add(hook)

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

	infoLog.Infof("%v init success, port[%v]", os.Args[0], config.GetInt("service.port"))

	interceptor := hgrpc.NewGrpcInterceptor(infoLog, warnLog, accessLog)
	// run server
	server := grpc.NewServer(
		grpc.UnaryInterceptor(interceptor.Interceptor),
	)
	go func() {
		svc := service.NewService(rc)
		svc.SetLogger(infoLog, warnLog, accessLog)
		api.RegisterServiceServer(server, svc)
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
	_ = warnLog.Out.(*rotatelogs.RotateLogs).Close()
	_ = accessLog.Out.(*rotatelogs.RotateLogs).Close()
	infoLog.Errorf("%v shutdown success", os.Args[0])
	_ = infoLog.Out.(*rotatelogs.RotateLogs).Close()
}
