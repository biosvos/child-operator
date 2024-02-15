package flow

import (
	"github.com/pkg/errors"
)

type Service struct{}

func NewService() *Service {
	return &Service{}
}

func (s *Service) Work() error {
	return errors.New("failed to run")
}
