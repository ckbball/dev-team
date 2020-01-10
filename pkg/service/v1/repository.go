package v1

import (
  "context"
  "database/sql"
  "errors"
)

type repository interface {
  CreateTeam(event interface{}) error
}

type TeamRepository struct {
  Db *sql.DB
}
