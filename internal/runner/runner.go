package runner

import (
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"internal/pb"

	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func Run(task *pb.Task) (<-chan *pb.Task, error) {
	cmd := exec.Command("sh", "-c", task.Commandline)
	cmd.Dir = task.WorkingDirectory

	outputPathAbsolute := task.GetOutput()
	if !filepath.IsAbs(task.Output) {
		outputPathAbsolute = filepath.Join(task.GetWorkingDirectory(), task.Output)
	}
	outputFile, err := os.Create(outputPathAbsolute)
	if err != nil {
		return nil, err
	}
	defer outputFile.Close()

	cmd.Stdout = outputFile
	cmd.Stderr = outputFile

	startTime := time.Now()
	err = cmd.Start()
	if err != nil {
		return nil, err
	}

	ch := make(chan *pb.Task)

	go func() {
		defer close(ch)
		err := cmd.Wait()
		if err != nil {
			ch <- nil
			return
		}
		finishTime := time.Now()
		task.FinishTime = timestamppb.New(finishTime)
		task.ReturnCode = int32(cmd.ProcessState.ExitCode())
		task.ExecutionTime = durationpb.New(finishTime.Sub(startTime))
		ch <- task
	}()

	return ch, nil
}
