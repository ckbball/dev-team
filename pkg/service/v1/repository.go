package v1

import (
  "context"
  "database/sql"
  //"errors"
  "fmt"
  "os"
  "strconv"
  "strings"

  v1 "github.com/ckbball/dev-team/pkg/api/v1"
)

type repository interface {
  CreateTeam(context.Context, *v1.Team) (string, error)
  DeleteTeam(context.Context, string) (int64, int64, int64, error)
  //GetTeamByTeamId(context.Context, string) (*v1.Team, error)
  // GetTeamsByUserId(context.Context, string) ([]v1.Team, error)
  AddMember(context.Context, string, string) (string, error)
  RemoveMember(context.Context, string, string) (int64, error)
  //UpsertProject(context.Context, string, *v1.Project) (int64, error)
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
    return nil, err
  }
  return c, nil
}

// Creates a Team
// input: context-the current handler context, team-team object from gRPC endpoint handler
// output ON SUCCESS: string - id of newly inserted team, error - nil
// output ON FAILURE: string - nil, error - the error object from whatever created the error
func (r *teamRepository) CreateTeam(ctx context.Context, team *v1.Team) (string, error) {
  // prepare sql statements for teams, skills, members
  teamStmt := `INSERT INTO teams (leader, team_name, open_roles, size, last_active) VALUES(?, ?, ?, ?, ?)`
  memberStmt := `INSERT INTO members (member_id, member_name, team_id) VALUES %s`
  skillStmt := `INSERT INTO skills (skill_name, team_id) VALUES %s`

  fmt.Fprintf(os.Stderr, "In createteam repo\n")

  // start transaction
  tx, err := r.db.BeginTx(ctx, &sql.TxOptions{})
  if err != nil {
    return "transaction begin", err
  }

  // insert team into teams table capturing the id
  result, err := tx.Exec(teamStmt, team.Leader, team.Name, team.OpenRoles, team.Size, team.LastActive)
  if err != nil {
    tx.Rollback()
    return "Exec team stmt", err
  }
  // gather the id of the inserted team
  teamId, err := result.LastInsertId()
  if err != nil {
    tx.Rollback()
    return "team insertId()", err
  }

  // create bulk array insert values.
  memberStrings := []string{}
  memberArgs := []interface{}{}
  for _, w := range team.Members {
    memberStrings = append(memberStrings, "(?, ?, ?)")

    memberArgs = append(memberArgs, w.Id)
    memberArgs = append(memberArgs, w.Name)
    memberArgs = append(memberArgs, teamId)
  }

  // create member sql statement
  memberStmt = fmt.Sprintf(memberStmt, strings.Join(memberStrings, ","))
  fmt.Fprintf(os.Stderr, "memberStmt: %v\n", memberStmt)

  // insert members into members table including team_id field
  _, err = tx.Exec(memberStmt, memberArgs...)
  if err != nil {
    tx.Rollback()
    return "Exec member stmt", err
  }

  // create bulk array insert values.
  skillStrings := []string{}
  skillArgs := []interface{}{}
  for _, w := range team.Skills {
    skillStrings = append(skillStrings, "(?, ?)")

    skillArgs = append(skillArgs, w)
    skillArgs = append(skillArgs, teamId)
  }

  // create skill sql statement
  skillStmt = fmt.Sprintf(skillStmt, strings.Join(skillStrings, ","))
  fmt.Fprintf(os.Stderr, "skillStmt: %v\n", skillStmt)

  // insert skills into skills table including team_id field
  _, err = tx.Exec(skillStmt, skillArgs...)
  if err != nil {
    tx.Rollback()
    return "Exec skill stmt", err
  }

  // commit transaction
  err = tx.Commit()
  if err != nil {
    return "Commit()", err
  }

  // return id of newly inserted team and no error
  return strconv.FormatInt(teamId, 10), nil
}

// Deletes a Team
// input: context-the current handler context, team-team object from gRPC endpoint handler
// output ON SUCCESS: string - id of newly inserted team, error - nil
// output ON FAILURE: string - nil, error - the error object from whatever created the error
func (r *teamRepository) DeleteTeam(ctx context.Context, id string) (int64, int64, int64, error) {
  // prepare sql statements for teams, skills, members
  teamStmt := `DELETE FROM teams WHERE id=?`
  memberStmt := `DELETE FROM members WHERE team_id=?`
  skillStmt := `DELETE FROM skills WHERE team_id=?`

  fmt.Fprintf(os.Stderr, "In createteam repo\n")
  idAsInt, _ := strconv.Atoi(id)

  // start transaction
  tx, err := r.db.BeginTx(ctx, &sql.TxOptions{})
  if err != nil {
    return -1, -1, -1, err
  }

  // insert team into teams table capturing the id
  memResult, err := tx.Exec(memberStmt, idAsInt)
  if err != nil {
    tx.Rollback()
    return -1, -1, -1, err
  }
  // gather the id of the inserted team
  memRows, err := memResult.RowsAffected()
  if err != nil {
    tx.Rollback()
    return -1, -1, -1, err
  }

  // insert team into teams table capturing the id
  skillResult, err := tx.Exec(skillStmt, idAsInt)
  if err != nil {
    tx.Rollback()
    return -1, -1, -1, err
  }
  // gather the id of the inserted team
  skillRows, err := skillResult.RowsAffected()
  if err != nil {
    tx.Rollback()
    return -1, -1, -1, err
  }

  // delete team
  result, err := tx.Exec(teamStmt, idAsInt)
  if err != nil {
    tx.Rollback()
    return -1, -1, -1, err
  }
  // gather the num of rows deleted
  teamRows, err := result.RowsAffected()
  if err != nil {
    tx.Rollback()
    return -1, -1, -1, err
  }

  // commit transaction
  err = tx.Commit()
  if err != nil {
    return -1, -1, -1, err
  }

  // return number of rows deleted per object
  return teamRows, memRows, skillRows, nil
}

// Adds a member to a team
// input: context-the current handler context, id of team to be inserted to, user id of new member
// output ON SUCCESS: string - member number of new member within team, error - nil
// output ON FAILURE: string - nil, error - the error object from whatever created the error
func (r *teamRepository) AddMember(ctx context.Context, teamId string, userId string) (string, error) {
  // prepare sql statements for teams, skills, members
  // need to change this to update
  teamStmt := `INSERT INTO teams (leader, team_name, open_roles, size, last_active) VALUES(?, ?, ?, ?, ?)`

  memberStmt := `INSERT INTO members (member_id, team_id) VALUES (?, ?)`

  // start transaction
  tx, err := r.db.BeginTx(ctx, &sql.TxOptions{})
  if err != nil {
    return -1, -1, -1, err
  }

  // insert team into teams table capturing the id
  memResult, err := tx.Exec(memberStmt, userId, teamId)
  if err != nil {
    tx.Rollback()
    return -1, err
  }
  // gather the id of the inserted team
  memRows, err := memResult.RowsAffected()
  if err != nil {
    tx.Rollback()
    return -1, err
  }

  // commit transaction
  err = tx.Commit()
  if err != nil {
    return -1, err
  }
  return memRows, nil
}

// Removes a member from a team
// input: context-the current handler context, id of team to be inserted to, user id of new member
// output ON SUCCESS: string - member number of new member within team, error - nil
// output ON FAILURE: string - nil, error - the error object from whatever created the error
func (r *teamRepository) RemoveMember(ctx context.Context, teamId string, memberId string) (string, error) {
  // prepare sql statements for teams, skills, members
  // need to change this to update
  teamStmt := `INSERT INTO teams (leader, team_name, open_roles, size, last_active) VALUES(?, ?, ?, ?, ?)`

  memberStmt := `DELETE FROM members WHERE team_id=? AND member_id=?`
  // start transaction
  tx, err := r.db.BeginTx(ctx, &sql.TxOptions{})
  if err != nil {
    return -1, -1, -1, err
  }

  // insert team into teams table capturing the id
  memResult, err := tx.Exec(memberStmt, teamId, memberId)
  if err != nil {
    tx.Rollback()
    return -1, err
  }
  // gather the id of the inserted team
  memRows, err := memResult.RowsAffected()
  if err != nil {
    tx.Rollback()
    return -1, err
  }

  // commit transaction
  err = tx.Commit()
  if err != nil {
    return -1, err
  }
  return memRows, nil
}
