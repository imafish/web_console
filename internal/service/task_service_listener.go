package service

import (
	"internal/pb"
)

type TaskServiceListener interface {
	OnTaskCreated(task *pb.Task)
	OnTaskUpdated(task *pb.Task)
	OnTaskDeleted(task *pb.Task)
}
