package v1

import (
  "context"
  "database/sql"
  "errors"
  v1 "github.com/ckbball/dev-team/pkg/api/v1"
)

type repository interface {
  CreateTeam(*v1.Team) (string, error)
  DeleteTeam(string) (int64, error)
  AddMember(string, string) (string, error)
  RemoveMember(string, string) (int64, error)
  UpsertProject(string, *v1.Project) (int64, error)
}

type TeamRepository struct {
  Db *sql.DB
}
