package service

import (
	"context"
	"github.com/go-redis/redis"
	api "github.com/hpifu/go-godtoken/api"
	"github.com/hpifu/go-kit/hrand"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/peer"
	"time"
)

var InfoLog *logrus.Logger = logrus.New()
var WarnLog *logrus.Logger = logrus.New()
var AccessLog *logrus.Logger = logrus.New()

func NewService(rc *redis.Client) *Service {
	return &Service{
		rc: rc,
	}
}

type Service struct {
	rc *redis.Client
}

func (s *Service) GetToken(ctx context.Context, req *api.GetTokenReq) (*api.GetTokenRes, error) {
	token, err := s.rc.Get("godtoken").Result()
	if err != nil && err != redis.Nil {
		return nil, err
	}
	if err == redis.Nil || token == "" {
		token = hrand.NewToken()
		if err := s.rc.Set("godtoken", token, 5*time.Second).Err(); err != nil {
			return nil, err
		}
	}

	res := &api.GetTokenRes{
		Token: token,
	}

	return res, nil
}

func (s *Service) Verify(ctx context.Context, req *api.VerifyReq) (*api.VerifyRes, error) {
	token, err := s.rc.Get("godtoken").Result()
	if err != nil && err != redis.Nil {
		return nil, err
	}

	res := &api.VerifyRes{
		Ok: token == req.Token,
	}

	return res, nil
}

func Interceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
	ts := time.Now()
	p, ok := peer.FromContext(ctx)
	clientIP := ""
	if ok && p != nil {
		clientIP = p.Addr.String()
	}

	res, err := handler(ctx, req)

	AccessLog.WithFields(logrus.Fields{
		"client":    clientIP,
		"url":       info.FullMethod,
		"req":       req,
		"res":       res,
		"err":       err,
		"resTimeNs": time.Now().Sub(ts).Nanoseconds(),
	}).Info()

	return res, err
}
