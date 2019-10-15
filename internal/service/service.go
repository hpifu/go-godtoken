package service

import (
	"context"
	api "github.com/hpifu/go-godtoken/api"
	"github.com/sirupsen/logrus"
)

var InfoLog *logrus.Logger = logrus.New()
var WarnLog *logrus.Logger = logrus.New()
var AccessLog *logrus.Logger = logrus.New()

func NewService() *Service {
	return &Service{}
}

type Service struct{}

func (s *Service) Do(ctx context.Context, request *api.Request) (*api.Response, error) {
	response := &api.Response{
		Message: request.Message,
	}

	AccessLog.WithFields(logrus.Fields{
		"request":  request,
		"response": response,
	}).Info()

	return response, nil
}
