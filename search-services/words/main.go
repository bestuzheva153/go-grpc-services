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

const maxPhraseLen = 4096

type server struct {
	wordspb.UnimplementedWordsServer
}
type Config struct {
	Port int `yaml:"port" env:"PORT" env-default:"8080"`
}

func (s *server) Ping(_ context.Context, in *emptypb.Empty) (*emptypb.Empty, error) {
	return &emptypb.Empty{}, nil
}
func (s *server) Norm(_ context.Context, req *wordspb.WordsRequest) (*wordspb.WordsReply, error) {
	if len([]byte(req.GetPhrase())) > maxPhraseLen {
		return nil, status.Error(codes.ResourceExhausted, "message exceeds 4 KiB")
	}
	normalized := wordslogic.Norm(req.GetPhrase())
	return &wordspb.WordsReply{
		Words: normalized,
	}, nil
}
func mustLoadConfig() Config {
	var cfg Config
	var configPath string
	flag.StringVar(&configPath, "config", "config.yaml", "path to config file")
	flag.Parse()
	if configPath != "" {
		if err := cleanenv.ReadConfig(configPath, &cfg); err != nil {
			log.Fatalf("cannot read config %q: %v", configPath, err)
		}
		return cfg
	}
	if err := cleanenv.ReadEnv(&cfg); err != nil {
		log.Fatalf("cannot read env: %v", err)
	}
	return cfg
}
func main() {
	cfg := mustLoadConfig()
	address := fmt.Sprintf(":%d", cfg.Port)
	lis, err := net.Listen("tcp", address)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	grpcServer := grpc.NewServer()
	wordspb.RegisterWordsServer(grpcServer, &server{})
	reflection.Register(grpcServer)
	log.Printf("words gRPC server listening on %s", address)
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
