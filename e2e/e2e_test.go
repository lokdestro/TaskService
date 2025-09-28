package e2e

import (
	"TaskService/internal/dto"
	"TaskService/internal/service"
	"TaskService/internal/storage"
	"TaskService/internal/storage/postgres"
	"TaskService/pkg/kafka"
	"TaskService/pkg/logger"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"testing"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	testKfk "github.com/testcontainers/testcontainers-go/modules/kafka"
	testPsql "github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

var (
	dbContainer     *testPsql.PostgresContainer
	kafkaContainer  *testKfk.KafkaContainer
	storageInstance storage.Storage
	kafkaClient     kafka.Kafka
	taskService     service.Service
	dbURL           string
)

func TestMain(m *testing.M) {
	ctx := context.Background()

	loggerCfg := logger.Config{
		Level:             logger.LevelDebug,
		DuplicateToStdout: true,
		Dir:               "",
	}
	if err := logger.Init(loggerCfg); err != nil {
		fmt.Println("failed to init logger")
	}

	if err := setupContainers(ctx); err != nil {
		log.Fatalf("Failed to setup containers: %v", err)
	}
	defer teardownContainers(ctx)

	if err := setupApplication(ctx); err != nil {
		log.Fatalf("Failed to setup application: %v", err)
	}

	code := m.Run()

	os.Exit(code)
}

func setupContainers(ctx context.Context) error {
	var err error

	dbContainer, err = testPsql.Run(ctx,
		"postgres:15-alpine",
		testPsql.WithDatabase("testdb"),
		testPsql.WithUsername("testuser"),
		testPsql.WithPassword("testpass"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(30*time.Second)),
	)
	if err != nil {
		return fmt.Errorf("failed to start PostgreSQL container: %w", err)
	}

	kafkaContainer, err = testKfk.Run(ctx,
		"confluentinc/cp-kafka:7.4.0",
		testKfk.WithClusterID("test-cluster"),
	)
	if err != nil {
		return fmt.Errorf("failed to start Kafka container: %w", err)
	}

	time.Sleep(10 * time.Second)
	return nil
}

func setupApplication(ctx context.Context) error {
	var err error

	dbConnStr, err := dbContainer.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		return fmt.Errorf("failed to get DB connection string: %w", err)
	}
	dbURL = dbConnStr

	brokers, err := kafkaContainer.Brokers(ctx)
	if err != nil {
		return fmt.Errorf("failed to get Kafka brokers: %w", err)
	}

	if len(brokers) == 0 {
		return fmt.Errorf("no Kafka brokers available")
	}
	broker := brokers[0]

	if err := createTables(ctx); err != nil {
		return fmt.Errorf("failed to create tables: %w", err)
	}

	pgConfig := postgres.Config{
		URL:    dbURL,
		Driver: "postgres",
	}

	storageInstance, err = storage.New(pgConfig)
	if err != nil {
		return fmt.Errorf("failed to create storage: %w", err)
	}

	kafkaConfig := kafka.Config{
		Broker: broker,
		Topic:  "test-tasks",
	}

	kafkaClient, err = kafka.NewKafkaClient(kafkaConfig)
	if err != nil {
		return fmt.Errorf("failed to create Kafka client: %w", err)
	}

	taskService = service.New(storageInstance, kafkaClient)

	taskService.Task().ProcessTasks()

	return nil
}

func createTables(ctx context.Context) error {
	db, err := sqlx.Connect("postgres", dbURL)
	if err != nil {
		return err
	}
	defer db.Close()

	_, err = db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS tasks (
			id SERIAL PRIMARY KEY,
			title VARCHAR(255) NOT NULL,
			description TEXT,
			status VARCHAR(50) DEFAULT 'created'
		)
	`)
	return err
}

func cleanupDatabase(ctx context.Context) {
	db, err := sqlx.Connect("postgres", dbURL)
	if err != nil {
		log.Printf("Failed to connect for cleanup: %v", err)
		return
	}
	defer db.Close()

	_, err = db.ExecContext(ctx, "DELETE FROM tasks")
	if err != nil {
		log.Printf("Failed to cleanup database: %v", err)
	}
}

func teardownContainers(ctx context.Context) {
	if dbContainer != nil {
		if err := dbContainer.Terminate(ctx); err != nil {
			log.Printf("Failed to terminate DB container: %v", err)
		}
	}
	if kafkaContainer != nil {
		if err := kafkaContainer.Terminate(ctx); err != nil {
			log.Printf("Failed to terminate Kafka container: %v", err)
		}
	}
}

func TestTaskLifecycle(t *testing.T) {
	ctx := context.Background()
	cleanupDatabase(ctx)

	t.Run("CreateTask", func(t *testing.T) {
		req := dto.CreateTaskRequest{
			Title:       "E2E Test Task",
			Description: "E2E Test Description",
		}

		err := taskService.Task().Create(ctx, req)
		require.NoError(t, err)

		tasks, err := taskService.Task().GetList(ctx)
		require.NoError(t, err)
		require.Len(t, tasks.Tasks, 1)

		task := tasks.Tasks[0]
		assert.Equal(t, req.Title, task.Title)
		assert.Equal(t, req.Description, task.Description)
		assert.Equal(t, "created", task.Status)
	})

	t.Run("GetTask", func(t *testing.T) {
		tasks, err := taskService.Task().GetList(ctx)
		require.NoError(t, err)
		require.Len(t, tasks.Tasks, 1)

		taskID := tasks.Tasks[0].ID
		task, err := taskService.Task().Get(ctx, taskID)
		require.NoError(t, err)

		assert.Equal(t, taskID, task.ID)
		assert.Equal(t, "E2E Test Task", task.Title)
		assert.Equal(t, "created", task.Status)
	})

	t.Run("UpdateTask", func(t *testing.T) {
		tasks, err := taskService.Task().GetList(ctx)
		require.NoError(t, err)
		require.Len(t, tasks.Tasks, 1)

		taskID := tasks.Tasks[0].ID
		updateReq := dto.UpdateTaskRequest{
			ID:          taskID,
			Title:       "Updated E2E Task",
			Description: "Updated E2E Description",
			Status:      "done",
		}

		err = taskService.Task().Update(ctx, updateReq)
		require.NoError(t, err)

		task, err := taskService.Task().Get(ctx, taskID)
		require.NoError(t, err)

		assert.Equal(t, updateReq.Title, task.Title)
		assert.Equal(t, updateReq.Description, task.Description)
		assert.Equal(t, updateReq.Status, task.Status)
	})
}

func TestTaskKafkaIntegration(t *testing.T) {
	ctx := context.Background()
	cleanupDatabase(ctx)

	t.Run("KafkaMessageProcessing", func(t *testing.T) {
		createReq := dto.CreateTaskRequest{
			Title:       "Kafka Test Task",
			Description: "Kafka Test Description",
		}

		err := taskService.Task().Create(ctx, createReq)
		require.NoError(t, err)

		tasks, err := taskService.Task().GetList(ctx)
		require.NoError(t, err)
		require.Len(t, tasks.Tasks, 1)

		taskID := tasks.Tasks[0].ID

		message, err := json.Marshal(taskID)
		require.NoError(t, err)

		err = kafkaClient.SendMessage(message)
		require.NoError(t, err)

		time.Sleep(5 * time.Second)

		task, err := taskService.Task().Get(ctx, taskID)
		require.NoError(t, err)

		assert.Equal(t, "done", task.Status)
	})
}

func TestConcurrentTaskOperations(t *testing.T) {
	ctx := context.Background()
	cleanupDatabase(ctx)

	const numTasks = 10

	t.Run("ConcurrentCreate", func(t *testing.T) {
		done := make(chan bool, numTasks)
		errCh := make(chan error, numTasks)

		for i := 0; i < numTasks; i++ {
			go func(index int) {
				req := dto.CreateTaskRequest{
					Title:       fmt.Sprintf("Concurrent Task %d", index),
					Description: fmt.Sprintf("Concurrent Description %d", index),
				}

				err := taskService.Task().Create(ctx, req)
				if err != nil {
					errCh <- err
				} else {
					done <- true
				}
			}(i)
		}

		completed := 0
		timeout := time.After(30 * time.Second)

		for completed < numTasks {
			select {
			case <-done:
				completed++
			case err := <-errCh:
				t.Fatalf("Concurrent create failed: %v", err)
			case <-timeout:
				t.Fatalf("Timeout waiting for concurrent operations")
			}
		}

		tasks, err := taskService.Task().GetList(ctx)
		require.NoError(t, err)
		assert.Len(t, tasks.Tasks, numTasks)
	})
}

func TestErrorScenarios(t *testing.T) {
	ctx := context.Background()
	cleanupDatabase(ctx)

	t.Run("GetNonExistentTask", func(t *testing.T) {
		_, err := taskService.Task().Get(ctx, 99999)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "no rows")
	})

	t.Run("InvalidStatus", func(t *testing.T) {
		createReq := dto.CreateTaskRequest{
			Title: "Test Task for Invalid Status",
		}
		err := taskService.Task().Create(ctx, createReq)
		require.NoError(t, err)

		tasks, err := taskService.Task().GetList(ctx)
		require.NoError(t, err)
		taskID := tasks.Tasks[0].ID

		updateReq := dto.UpdateTaskRequest{
			ID:     taskID,
			Title:  "Test Task",
			Status: "invalid_status",
		}

		err = taskService.Task().Update(ctx, updateReq)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "invalid status")
	})
}

func TestPerformance(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping performance test in short mode")
	}

	ctx := context.Background()
	cleanupDatabase(ctx)

	const batchSize = 100

	t.Run("BatchOperations", func(t *testing.T) {
		start := time.Now()

		for i := 0; i < batchSize; i++ {
			req := dto.CreateTaskRequest{
				Title:       fmt.Sprintf("Performance Task %d", i),
				Description: fmt.Sprintf("Performance Description %d", i),
			}

			err := taskService.Task().Create(ctx, req)
			require.NoError(t, err)
		}

		createDuration := time.Since(start)
		t.Logf("Created %d tasks in %v", batchSize, createDuration)

		start = time.Now()
		tasks, err := taskService.Task().GetList(ctx)
		require.NoError(t, err)
		readDuration := time.Since(start)

		assert.Len(t, tasks.Tasks, batchSize)
		t.Logf("Read %d tasks in %v", batchSize, readDuration)

		assert.Less(t, createDuration, 10*time.Second, "Create operations took too long")
		assert.Less(t, readDuration, 5*time.Second, "Read operations took too long")
	})
}
