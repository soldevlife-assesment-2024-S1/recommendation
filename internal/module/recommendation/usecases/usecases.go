package usecases

import (
	"recommendation-service/internal/module/recommendation/repositories"
)

type usecases struct {
	repo repositories.Repositories
}

type Usecases interface {
}

func New(repo repositories.Repositories) Usecases {
	return &usecases{
		repo: repo,
	}
}
