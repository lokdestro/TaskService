package task

import (
	"TaskService/internal/dto"
	"TaskService/internal/model"
	"TaskService/internal/storage"
	"TaskService/pkg/kafka"
	"TaskService/pkg/logger"
	"context"
	"encoding/json"
	"errors"
	"github.com/IBM/sarama"
	"strconv"
	"time"

	"github.com/gammazero/workerpool"
)

var ErrInvalidStatus = errors.New("invalid status")

type Service interface {
	Get(ctx context.Context, id int) (dto.GetTaskResponse, error)
	GetList(ctx context.Context) (dto.GetTaskListResponse, error)
	Update(ctx context.Context, req dto.UpdateTaskRequest) error
	Create(ctx context.Context, req dto.CreateTaskRequest) error
	ProcessTasks()
}

type service struct {
	st storage.Storage
	kc kafka.Kafka
	wp *workerpool.WorkerPool
}

func New(st storage.Storage, kc kafka.Kafka) Service {
	result := &service{
		st: st,
		kc: kc,
		wp: workerpool.New(100),
	}

	return result
}

func (s *service) Get(ctx context.Context, id int) (dto.GetTaskResponse, error) {
	var resp dto.GetTaskResponse

	log := logger.Get()

	task, err := s.st.DB().Get(ctx, id)

	if err != nil {
		log.Info().Err(err).Msg("get task failed")
		return resp, err
	}

	resp = dto.GetTaskResponse{
		ID:          task.ID,
		Title:       task.Title,
		Description: task.Description,
		Status:      task.Status,
	}

	return resp, nil
}

func (s *service) GetList(ctx context.Context) (dto.GetTaskListResponse, error) {
	resp := dto.GetTaskListResponse{
		Tasks: make([]dto.GetTaskResponse, 0),
	}

	log := logger.Get()

	tasks, err := s.st.DB().GetList(ctx)
	if err != nil {
		log.Info().Err(err).Msg("get tasks failed")
		return resp, err
	}

	for _, task := range tasks {
		simpleTask := dto.GetTaskResponse{
			ID:          task.ID,
			Title:       task.Title,
			Description: task.Description,
			Status:      task.Status,
		}

		resp.Tasks = append(resp.Tasks, simpleTask)
	}

	return resp, nil
}

func (s *service) Update(ctx context.Context, req dto.UpdateTaskRequest) error {
	log := logger.Get()

	if err := validateStatus(req.Status); err != nil {
		log.Error().Err(err).Msg("validateStatus failed")
		return err
	}

	task := model.Task{
		ID:          req.ID,
		Title:       req.Title,
		Description: req.Description,
		Status:      req.Status,
	}

	err := s.st.DB().Update(ctx, task)
	if err != nil {
		log.Info().Err(err).Msg("update task failed")
	}

	return err
}

func (s *service) Create(ctx context.Context, req dto.CreateTaskRequest) error {
	log := logger.Get()

	tx, err := s.st.DB().BeginTx(ctx)
	defer tx.Rollback()

	task := model.Task{
		Title:       req.Title,
		Description: req.Description,
	}

	id, err := s.st.DB().Create(ctx, tx, task)
	if err != nil {
		log.Info().Err(err).Msg("create task failed")
		return err
	}

	message, err := json.Marshal(id)
	if err != nil {
		log.Info().Err(err).Msg("create task failed")
		return err
	}

	err = s.kc.SendMessage(message)
	if err != nil {
		log.Info().Err(err).Msg("send message failed")
		return err
	}

	return tx.Commit()
}

func (s *service) ProcessTasks() {
	handler := func(message *sarama.ConsumerMessage) {
		var id int
		log := logger.Get()
		if err := json.Unmarshal(message.Value, &id); err != nil {
			return
		}

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		task, err := s.st.DB().Get(ctx, id)
		if err != nil {
			return
		}

		log.Info().
			Str("id", strconv.Itoa(id)).
			Str("title", task.Title).
			Str("description", task.Description).
			Msg("process task")

		if task.Status != "" {
			updateReq := dto.UpdateTaskRequest{
				ID:          task.ID,
				Title:       task.Title,
				Description: task.Description,
				Status:      "done",
			}
			if err := s.Update(ctx, updateReq); err != nil {
				log.Info().Err(err).Msg("update task failed")
			} else {
				log.Info().Msg("success")
			}
		}
	}

	go func() {
		if err := s.kc.ConsumeMessages(handler); err != nil {
		}
	}()
}

func validateStatus(status string) error {
	if status == "done" || status == "created" {
		return nil
	}
	return ErrInvalidStatus
}
