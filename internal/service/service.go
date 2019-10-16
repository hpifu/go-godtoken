package service

import (
	"context"

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
	res := &api.GetTokenRes{
		Token: "123",
	}

	AccessLog.WithFields(logrus.Fields{
		"req": req,
		"res": res,
	}).Info()

	return res, nil
}

func (s *Service) Verify(ctx context.Context, req *api.VerifyReq) (*api.VerifyRes, error) {
	res := &api.VerifyRes{
		Ok: true,
	}

	return res, nil
}
