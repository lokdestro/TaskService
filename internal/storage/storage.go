package storage

import (
	"TaskService/internal/storage/postgres"
)

//go:generate mockery --name=Storage --dir=. --output=./mocks
type Storage interface {
	DB() postgres.Storage

}

type repo struct {
	psql postgres.Storage
}

func (r *repo) DB() postgres.Storage {
	return r.psql
}

func New(pcfg postgres.Config) (Storage, error) {
	psql, err := postgres.New(pcfg)
	if err != nil {
		return nil, err
	}

	result := &repo{
		psql: psql,
	}

	return result, nil
}
