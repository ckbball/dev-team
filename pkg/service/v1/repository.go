package v1

import (
  "context"
  "database/sql"
  "errors"
  v1 "github.com/ckbball/dev-team/pkg/api/v1"
)

type repository interface {
  CreateTeam(context.Context, *v1.Team) (string, error)
  DeleteTeam(context.Context, string) (int64, error)
  GetTeamByTeamId(context.Context, string) (*v1.Team, error)
  GetTeamsByUserId(context.Context, string) ([]v1.Team, error)
  AddMember(context.Context, string, string) (string, error)
  RemoveMember(context.Context, string, string) (int64, error)
  UpsertProject(context.Context, string, *v1.Project) (int64, error)
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

func (r *teamRepository) CreateTeam(ctx context.Context, team *v1.Team) (string, error) {
  // prepare sql statements for teams, skills, members
  stmt := `INSERT INTO teams (leader, name, open_roles, size, last_active)
  VALUES(?, ?, ?, ?, ?))`

  // start transaction

  // insert team into teams table capturing the id

  // insert members into members table including team_id field

  // insert skills into skills table including team_id field

  result, err := r.db.Exec(stmt, team.Leader)

}
