package service

import (
	"TaskService/internal/service/task"
	"TaskService/internal/storage"
	"TaskService/pkg/kafka"
)

type Service interface {
	Task() task.Service
}

type service struct {
	task task.Service
}

func New(st storage.Storage, kc kafka.Kafka) Service {

	result := &service{
		task: task.New(st, kc),
	}

	return result
}

func (s *service) Task() task.Service {
	return s.task
}
