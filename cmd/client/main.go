// usage: go run cmd/client/main.go --token token create/list

package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"os"

	pb "github.com/kurt4ins/taskmanager/gen"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
)

func main() {
	conn, err := grpc.NewClient(
		"localhost:8081",
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	client := pb.NewTaskServiceClient(conn)

	token := flag.String("token", "", "Bearer JWT token")
	flag.Parse()
	ctx := metadata.AppendToOutgoingContext(context.Background(), "authorization", fmt.Sprintf("Bearer %s", *token))

	args := flag.Args()
	if len(args) < 1 {
		fmt.Println("no arguments provided")
		os.Exit(1)
	}

	switch args[0] {
	case "create":
		task, err := client.CreateTask(ctx, &pb.CreateTaskReq{
			Name:        "grpc task",
			Description: "mock data",
		})
		if err != nil {
			fmt.Println("CreateTask err:", err)
			os.Exit(1)
		}
		fmt.Printf("task '%s' (id: %s) created\n", task.Name, task.Id)

	case "list":
		stream, err := client.ListTasks(ctx, &pb.ListTasksReq{})
		if err != nil {
			fmt.Println("ListTasks err:", err)
			os.Exit(1)
		}

		for {
			task, err := stream.Recv()
			if err == io.EOF {
				break
			}

			if err != nil {
				fmt.Println("stream err:", err)
				os.Exit(1)
			}

			fmt.Printf("task: '%s' (id: %s)\ndescription: '%s'\n\n", task.Name, task.Id, task.Description)
		}
	}
}
