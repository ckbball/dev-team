package v1

import (
  "context"
  "database/sql"
  "errors"
  v1 "github.com/ckbball/dev-team/pkg/api/v1"
)

type repository interface {
  CreateTeam(*v1.Team) error
}

type TeamRepository struct {
  Db *sql.DB
}
