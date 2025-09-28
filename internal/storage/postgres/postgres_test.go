package postgres

import (
	"TaskService/internal/model"
	"context"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
)

func TestPostgresStorage_Get(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "sqlmock")

	storage := &repo{db: sqlxDB}

	ctx := context.Background()
	taskID := 1
	expectedTask := model.Task{
		ID:          taskID,
		Title:       "Test Task",
		Description: "Test Description",
		Status:      "created",
	}

	rows := sqlmock.NewRows([]string{"id", "title", "description", "status"}).
		AddRow(expectedTask.ID, expectedTask.Title, expectedTask.Description, expectedTask.Status)

	mock.ExpectQuery("SELECT \\* FROM tasks WHERE id = \\$1").
		WithArgs(taskID).
		WillReturnRows(rows)

	result, err := storage.Get(ctx, taskID)

	assert.NoError(t, err)
	assert.Equal(t, expectedTask, result)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestPostgresStorage_GetList(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "sqlmock")
	storage := &repo{db: sqlxDB}

	ctx := context.Background()
	expectedTasks := []model.Task{
		{ID: 1, Title: "Task 1", Description: "Desc 1", Status: "created"},
		{ID: 2, Title: "Task 2", Description: "Desc 2", Status: "done"},
	}

	rows := sqlmock.NewRows([]string{"id", "title", "description", "status"}).
		AddRow(expectedTasks[0].ID, expectedTasks[0].Title, expectedTasks[0].Description, expectedTasks[0].Status).
		AddRow(expectedTasks[1].ID, expectedTasks[1].Title, expectedTasks[1].Description, expectedTasks[1].Status)

	mock.ExpectQuery("SELECT \\* FROM tasks").WillReturnRows(rows)

	result, err := storage.GetList(ctx)

	assert.NoError(t, err)
	assert.Equal(t, expectedTasks, result)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestPostgresStorage_Create(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "sqlmock")
	storage := &repo{db: sqlxDB}

	ctx := context.Background()
	task := model.Task{
		Title:       "New Task",
		Description: "New Description",
	}
	expectedID := 1

	mock.ExpectBegin()

	tx, err := storage.BeginTx(ctx)
	assert.NoError(t, err)

	mock.ExpectQuery("INSERT INTO tasks \\(title, description\\) VALUES \\(\\$1, \\$2\\) RETURNING id").
		WithArgs(task.Title, task.Description).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(expectedID))

	mock.ExpectCommit()

	resultID, err := storage.Create(ctx, tx, task)
	assert.NoError(t, err)
	assert.Equal(t, expectedID, resultID)

	err = tx.Commit()
	assert.NoError(t, err)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestPostgresStorage_Update(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "sqlmock")
	storage := &repo{db: sqlxDB}

	ctx := context.Background()
	task := model.Task{
		ID:          1,
		Title:       "Updated Task",
		Description: "Updated Description",
		Status:      "done",
	}

	mock.ExpectBegin()

	tx, err := storage.BeginTx(ctx)
	assert.NoError(t, err)

	mock.ExpectExec("UPDATE tasks SET title = \\$1, description = \\$2, status = \\$3 WHERE id = \\$4").
		WithArgs(task.Title, task.Description, task.Status, task.ID).
		WillReturnResult(sqlmock.NewResult(1, 1))

	mock.ExpectCommit()

	err = storage.Update(ctx, tx, task)
	assert.NoError(t, err)

	err = tx.Commit()
	assert.NoError(t, err)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestPostgresStorage_BeginTx(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "sqlmock")
	storage := &repo{db: sqlxDB}

	ctx := context.Background()

	mock.ExpectBegin()

	tx, err := storage.BeginTx(ctx)
	assert.NoError(t, err)
	assert.NotNil(t, tx)

	mock.ExpectRollback()

	err = tx.Rollback()
	assert.NoError(t, err)

	assert.NoError(t, mock.ExpectationsWereMet())
}
