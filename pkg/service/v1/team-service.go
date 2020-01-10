package v1

import (
  "context"
  "errors"
  "fmt"
  "log"
  "strconv"
  "time"

  //"github.com/golang/protobuf/ptypes"
  "encoding/json"
  "github.com/ThreeDotsLabs/watermill"
  "github.com/ThreeDotsLabs/watermill/message"
  // "github.com/go-redis/cache/v7"
  "google.golang.org/grpc"
  "google.golang.org/grpc/codes"
  "google.golang.org/grpc/status"

  v1 "github.com/ckbball/dev-team/pkg/api/v1"
)

const (
  apiVersion = "v1"
  eventName  = "team_created"
)

type handler struct {
  repo       repository
  subscriber message.Subscriber
  publisher  message.Publisher
}

func NewTeamServiceServer(repo repository,
  subscriber message.Subscriber, publisher message.Publisher) *handler {
  return &handler{
    repo:       repo,
    subscriber: subscriber,
    publisher:  publisher,
  }
}

func (s *handler) checkAPI(api string) error {
  if len(api) > 0 {
    if apiVersion != api {
      return status.Errorf(codes.Unimplemented,
        "unsupported API version: service implements API version '%s', but asked for '%s'", apiVersion, api)
    }
  }
  return nil
}

/* Team handles api calls to grpc method Team and REST endpoint: /v1/Team
any error generated or nil if no errors.
*/
func (s *handler) Team(ctx context.Context, req *v1.Request) (*v1.Response, error) {
  // check api version
  if err := s.checkAPI(req.Api); err != nil {
    return nil, err
  }

}
