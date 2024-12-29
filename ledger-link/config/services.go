package config

import (
	"context"

	"gorm.io/gorm"

	"ledger-link/internal/repositories"
	"ledger-link/internal/services"
	"ledger-link/pkg/logger"
)

type ServiceContainer struct {
	db  *gorm.DB
	log *logger.Logger

	UserRepo *repositories.UserRepository
	UserService *services.UserService
}

func NewServiceContainer(db *gorm.DB, log *logger.Logger) *ServiceContainer {
	container := &ServiceContainer{
		db:  db,
		log: log,
	}
	container.initRepositories()
	container.initServices()
	return container
}

func (c *ServiceContainer) initRepositories() {
	c.UserRepo = repositories.NewUserRepository(c.db)
}

func (c *ServiceContainer) initServices() {
	c.UserService = services.NewUserService(c.UserRepo, nil, nil, c.log)
}

func (c *ServiceContainer) Start(ctx context.Context) error {
	return nil
}

func (c *ServiceContainer) Stop() {
}