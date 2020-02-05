package grpc

import (
  "context"
  "log"
  "net"
  "os"
  "os/signal"

  "google.golang.org/grpc"

  v1 "github.com/ckbball/dev-team/pkg/api/v1"
  "github.com/ckbball/dev-team/pkg/logger"
  "github.com/ckbball/dev-team/pkg/protocol/grpc/middleware"
)

// RunServer runs gRPC service to publish Team service
func RunServer(ctx context.Context, v1API v1.TeamServiceServer, port string) error {
  listen, err := net.Listen("tcp", ":"+port)
  if err != nil {
    return err
  }

  opts := []grpc.ServerOption{}

  opts = middleware.AddLogging(logger.Log, opts)

  // register service
  server := grpc.NewServer(opts...)
  v1.RegisterTeamServiceServer(server, v1API)

  // graceful shutdown
  c := make(chan os.Signal, 1)
  signal.Notify(c, os.Interrupt)
  go func() {
    for range c {
      // sig is a ^C, handle it
      log.Println("shutting down gRPC server...")

      server.GracefulStop()

      <-ctx.Done()
    }
  }()

  // start gRPC server
  log.Println("starting gRPC server...")
  return server.Serve(listen)
}
