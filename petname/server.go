package main

import (
	"context"
	"flag"
	"fmt"
	petname "github.com/dustinkirkland/golang-petname"
	"github.com/ilyakaznacheev/cleanenv"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"log"
	"net"
	petnamepb "yadro.com/course/proto"
)

type Config struct {
	GRPCPort int `yaml:"grpc_port" env:"PETNAME_GRPC_PORT" env-default:"28081"`
}

type server struct {
	petnamepb.UnimplementedPetnameGeneratorServer
}

func (s *server) Ping(_ context.Context, in *emptypb.Empty) (*emptypb.Empty, error) {
	return &emptypb.Empty{}, nil
}
func (s *server) Generate(ctx context.Context, req *petnamepb.PetnameRequest) (*petnamepb.PetnameResponse, error) {
	if req.GetWords() <= 0 {
		return nil, status.Error(codes.InvalidArgument, "words must be > 0")
	}
	name := petname.Generate(int(req.GetWords()), req.GetSeparator())
	return &petnamepb.PetnameResponse{Name: name}, nil
}
func (s *server) GenerateMany(req *petnamepb.PetnameStreamRequest, stream grpc.ServerStreamingServer[petnamepb.PetnameResponse]) error {
	if req.GetWords() <= 0 || req.GetNames() <= 0 {
		return status.Error(codes.InvalidArgument, "words and names must be > 0")
	}
	for i := 0; i < int(req.GetNames()); i++ {
		name := petname.Generate(int(req.GetWords()), req.GetSeparator())
		err := stream.Send(&petnamepb.PetnameResponse{Name: name})
		if err != nil {
			return err
		}
	}
	return nil
}
func loadConfig() Config {
	var cfg Config
	var configPath string
	flag.StringVar(&configPath, "config", "", "path to config file")
	flag.Parse()
	if configPath != "" {
		if err := cleanenv.ReadConfig(configPath, &cfg); err != nil {
			log.Fatalf("read config: %v", err)
		}
		return cfg
	}
	if err := cleanenv.ReadEnv(&cfg); err != nil {
		log.Fatalf("read env: %v", err)
	}
	return cfg
}

func main() {
	cfg := loadConfig()
	addr := fmt.Sprintf(":%d", cfg.GRPCPort)
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatalf("listen %s: %v", addr, err)
	}
	grpcServer := grpc.NewServer()
	petnamepb.RegisterPetnameGeneratorServer(grpcServer, &server{})
	reflection.Register(grpcServer)
	log.Printf("petname gRPC server listening on %s", addr)
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("serve: %v", err)
	}
}
