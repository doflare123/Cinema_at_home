package container

import (
	"cinema/config"
	"cinema/internal/logger"
	"cinema/internal/repository"
)

type Container interface {
	GetConfig() *config.Config
	GetLogger() logger.Logger
	GetRepository() repository.Repository
}

type container struct {
	config *config.Config
	rep    repository.Repository
	logger logger.Logger
}

func NewContainer(rep repository.Repository, logger logger.Logger, conf config.Config) (Container, error) {
	return &container{
		config: &conf,
		rep:    rep,
		logger: logger,
	}, nil
}

func (c *container) GetConfig() *config.Config {
	return c.config
}

func (c *container) GetLogger() logger.Logger {
	return c.logger
}

func (c *container) GetRepository() repository.Repository {
	return c.rep
}
