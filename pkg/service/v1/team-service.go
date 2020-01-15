package v1

import (
  "context"
  //"errors"
  "fmt"
  // "log"
  // "strconv"
  // "time"
  "os"

  //"github.com/golang/protobuf/ptypes"
  // "encoding/json"
  // "github.com/ThreeDotsLabs/watermill"
  "github.com/ThreeDotsLabs/watermill/message"
  // "github.com/go-redis/cache/v7"
  // "google.golang.org/grpc"
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
func (s *handler) CreateTeam(ctx context.Context, req *v1.TeamUpsertRequest) (*v1.TeamUpsertResponse, error) {
  // check api version
  if err := s.checkAPI(req.Api); err != nil {
    return nil, err
  }

  fmt.Fprintf(os.Stderr, "team from request: %v\n", req.Team)

  newId, err := s.repo.CreateTeam(ctx, req.Team)
  if err != nil {
    fmt.Fprintf(os.Stderr, "error from Repo CreateTeam: %v\n", newId)
    return nil, err
  }
  fmt.Fprintf(os.Stderr, "Does repo work?\n")

  // publish team_created Event here

  return &v1.TeamUpsertResponse{
    Api:    "v1",
    Status: "Upserted",
    Id:     newId,
  }, nil

}

func (s *handler) DeleteTeam(ctx context.Context, req *v1.TeamDeleteRequest) (*v1.TeamDeleteResponse, error) {
  // check api version
  if err := s.checkAPI(req.Api); err != nil {
    return nil, err
  }

  fmt.Fprintf(os.Stderr, "id from request: %v\n", req.Id)

  teamRows, memRows, skillRows, err := s.repo.DeleteTeam(ctx, req.Id)
  if err != nil {
    fmt.Fprintf(os.Stderr, "error from Repo DeleteTeam: %v\n", req.Id)
    return nil, err
  }
  fmt.Fprintf(os.Stderr, "Does repo work?\n")

  // publish team_created Event here

  return &v1.TeamDeleteResponse{
    Api:     "v1",
    Status:  "Deleted",
    Teams:   teamRows,
    Skills:  skillRows,
    Members: memRows,
    Id:      req.Id,
  }, nil
}
func (s *handler) AddMember(ctx context.Context, req *v1.MemberUpsertRequest) (*v1.MemberUpsertResponse, error) {
  // check api version
  if err := s.checkAPI(req.Api); err != nil {
    return nil, err
  }

  // add in here somewhere maybe in future to get new member's name from their account as an additional
  // field

  newId, err := s.repo.AddMember(ctx, req)
  if err != nil {
    fmt.Fprintf(os.Stderr, "error from Repo AddMember: %v\n", req.Id)
    return nil, err
  }
  fmt.Fprintf(os.Stderr, "Does repo work?\n")

  // publish team_created Event here

  return &v1.MemberUpsertResponse{
    Api:          "v1",
    Status:       "Upserted",
    MemberNumber: newId,
  }, nil
}

// untested
func (s *handler) RemoveMember(ctx context.Context, req *v1.MemberDeleteRequest) (*v1.MemberDeleteResponse, error) {
  // check api version
  if err := s.checkAPI(req.Api); err != nil {
    return nil, err
  }

  count, err := s.repo.RemoveMember(ctx, req.Id, req.MemberNumber)
  if err != nil {
    fmt.Fprintf(os.Stderr, "error from Repo RemoveMember: %v\n", req.Id)
    return nil, err
  }
  fmt.Fprintf(os.Stderr, "Does repo work?\n")

  // publish team_created Event here

  return &v1.MemberDeleteResponse{
    Api:    "v1",
    Status: "Member Deleted",
    Count:  count,
  }, nil
}

func (s *handler) UpsertTeamProject(ctx context.Context, req *v1.ProjectUpsertRequest) (*v1.ProjectUpsertResponse, error) {
  // check api version
  if err := s.checkAPI(req.Api); err != nil {
    return nil, err
  }

  // add in here somewhere maybe in future to get new member's name from their account as an additional
  // field

  _, err := s.repo.UpsertProject(ctx, req.Id, req.Project)
  if err != nil {
    fmt.Fprintf(os.Stderr, "error from Repo UpsertProject: %v\n", req.Id)
    return nil, err
  }
  fmt.Fprintf(os.Stderr, "Does repo work?\n")

  // publish team_created Event here

  return &v1.ProjectUpsertResponse{
    Api:    "v1",
    Status: "Project Upserted",
  }, nil
}

func (s *handler) GetTeamByTeamId(ctx context.Context, req *v1.GetByTeamIdRequest) (*v1.GetByTeamIdResponse, error) {
  // check api version
  if err := s.checkAPI(req.Api); err != nil {
    return nil, err
  }

  team, err := s.repo.GetTeamByTeamId(ctx, req.Id)
  if err != nil {
    fmt.Fprintf(os.Stderr, "error from Repo GetByTeamId: %v\n", req.Id)
    return nil, err
  }
  fmt.Fprintf(os.Stderr, "Does repo work?\n")

  return &v1.GetByTeamIdResponse{
    Api:    "v1",
    Status: "Team Retrieved",
    Team:   team,
    Id:     team.Id,
  }, nil
}

func (s *handler) GetTeamsByUserId(ctx context.Context, req *v1.GetByUserIdRequest) (*v1.GetByUserIdResponse, error) {
  // check api version
  if err := s.checkAPI(req.Api); err != nil {
    return nil, err
  }

  teams, err := s.repo.GetTeamsByUserId(ctx, req.Id)
  if err != nil {
    fmt.Fprintf(os.Stderr, "error from Repo GetByUserId: %v\n", req.Id)
    return nil, err
  }

  fmt.Fprintf(os.Stderr, "checking teams on request where user has no teams: %v\n", teams)

  if len(teams) == 0 {
    return &v1.GetByUserIdResponse{
      Api:    "v1",
      Status: "User has no teams",
      Teams:  teams,
      Id:     req.Id,
    }, nil
  }

  return &v1.GetByUserIdResponse{
    Api:    "v1",
    Status: "Teams retrieived",
    Teams:  teams,
    Id:     req.Id,
  }, nil
}

func (s *handler) GetTeams(ctx context.Context, req *v1.GetTeamsRequest) (*v1.GetTeamsResponse, error) {
  // check api version
  if err := s.checkAPI(req.Api); err != nil {
    return nil, err
  }

  teams, err := s.repo.GetTeams(ctx, req.Page, req.Limit)
  if err != nil {
    fmt.Fprintf(os.Stderr, "error from Repo GetByUserId: %v\n", req.Id)
    return nil, err
  }

  if len(teams) == 0 {
    return &v1.GetTeamsResponse{
      Api:    "v1",
      Status: "No teams found",
      Teams:  teams,
    }, nil
  }

  return &v1.GetTeamsResponse{
    Api:    "v1",
    Status: "Teams retrieived",
    Teams:  teams,
  }, nil
}
