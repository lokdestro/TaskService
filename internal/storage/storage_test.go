package storage

import (
	"TaskService/internal/model"
	"TaskService/internal/storage/postgres"
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockPostgresStorage struct {
	mock.Mock
}

func (m *MockPostgresStorage) Get(ctx context.Context, id int) (model.Task, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(model.Task), args.Error(1)
}

func (m *MockPostgresStorage) GetList(ctx context.Context) ([]model.Task, error) {
	args := m.Called(ctx)
	return args.Get(0).([]model.Task), args.Error(1)
}

func (m *MockPostgresStorage) Update(ctx context.Context, tx postgres.Tx, req model.Task) error {
	args := m.Called(ctx, tx, req)
	return args.Error(0)
}

func (m *MockPostgresStorage) Create(ctx context.Context, tx postgres.Tx, task model.Task) (int, error) {
	args := m.Called(ctx, tx, task)
	return args.Int(0), args.Error(1)
}

func (m *MockPostgresStorage) BeginTx(ctx context.Context) (postgres.Tx, error) {
	args := m.Called(ctx)
	return args.Get(0).(postgres.Tx), args.Error(1)
}

func TestStorage_DB(t *testing.T) {
	mockPostgres := new(MockPostgresStorage)
	storage := &repo{psql: mockPostgres}

	result := storage.DB()

	assert.Equal(t, mockPostgres, result)
}

func TestNew(t *testing.T) {
	cfg := postgres.Config{
		URL:    "invalid_url",
		Driver: "postgres",
	}

	_, err := New(cfg)
	assert.Error(t, err)
}
