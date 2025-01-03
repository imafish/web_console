package db

import (
	"internal/pb"
	"time"

	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type task struct {
	ID               int64
	Status           pb.TaskStatus
	ReturnCode       int32
	Output           string
	StartTime        time.Time
	FinishTime       time.Time
	ExecutionTime    time.Duration
	WorkingDirectory string
	Commandline      string
	CreateTime       time.Time
}

func (t *task) ToProto() *pb.Task {
	pbTask := &pb.Task{
		Id:               t.ID,
		Status:           t.Status,
		ReturnCode:       t.ReturnCode,
		Output:           t.Output,
		WorkingDirectory: t.WorkingDirectory,
		Commandline:      t.Commandline,
	}

	if !t.StartTime.IsZero() {
		pbTask.StartTime = timestamppb.New(t.StartTime)
	}
	if !t.FinishTime.IsZero() {
		pbTask.FinishTime = timestamppb.New(t.FinishTime)
	}
	if t.ExecutionTime != 0 {
		pbTask.ExecutionTime = durationpb.New(t.ExecutionTime)
	}
	if !t.CreateTime.IsZero() {
		pbTask.CreateTime = timestamppb.New(t.CreateTime)
	}

	return pbTask
}

func TaskFromProto(pbTask *pb.Task) *task {
	t := &task{
		ID:               pbTask.Id,
		Status:           pbTask.Status,
		ReturnCode:       pbTask.ReturnCode,
		Output:           pbTask.Output,
		WorkingDirectory: pbTask.WorkingDirectory,
		Commandline:      pbTask.Commandline,
	}

	if pbTask.StartTime != nil {
		t.StartTime = pbTask.StartTime.AsTime()
	}
	if pbTask.FinishTime != nil {
		t.FinishTime = pbTask.FinishTime.AsTime()
	}
	if pbTask.ExecutionTime != nil {
		t.ExecutionTime = pbTask.ExecutionTime.AsDuration()
	}
	if pbTask.CreateTime != nil {
		t.CreateTime = pbTask.CreateTime.AsTime()
	}

	return t
}
