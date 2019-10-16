package service

import (
	"context"
	"time"

	"github.com/hpifu/go-kit/hrand"

	"github.com/go-redis/redis"
	api "github.com/hpifu/go-godtoken/api"
	"github.com/sirupsen/logrus"
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

	AccessLog.WithFields(logrus.Fields{
		"req": req,
		"res": res,
	}).Info()

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

	AccessLog.WithFields(logrus.Fields{
		"req": req,
		"res": res,
	}).Info()

	return res, nil
}
