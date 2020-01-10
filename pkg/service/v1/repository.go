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

type teamRepository struct {
  db *sql.DB
}

func NewTeamRepository(db *sql.DB) *teamRepository {
  return &teamRepository{
    db: db,
  }
}

func (r *teamRepository) connect(ctx context.Context) (*sql.Conn, error) {
  c, err := r.db.Conn(ctx)
  if err != nil {
    return nil, status.Error(codes.Unknown, "failed to connect to database-> "+err.Error())
  }
  return c, nil
}
