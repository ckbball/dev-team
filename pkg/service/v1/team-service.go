package v1

import (
  "context"
  "errors"
  "fmt"
  // "log"
  // "strconv"
  // "time"
  "os"

  //"github.com/golang/protobuf/ptypes"
  // "encoding/json"
  // "github.com/ThreeDotsLabs/watermill"
  //"github.com/ThreeDotsLabs/watermill/message"
  // "github.com/go-redis/cache/v7"
  "google.golang.org/grpc/codes"
  "google.golang.org/grpc/status"

  v1 "github.com/ckbball/dev-team/pkg/api/v1"
)

const (
  apiVersion = "v1"
  eventName  = "team_created"
)

type handler struct {
  repo        repository
  userSvcAddr string
  //subscriber message.Subscriber
  //publisher  message.Publisher
}

func NewTeamServiceServer(repo repository, user string) *handler {
  return &handler{
    repo:        repo,
    userSvcAddr: user,
    //subscriber: subscriber,
    //publisher:  publisher,
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
  // need to make sure team_name is unique
  teamTemp, err := s.repo.GetTeamByTeamName(ctx, req.Team.Name)
  fmt.Fprintf(os.Stderr, "Error: from Repo GetTeamByTeamName: %v\n", err)
  fmt.Fprintf(os.Stderr, "Team: from Repo GetTeamByTeamName: %v\n", teamTemp)
  if err != nil && err.Error() != "team Query: no matching record found" {
    return nil, err
  } else if teamTemp != nil {
    return &v1.TeamUpsertResponse{
      Api:    "v1",
      Status: "error:duplicatename",
    }, nil
  }
  // team name is now unique

  // need to make sure auth token user has less than 5 teams
  // get number of teams user owns
  count, err := s.repo.CountUserTeams(ctx, req.UserId)
  if err != nil || count == -1 {
    return nil, err
  }
  // if >= 5 deny request
  if count >= 5 {
    return &v1.TeamUpsertResponse{
      Api:    "v1",
      Status: "error:maxteamcount",
    }, nil
  }

  // else continue

  // call repo func to create a new team
  fmt.Fprintf(os.Stderr, "State Team: about to call CreateTeam: %v\n", teamTemp)
  newId, err := s.repo.CreateTeam(ctx, req.Team)
  if err != nil {
    fmt.Fprintf(os.Stderr, "error from Repo CreateTeam: %v\n", newId)
    return nil, err
  }

  // publish team_created Event here

  // return successful response
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

  // Check if user owns team correlating to req.TeamId, using req.UserId
  owns, err := s.repo.CheckUserOwnsTeam(ctx, req.UserId, req.TeamId)
  if err != nil {
    fmt.Fprintf(os.Stderr, "error from Repo CheckIfUserOwnsTeam: %v\n", req.TeamId)
    return nil, err
  }
  if !owns {
    fmt.Fprintf(os.Stderr, "user %v doesn't own team: %v\n", req.UserId, req.TeamId)
    return nil, errors.New("invalid")
  }

  // delete the team corresponding to TeamId
  teamRows, memRows, skillRows, err := s.repo.DeleteTeam(ctx, req.TeamId)
  if err != nil {
    fmt.Fprintf(os.Stderr, "error from Repo DeleteTeam: %v\n", req.TeamId)
    return nil, err
  }

  // publish team_deleted Event here

  return &v1.TeamDeleteResponse{
    Api:     "v1",
    Status:  "Deleted",
    Teams:   teamRows,
    Skills:  skillRows,
    Members: memRows,
    Id:      req.TeamId,
  }, nil
}
func (s *handler) AddMember(ctx context.Context, req *v1.MemberUpsertRequest) (*v1.MemberUpsertResponse, error) {
  // check api version
  if err := s.checkAPI(req.Api); err != nil {
    return nil, err
  }

  // Check if team is at max size
  max, err := s.repo.CheckTeamSize(ctx, req.TeamId)
  if err != nil {
    fmt.Fprintf(os.Stderr, "Error AddMember: Repo CheckIfUserOwnsTeam: %v\n", req.TeamId)
    return nil, err
  }
  if max {
    fmt.Fprintf(os.Stderr, "max team size: %v\n", req.UserId, req.TeamId)
    return &v1.MemberUpsertResponse{
      Api:    "v1",
      Status: "error:maxmembercount",
    }, nil
  }

  // Check if user owns team correlating to req.TeamId, using req.UserId
  owns, err := s.repo.CheckUserOwnsTeam(ctx, req.UserId, req.TeamId)
  if err != nil {
    fmt.Fprintf(os.Stderr, "Error AddMember: Repo CheckIfUserOwnsTeam: %v\n", req.TeamId)
    return nil, err
  }
  if !owns {
    fmt.Fprintf(os.Stderr, "user %v doesn't own team: %v\n", req.UserId, req.TeamId)
    return nil, errors.New("invalid")
  }

  // need to check if trying to add duplicate user
  // does member_id exist in members table where team_id == req.TeamId
  exists, err := s.repo.CheckMemberExists(ctx, req.MemberId, req.TeamId)
  if err != nil {
    fmt.Fprintf(os.Stderr, "Error AddMember: Repo CheckMemberExists: %v\n", req.TeamId)
    return nil, err
  }
  if exists {
    fmt.Fprintf(os.Stderr, "user %v exists on team: %v\n", req.UserId, req.TeamId)
    return &v1.MemberUpsertResponse{
      Api:    "v1",
      Status: "error:exists",
    }, nil
  }

  // need to grab user's id by email because team owner wont know user's id fullfil this in edge?

  newId, err := s.repo.AddMember(ctx, req)
  if err != nil {
    fmt.Fprintf(os.Stderr, "error from Repo AddMember: %v\n", req.TeamId)
    return nil, err
  }

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

  // check req.UserId owns team=req.TeamId

  count, err := s.repo.RemoveMember(ctx, req.TeamId, req.MemberNumber)
  if err != nil {
    fmt.Fprintf(os.Stderr, "error from Repo RemoveMember: %v\n", req.TeamId)
    return nil, err
  }

  // publish MemberRemoved Event here

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

  _, err := s.repo.UpsertProject(ctx, req.TeamId, req.Project)
  if err != nil {
    fmt.Fprintf(os.Stderr, "error from Repo UpsertProject: %v\n", req.TeamId)
    return nil, err
  }

  // publish team_created Event here

  return &v1.ProjectUpsertResponse{
    Api:    "v1",
    Status: "Project Upserted",
  }, nil
}

// change to team name
func (s *handler) GetTeamByTeamName(ctx context.Context, req *v1.GetByTeamNameRequest) (*v1.GetByTeamNameResponse, error) {
  // check api version
  if err := s.checkAPI(req.Api); err != nil {
    return nil, err
  }

  /*
     team, err := s.repo.GetTeamByTeamName(ctx, req.Name)
     if err != nil {
       fmt.Fprintf(os.Stderr, "error from Repo GetByTeamName: %v\n", req.Name)
       return nil, err
     }
  */

  return &v1.GetByTeamNameResponse{
    Api:    "v1",
    Status: "deprecated",
  }, nil
}

// needs to be reworked
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

  if len(teams) == 0 {
    return &v1.GetByUserIdResponse{
      Api:    "v1",
      Status: "empty",
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

// Fetches user's teams by accessing valid jwt token sent in the headers,
func (s *handler) GetTeamsByCurrentUser(ctx context.Context, req *v1.GetByUserIdRequest) (*v1.GetByUserIdResponse, error) {
  // check api version
  if err := s.checkAPI(req.Api); err != nil {
    return nil, err
  }

  // call repo method to get teams sending it id you get back from token
  teams, err := s.repo.GetTeamsByUserId(ctx, req.Id)
  if err != nil {
    // if error occured accessing db return it here
    return nil, err
  }
  if len(teams) == 0 {
    return &v1.GetByUserIdResponse{
      Api:    apiVersion,
      Status: "empty",
      Teams:  teams,
      Id:     req.Id,
    }, nil
  }

  // return response struct
  return &v1.GetByUserIdResponse{
    Api:    apiVersion,
    Status: "teams",
    Teams:  teams,
    // maybe in future add more data to response about the added user.
  }, nil
}

func (s *handler) GetTeams(ctx context.Context, req *v1.GetTeamsRequest) (*v1.GetTeamsResponse, error) {
  // check api version
  if err := s.checkAPI(req.Api); err != nil {
    return nil, err
  }

  fmt.Fprintf(os.Stderr, "\npage: %v\nlimit: %v\n", req.Page, req.Limit)

  teams, err := s.repo.GetTeams(ctx, req)
  if err != nil {
    fmt.Fprintf(os.Stderr, "error from Repo GetTeams:\n")
    return nil, err
  }

  if len(teams) == 0 {
    return &v1.GetTeamsResponse{
      Api:    "v1",
      Status: "empty",
      Teams:  teams,
    }, nil
  }

  return &v1.GetTeamsResponse{
    Api:    "v1",
    Status: "teams",
    Teams:  teams,
  }, nil
}
