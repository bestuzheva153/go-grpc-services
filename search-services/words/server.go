package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/ilyakaznacheev/cleanenv"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"log"
	"net"
	wordspb "yadro.com/course/proto/words"
	wordslogic "yadro.com/course/words/words"
)

const maxMessageSize = 4 * 1024

type Config struct {
	GRPCPort int `yaml:"grpc_port" env:"WORDS_GRPC_PORT" env-default:"28082"`
}

type server struct {
	wordspb.UnimplementedWordsServer
}

func (s *server) Ping(_ context.Context, _ *emptypb.Empty) (*emptypb.Empty, error) {
	return &emptypb.Empty{}, nil
}

func (s *server) Norm(_ context.Context, req *wordspb.WordsRequest) (*wordspb.WordsReply, error) {
	if len([]byte(req.GetPhrase())) > maxMessageSize {
		return nil, status.Error(codes.ResourceExhausted, "message exceeds 4 KiB")
	}
	normalized := wordslogic.Normalize(req.GetPhrase())
	return &wordspb.WordsReply{
		Words: normalized,
	}, nil
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
	address := fmt.Sprintf(":%d", cfg.GRPCPort)
	listener, err := net.Listen("tcp", address)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer()
	wordspb.RegisterWordsServer(s, &server{})
	reflection.Register(s)
	log.Printf("words gRPC server listening on %s", address)
	if err := s.Serve(listener); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
