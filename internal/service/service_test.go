package service_test

import (
	"TaskService/internal/dto"
	"TaskService/internal/model"
	"TaskService/internal/service/task"
	"TaskService/internal/storage/postgres"
	"TaskService/pkg/logger"
	"context"
	"database/sql"
	"fmt"
	"os"
	"testing"

	"github.com/IBM/sarama"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Mock Storage
type MockStorage struct {
	mock.Mock
}

func (m *MockStorage) DB() postgres.Storage {
	args := m.Called()
	return args.Get(0).(postgres.Storage)
}

// Mock PostgresStorage
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

// Mock Kafka
type MockKafka struct {
	mock.Mock
}

func (m *MockKafka) SendMessage(message []byte) error {
	args := m.Called(message)
	return args.Error(0)
}

func (m *MockKafka) ConsumeMessages(handler func(message *sarama.ConsumerMessage)) error {
	args := m.Called(handler)
	return args.Error(0)
}

func (m *MockKafka) Close() error {
	args := m.Called()
	return args.Error(0)
}

// MockTx реализует postgres.Tx интерфейс
type MockTx struct {
	mock.Mock
}

func (m *MockTx) Commit() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockTx) Rollback() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockTx) ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	mockArgs := m.Called(ctx, query, args)
	return mockArgs.Get(0).(sql.Result), mockArgs.Error(1)
}

func (m *MockTx) QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row {
	mockArgs := m.Called(ctx, query, args)
	return mockArgs.Get(0).(*sql.Row)
}

// MockResult для реализации sql.Result
type MockResult struct {
	mock.Mock
}

func (m *MockResult) LastInsertId() (int64, error) {
	args := m.Called()
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockResult) RowsAffected() (int64, error) {
	args := m.Called()
	return args.Get(0).(int64), args.Error(1)
}

func TestMain(m *testing.M) {
	loggerCfg := logger.Config{
		Level:             logger.LevelDebug,
		DuplicateToStdout: true,
		Dir:               "",
	}
	if err := logger.Init(loggerCfg); err != nil {
		fmt.Println("failed to init logger")
	}

	code := m.Run()

	os.Exit(code)
}

// setupTest создает моки для каждого теста
func setupTest(t *testing.T) (*MockStorage, *MockPostgresStorage, *MockKafka, *MockTx) {
	t.Helper()
	return &MockStorage{}, &MockPostgresStorage{}, &MockKafka{}, &MockTx{}
}

func TestTaskService_Get(t *testing.T) {
	mockStorage, mockPostgres, mockKafka, _ := setupTest(t)

	ctx := context.Background()
	taskID := 1
	expectedTask := model.Task{
		ID:          1,
		Title:       "Test Task",
		Description: "Test Description",
		Status:      "created",
	}

	mockStorage.On("DB").Return(mockPostgres)
	mockPostgres.On("Get", ctx, taskID).Return(expectedTask, nil)

	service := task.New(mockStorage, mockKafka)
	result, err := service.Get(ctx, taskID)

	assert.NoError(t, err)
	assert.Equal(t, expectedTask.ID, result.ID)
	assert.Equal(t, expectedTask.Title, result.Title)
	assert.Equal(t, expectedTask.Description, result.Description)
	assert.Equal(t, expectedTask.Status, result.Status)

	mockStorage.AssertExpectations(t)
	mockPostgres.AssertExpectations(t)
}

func TestTaskService_GetList(t *testing.T) {
	mockStorage, mockPostgres, mockKafka, _ := setupTest(t)

	ctx := context.Background()
	expectedTasks := []model.Task{
		{ID: 1, Title: "Task 1", Description: "Desc 1", Status: "created"},
		{ID: 2, Title: "Task 2", Description: "Desc 2", Status: "done"},
	}

	mockStorage.On("DB").Return(mockPostgres)
	mockPostgres.On("GetList", ctx).Return(expectedTasks, nil)

	service := task.New(mockStorage, mockKafka)
	result, err := service.GetList(ctx)

	assert.NoError(t, err)
	assert.Len(t, result.Tasks, 2)
	assert.Equal(t, expectedTasks[0].ID, result.Tasks[0].ID)
	assert.Equal(t, expectedTasks[1].Title, result.Tasks[1].Title)

	mockStorage.AssertExpectations(t)
	mockPostgres.AssertExpectations(t)
}

func TestTaskService_Create(t *testing.T) {
	mockStorage, mockPostgres, mockKafka, mockTx := setupTest(t)

	ctx := context.Background()
	createReq := dto.CreateTaskRequest{
		Title:       "New Task",
		Description: "New Description",
	}
	expectedID := 1

	mockStorage.On("DB").Return(mockPostgres)
	mockPostgres.On("BeginTx", ctx).Return(mockTx, nil)
	mockPostgres.On("Create", ctx, mockTx, model.Task{
		Title:       createReq.Title,
		Description: createReq.Description,
	}).Return(expectedID, nil)
	mockKafka.On("SendMessage", mock.Anything).Return(nil)
	mockTx.On("Commit").Return(nil)
	mockTx.On("Rollback").Return(nil)

	service := task.New(mockStorage, mockKafka)
	err := service.Create(ctx, createReq)

	assert.NoError(t, err)
	mockStorage.AssertExpectations(t)
	mockPostgres.AssertExpectations(t)
	mockKafka.AssertExpectations(t)
	mockTx.AssertExpectations(t)
}

func TestTaskService_Update(t *testing.T) {
	mockStorage, mockPostgres, mockKafka, mockTx := setupTest(t)

	ctx := context.Background()
	updateReq := dto.UpdateTaskRequest{
		ID:          1,
		Title:       "Updated Task",
		Description: "Updated Description",
		Status:      "done",
	}

	// ДОБАВЛЕНО: моки для транзакции в Update
	mockStorage.On("DB").Return(mockPostgres)
	mockPostgres.On("BeginTx", ctx).Return(mockTx, nil)
	mockPostgres.On("Update", ctx, mockTx, model.Task{
		ID:          updateReq.ID,
		Title:       updateReq.Title,
		Description: updateReq.Description,
		Status:      updateReq.Status,
	}).Return(nil)
	mockTx.On("Commit").Return(nil)
	mockTx.On("Rollback").Return(nil)

	service := task.New(mockStorage, mockKafka)
	err := service.Update(ctx, updateReq)

	assert.NoError(t, err)
	mockStorage.AssertExpectations(t)
	mockPostgres.AssertExpectations(t)
	mockTx.AssertExpectations(t) // ДОБАВЛЕНО: проверка мока транзакции
}

func TestTaskService_Update_InvalidStatus(t *testing.T) {
	mockStorage, _, mockKafka, _ := setupTest(t)

	ctx := context.Background()
	updateReq := dto.UpdateTaskRequest{
		ID:          1,
		Title:       "Updated Task",
		Description: "Updated Description",
		Status:      "invalid_status",
	}

	service := task.New(mockStorage, mockKafka)
	err := service.Update(ctx, updateReq)

	assert.Error(t, err)
	assert.Equal(t, "invalid status", err.Error())
}
