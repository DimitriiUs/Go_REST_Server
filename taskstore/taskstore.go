package taskstore

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// TaskDB is a database of tasks.
type TaskDB struct {
	pool *pgxpool.Pool
	err  error
}

func New() *TaskDB {
	tdb := &TaskDB{}
	dbURL := "postgres://postgres:root@localhost:5432/postgres"

	tdb.pool, tdb.err = pgxpool.New(context.Background(), dbURL)
	if tdb.err != nil {
		log.Fatalf("Unable to connection to database: %v\n", tdb.err)
	}

	log.Println("Connected!")

	return tdb
}

func (tdb *TaskDB) Close() {
	tdb.pool.Close()
}

type Task struct {
	Id   int       `json:"id"`
	Text string    `json:"text"`
	Due  time.Time `json:"due"`
}

func (tdb *TaskDB) CreateTask(text string, due time.Time) int {

	row := tdb.pool.QueryRow(context.Background(), "INSERT INTO tasks (task_description, due_date) VALUES ($1, $2) RETURNING task_id", text, due.Format(time.DateTime))
	var count int
	if err := row.Scan(&count); err != nil {
		log.Fatal(err)

	}

	return count
}

// GetTask retrieves a task from the store, by id. If no such id exists, an
// error is returned.
func (tdb *TaskDB) GetTask(id int) (Task, error) {
	task := Task{}
	row := tdb.pool.QueryRow(context.Background(), "SELECT * FROM tasks WHERE task_id = $1", id)
	if err := row.Scan(&task.Id, &task.Text, &task.Due); err != nil {
		return Task{}, err
	}

	return task, nil
}

// DeleteTask deletes the task with the given id. If no such id exists, an error
// is returned.
func (tdb *TaskDB) DeleteTask(id int) (string, error) {
	row := tdb.pool.QueryRow(context.Background(), "DELETE FROM tasks WHERE task_id = $1 RETURNING task_description", id)
	var description string
	if err := row.Scan(&description); err != nil {
		return "", err
	}

	return fmt.Sprintf("Task: `%s` was deleted", description), nil
}

// DeleteAllTasks deletes all tasks in the store.
func (tdb *TaskDB) DeleteAllTasks() (string, error) {
	res, err := tdb.pool.Exec(context.Background(), "TRUNCATE TABLE tasks")
	if err != nil {
		return "", err
	}

	if res.RowsAffected() == 0 {
		return "", errors.New("No tasks were deleted")
	}
	return fmt.Sprintf("Deleted %d tasks", res.RowsAffected()), nil
}

// GetAllTasks returns all the tasks in the store, in arbitrary order.
func (tdb *TaskDB) GetAllTasks() ([]Task, error) {
	rows, err := tdb.pool.Query(context.Background(), "SELECT * FROM tasks")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var allTasks []Task
	for rows.Next() {
		task := Task{}
		if err := rows.Scan(&task.Id, &task.Text, &task.Due); err != nil {
			return nil, err
		}
		allTasks = append(allTasks, task)
	}
	return allTasks, nil
}

// GetTasksByDueDate returns all the tasks that have the given due date, in
// arbitrary order.
func (tdb *TaskDB) GetTasksByDueDate(year int, month time.Month, day int) ([]Task, error) {
	dueDate := time.Date(year, month, day, 0, 0, 0, 0, time.UTC)
	rows, err := tdb.pool.Query(context.Background(), "SELECT * FROM tasks WHERE due_date = $1", dueDate)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tasks []Task
	for rows.Next() {
		task := Task{}
		if err := rows.Scan(&task.Id, &task.Text, &task.Due); err != nil {
			return nil, err
		}
		tasks = append(tasks, task)
	}

	if tasks == nil {
		return nil, errors.New("No tasks were found")
	}
	return tasks, nil
}
