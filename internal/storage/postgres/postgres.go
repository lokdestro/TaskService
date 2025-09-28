package postgres

import (
	"TaskService/internal/model"
	"TaskService/pkg/logger"
	"context"
	"database/sql"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

type Storage interface {
	Get(ctx context.Context, id int) (model.Task, error)
	GetList(ctx context.Context) ([]model.Task, error)
	Update(ctx context.Context, tx Tx, req model.Task) error
	Create(ctx context.Context, tx Tx, task model.Task) (int, error)
	BeginTx(ctx context.Context) (Tx, error)
}

type Tx interface {
	Commit() error
	Rollback() error
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
	QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row
}

type repo struct {
	db *sqlx.DB
}

func New(cfg Config) (Storage, error) {
	db, err := sqlx.Connect(cfg.Driver, cfg.URL)
	if err != nil {
		return nil, err
	}

	result := &repo{
		db: db,
	}

	log := logger.Get()

	log.Info().Msg("connect to postgres")

	return result, nil
}

func (r *repo) Get(ctx context.Context, id int) (model.Task, error) {
	var task model.Task

	query := "SELECT * FROM tasks WHERE id = $1"

	err := r.db.GetContext(ctx, &task, query, id)

	return task, err
}

func (r *repo) GetList(ctx context.Context) ([]model.Task, error) {
	var tasks []model.Task

	query := "SELECT * FROM tasks"

	rows, err := r.db.QueryxContext(ctx, query)
	if err != nil {
		return tasks, err
	}

	for rows.Next() {
		var task model.Task

		err = rows.StructScan(&task)
		if err != nil {
			return tasks, err
		}

		tasks = append(tasks, task)
	}

	return tasks, nil
}

func (r *repo) Update(ctx context.Context, tx Tx, req model.Task) error {
	query := "UPDATE tasks SET title = $1, description = $2, status = $3 WHERE id = $4"

	_, err := tx.ExecContext(ctx, query, req.Title, req.Description, req.Status, req.ID)

	return err
}

func (r *repo) Create(ctx context.Context, tx Tx, req model.Task) (int, error) {
	var id int

	query := "INSERT INTO tasks (title, description) VALUES ($1, $2) RETURNING id"

	err := tx.QueryRowContext(ctx, query, req.Title, req.Description).Scan(&id)
	if err != nil {
		return 0, err
	}

	return id, nil
}

func (r *repo) BeginTx(ctx context.Context) (Tx, error) {
	return r.db.BeginTxx(ctx, nil)
}
