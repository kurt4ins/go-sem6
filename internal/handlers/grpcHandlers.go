package handlers

import (
	"context"
	"time"

	"github.com/google/uuid"
	pb "github.com/kurt4ins/taskmanager/gen"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type TaskGRPCServer struct {
	pb.UnimplementedTaskServiceServer
}

func NewTaskGRPCServer() *TaskGRPCServer {
	return &TaskGRPCServer{}
}

func (s *TaskGRPCServer) CreateTask(ctx context.Context, r *pb.CreateTaskReq) (*pb.Task, error) {
	if r.Name == "" {
		return nil, status.Error(codes.InvalidArgument, "name is required")
	}

	task := &pb.Task{
		Id:          uuid.NewString(),
		Name:        r.Name,
		Description: r.Description,
		Completed:   false,
	}

	return task, nil
}

func (s *TaskGRPCServer) ListTasks(r *pb.ListTasksReq, stream pb.TaskService_ListTasksServer) error {
	tasks := []*pb.Task{
		{Id: uuid.NewString(), Name: "task 1", Description: "desc 1", Completed: true},
		{Id: uuid.NewString(), Name: "task 2", Description: "desc 2", Completed: false},
	}

	for _, task := range tasks {
		if err := stream.Send(task); err != nil {
			return err
		}
		time.Sleep(1 * time.Second)
	}

	return nil
}
