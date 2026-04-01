package main

import (
	"fmt"
	"net"

	"github.com/caarlos0/env/v11"
	"github.com/joho/godotenv"
	pb "github.com/kurt4ins/taskmanager/gen"
	"github.com/kurt4ins/taskmanager/internal/handlers"
	"github.com/kurt4ins/taskmanager/internal/middleware"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

type config struct {
	JWTSecret string `env:"JWT_SECRET,required"`
}

func main() {
	_ = godotenv.Load()

	var cfg config
	if err := env.Parse(&cfg); err != nil {
		panic(err)
	}

	lis, err := net.Listen("tcp", ":8081")
	if err != nil {
		panic(err)
	}

	serv := grpc.NewServer(
		grpc.UnaryInterceptor(middleware.GRPCAuthInterceptor([]byte(cfg.JWTSecret))),
		grpc.StreamInterceptor(middleware.GRPCStreamAuthInterceptor([]byte(cfg.JWTSecret))),
	)
	pb.RegisterTaskServiceServer(serv, handlers.NewTaskGRPCServer())
	reflection.Register(serv)

	fmt.Printf("grpc server starting at %s\n", lis.Addr().String())
	if err := serv.Serve(lis); err != nil {
		panic(err)
	}
}
