package container

import (
	"cinema/config"
	"cinema/internal/database"
	"cinema/internal/logger"
)

type Container interface {
	GetConfig() *config.Config
	GetLogger() logger.Logger
	GetRepository() database.Repository
}

type container struct {
	config *config.Config
	rep    database.Repository
	logger logger.Logger
}

func NewContainer(rep database.Repository, logger logger.Logger, conf config.Config) (Container, error) {
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

func (c *container) GetRepository() database.Repository {
	return c.rep
}
