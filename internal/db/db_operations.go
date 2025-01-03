package db

import (
	"database/sql"
	"fmt"
	"time"

	"internal/pb"

	_ "github.com/mattn/go-sqlite3"
	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type TaskDatabase interface {
	Init() error
	Uninit() error
	GetTasks() ([]*pb.Task, error)
	GetTask(id int64) (*pb.Task, error)
	DeleteTask(id int64) error
	CreateTask(task *pb.Task) (*pb.Task, error)
	UpdateTask(task *pb.Task) (*pb.Task, error)
	GetLatestTask() (*pb.Task, error)
}

type TaskDatabaseImpl struct {
	db *sql.DB
}

type ErrNoRows struct{}

func (e *ErrNoRows) Error() string {
	return "no rows found"
}

func NewTaskDatabase(dbPath string) (TaskDatabase, error) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, err
	}
	return &TaskDatabaseImpl{db}, nil
}

func NewMockTaskDatabase() (TaskDatabase, error) {
	return nil, fmt.Errorf("not implemented")
}

const (
	SQL_CREATE_TABLE = `CREATE TABLE IF NOT EXISTS tasks (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		status INTEGER,
		commandline TEXT,
		return_code INTEGER,
		start_time DATETIME,
		finish_time DATETIME,
		execution_time INTEGER,
		working_directory TEXT,
		output TEXT,
		create_time DATETIME DEFAULT CURRENT_TIMESTAMP
	);`

	SQL_QUERY_ONE_TASK = `SELECT
		id,
		status,
		commandline,
		return_code,
		start_time,
		finish_time,
		execution_time,
		working_directory,
		create_time,
		output
	FROM tasks WHERE id = ?`

	SQL_QUERY_TASKS = `SELECT
		id,
		status,
		commandline,
		return_code,
		start_time,
		finish_time,
		execution_time,
		working_directory,
		create_time,
		output
	FROM tasks`

	SQL_UPDATE_TASK = `UPDATE tasks SET
		status = ?,
		commandline = ?,
		return_code = ?,
		start_time = ?,
		finish_time = ?,
		execution_time = ?,
		working_directory = ?,
		output = ?
	WHERE id = ?`
	SQL_DELETE_TASK = `DELETE FROM tasks WHERE id = ?`

	SQL_INSERT_TASK = `INSERT INTO tasks (status, commandline, return_code, start_time, finish_time, execution_time, working_directory, output)
	VALUES (?, ?, ?, ?, ?, ?, ?, ?)`

	SQL_QUERY_LATEST_TASK = `SELECT id, status, commandline, return_code, start_time, finish_time, execution_time, working_directory, create_time, output
	FROM tasks 
	WHERE status = ? 
	ORDER BY create_time DESC 
	LIMIT 1`
)

func (database *TaskDatabaseImpl) Init() error {
	// Create a table
	_, err := database.db.Exec(SQL_CREATE_TABLE)
	if err != nil {
		return err
	}

	return nil
}

func (database *TaskDatabaseImpl) Uninit() error {
	err := database.db.Close()
	return err
}

func (database *TaskDatabaseImpl) GetTasks() ([]*pb.Task, error) {
	// Execute the query
	rows, err := database.db.Query(SQL_QUERY_TASKS)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// Iterate over the rows
	var tasks []*pb.Task
	for rows.Next() {
		var task pb.Task
		var createTime time.Time
		var startTime time.Time
		var finishTime time.Time
		var executionTime int64
		err := rows.Scan(
			&task.Id,
			&task.Status,
			&task.Commandline,
			&task.ReturnCode,
			&startTime,
			&finishTime,
			&executionTime,
			&task.WorkingDirectory,
			&createTime,
			&task.Output,
		)
		if err != nil {
			return nil, err
		}
		task.StartTime = timestamppb.New(startTime)
		task.FinishTime = timestamppb.New(finishTime)
		task.ExecutionTime = durationpb.New(time.Duration(executionTime) * time.Second)
		task.CreateTime = timestamppb.New(createTime)
		tasks = append(tasks, &task)
	}

	// Check for errors from iterating over rows
	if err = rows.Err(); err != nil {
		return nil, err
	}

	return tasks, nil
}

func (database *TaskDatabaseImpl) GetTask(id int64) (*pb.Task, error) {
	// Task struct to hold the data
	var task pb.Task

	// Query the task
	var startTime time.Time
	var finishTime time.Time
	var executionTime int64
	var createTime time.Time
	err := database.db.QueryRow(SQL_QUERY_ONE_TASK, id).Scan(
		&task.Id,
		&task.Status,
		&task.Commandline,
		&task.ReturnCode,
		&startTime,
		&finishTime,
		&executionTime,
		&task.WorkingDirectory,
		&createTime,
		&task.Output)
	if err != nil {
		return nil, err
	}
	task.StartTime = timestamppb.New(startTime)
	task.FinishTime = timestamppb.New(finishTime)
	task.ExecutionTime = durationpb.New(time.Duration(executionTime) * time.Second)
	task.CreateTime = timestamppb.New(createTime)
	return &task, nil
}

func (database *TaskDatabaseImpl) DeleteTask(id int64) error {
	_, err := database.db.Exec(SQL_DELETE_TASK, id)
	if err != nil {
		return fmt.Errorf("DeleteTask: %v", err)
	}

	return nil
}

func (database *TaskDatabaseImpl) CreateTask(task *pb.Task) (*pb.Task, error) {
	result, err := database.db.Exec(SQL_INSERT_TASK,
		task.Status,
		task.Commandline,
		task.ReturnCode,
		task.StartTime,
		task.FinishTime,
		task.ExecutionTime,
		task.WorkingDirectory,
		task.Output,
	)
	if err != nil {
		return nil, fmt.Errorf("CreateTask: %v", err)
	}

	// Get the ID of the new task
	taskID, err := result.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("CreateTask: get last insert ID: %v", err)
	}
	task.Id = taskID

	return task, nil
}

func (database *TaskDatabaseImpl) UpdateTask(task *pb.Task) (*pb.Task, error) {
	var startTime time.Time
	var finishTime time.Time
	var executionTime int64
	if task.StartTime != nil {
		startTime = task.StartTime.AsTime()
	}
	if task.FinishTime != nil {
		finishTime = task.FinishTime.AsTime()
	}
	if task.ExecutionTime != nil {
		executionTime = int64(task.ExecutionTime.Seconds)
	}
	result, err := database.db.Exec(SQL_UPDATE_TASK,
		task.Status,
		task.Commandline,
		task.ReturnCode,
		startTime,
		finishTime,
		executionTime,
		task.WorkingDirectory,
		task.Output,
		task.Id,
	)
	if err != nil {
		return nil, fmt.Errorf("UpdateTask: %v", err)
	}

	// Check if the task was actually updated by verifying the affected row count
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return nil, fmt.Errorf("UpdateTask: get rows affected: %v", err)
	}
	if rowsAffected == 0 {
		return nil, fmt.Errorf("UpdateTask: no task with ID %d found", task.Id)
	}

	return task, nil
}

func (database *TaskDatabaseImpl) GetLatestTask() (*pb.Task, error) {
	var task pb.Task
	var createTime time.Time
	err := database.db.QueryRow(SQL_QUERY_LATEST_TASK, pb.TaskStatus_NEW).Scan(
		&task.Id,
		&task.Status,
		&task.Commandline,
		&task.ReturnCode,
		&task.StartTime,
		&task.FinishTime,
		&task.ExecutionTime,
		&task.WorkingDirectory,
		&createTime,
		&task.Output,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, &ErrNoRows{} // No task found
		}
		return nil, fmt.Errorf("GetLatestTask: %v", err)
	}

	task.CreateTime = timestamppb.New(createTime)
	return &task, nil
}
