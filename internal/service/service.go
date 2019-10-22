package service

import (
	"context"
	"time"

	"github.com/go-redis/redis"
	api "github.com/hpifu/go-godtoken/api"
	"github.com/hpifu/go-kit/hrand"
	"github.com/sirupsen/logrus"
)

func NewService(rc *redis.Client) *Service {
	return &Service{
		rc:        rc,
		infoLog:   logrus.New(),
		warnLog:   logrus.New(),
		accessLog: logrus.New(),
	}
}

func (s *Service) SetLogger(infoLog, warnLog, accessLog *logrus.Logger) {
	s.infoLog = infoLog
	s.warnLog = warnLog
	s.accessLog = accessLog
}

type Service struct {
	rc        *redis.Client
	infoLog   *logrus.Logger
	warnLog   *logrus.Logger
	accessLog *logrus.Logger
}

func (s *Service) GetToken(ctx context.Context, req *api.GetTokenReq) (*api.GetTokenRes, error) {
	token, err := s.rc.Get("godtoken").Result()
	if err != nil && err != redis.Nil {
		return nil, err
	}
	if err == redis.Nil || token == "" {
		token = hrand.NewToken()
		if err := s.rc.Set("godtoken", token, 15*time.Minute).Err(); err != nil {
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
