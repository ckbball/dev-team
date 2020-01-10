package main

import (
  "context"
  "flag"
  "log"
  "time"

  //"github.com/golang/protobuf/ptypes"
  "google.golang.org/grpc"

  v1 "github.com/ckbball/dev-team/pkg/api/v1"
)

const (
  // apiVersion is version of API is provided by server
  apiVersion = "v1"
)

func main() {
  // get configuration
  address := flag.String("server", "", "gRPC server in format host:port")
  flag.Parse()

  // Set up a connection to the server.
  conn, err := grpc.Dial(*address, grpc.WithInsecure())
  if err != nil {
    log.Fatalf("did not connect: %v", err)
  }
  defer conn.Close()

  c := v1.NewTeamServiceClient(conn)

  ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
  defer cancel()

}
