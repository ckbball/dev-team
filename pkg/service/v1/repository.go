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
  AddMember(context.Context, *v1.MemberUpsertRequest) (string, error)
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
  memberStmt := `INSERT INTO members (user_id, member_email, team_id, member_role) VALUES %s`
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
  // iterate over each member and construct sql arguments
  for _, w := range team.Members {
    memberStrings = append(memberStrings, "(?, ?, ?, ?)")

    memberArgs = append(memberArgs, w.Id)
    memberArgs = append(memberArgs, w.Email)
    memberArgs = append(memberArgs, teamId)
    memberArgs = append(memberArgs, w.Role)
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

  // delete all members of a specific team
  memResult, err := tx.Exec(memberStmt, idAsInt)
  if err != nil {
    tx.Rollback()
    return -1, -1, -1, err
  }
  // gather the number of rows deleted
  memRows, err := memResult.RowsAffected()
  if err != nil {
    tx.Rollback()
    return -1, -1, -1, err
  }

  // delete skills of a specific team
  skillResult, err := tx.Exec(skillStmt, idAsInt)
  if err != nil {
    tx.Rollback()
    return -1, -1, -1, err
  }
  // gather the number of rows deleted
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
func (r *teamRepository) AddMember(ctx context.Context, req *v1.MemberUpsertRequest) (string, error) {
  // prepare sql statements for teams, skills, members
  // need to change this to update

  memberStmt := `INSERT INTO members (user_id, team_id, member_email, member_role) VALUES (?, ?, ?, ?)`

  // start transaction
  tx, err := r.db.BeginTx(ctx, &sql.TxOptions{})
  if err != nil {
    return "", err
  }

  // insert team into teams table capturing the id
  memResult, err := tx.Exec(memberStmt, strconv.ParseInt(req.MemberId, 10, 64), req.Id, req.MemberEmail, req.Role)
  if err != nil {
    tx.Rollback()
    return "", err
  }
  // gather the id of the inserted team
  memId, err := memResult.LastInsertId()
  if err != nil {
    tx.Rollback()
    return "", err
  }

  // commit transaction
  err = tx.Commit()
  if err != nil {
    return "", err
  }
  return strconv.FormatInt(memId, 10), nil
}

// Removes a member from a team
// input: context-the current handler context, id of team to be inserted to, user id of new member
// output ON SUCCESS: string - member number of new member within team, error - nil
// output ON FAILURE: string - nil, error - the error object from whatever created the error
func (r *teamRepository) RemoveMember(ctx context.Context, teamId string, memberId string) (int64, error) {
  // prepare sql statements for teams, skills, members
  // need to change this to update

  memberStmt := `DELETE FROM members WHERE team_id=? AND id=?`
  // start transaction
  tx, err := r.db.BeginTx(ctx, &sql.TxOptions{})
  if err != nil {
    return -1, err
  }

  // delete member from specified team
  memResult, err := tx.Exec(memberStmt, teamId, memberId)
  if err != nil {
    tx.Rollback()
    return -1, err
  }
  // gather the number of rows deleted
  numRows, err := memResult.RowsAffected()
  if err != nil {
    tx.Rollback()
    return -1, err
  }

  // commit transaction
  err = tx.Commit()
  if err != nil {
    return -1, err
  }
  return numRows, nil
}

func (r *teamRepository) UpsertProject(ctx context.Context, teamId string, project *v1.Project) (int64, error) {
  // prepare sql statements
  projStmt := `INSERT INTO projects (goal, project_name, github_link, team_id, complexity, duration) VALUES(?, ?, ?, ?, ?, ?)`
  langStmt := `INSERT INTO languages (lang_name, project_id) VALUES %s`

  fmt.Fprintf(os.Stderr, "In createteam repo\n")

  // start transaction
  tx, err := r.db.BeginTx(ctx, &sql.TxOptions{})
  if err != nil {
    fmt.Fprintf(os.Stderr, "error in BeginTx")
    return -1, err
  }

  // insert project into projects table capturing the id
  result, err := tx.Exec(projStmt, project.Description, project.Name, project.GithubLink, teamId, project.Complexity, project.Duration)
  if err != nil {
    fmt.Fprintf(os.Stderr, "error in Exec(Project)")
    tx.Rollback()
    return -1, err
  }
  // gather the id of the inserted project
  projectId, err := result.LastInsertId()
  if err != nil {
    fmt.Fprintf(os.Stderr, "error in LastInsertId")
    tx.Rollback()
    return -1, err
  }

  // create bulk array insert values.
  langStrings := []string{}
  langArgs := []interface{}{}
  for _, w := range project.Languages {
    langStrings = append(langStrings, "(?, ?)")

    langArgs = append(langArgs, w)
    langArgs = append(langArgs, projectId)
  }

  // create lang sql statement
  langStmt = fmt.Sprintf(langStmt, strings.Join(langStrings, ","))
  fmt.Fprintf(os.Stderr, "langStmt: %v\n", langStmt)

  // insert langs into langs table including team_id field
  _, err = tx.Exec(langStmt, langArgs...)
  if err != nil {
    fmt.Fprintf(os.Stderr, "error in Exec(Language)")
    tx.Rollback()
    return -1, err
  }

  // commit transaction
  err = tx.Commit()
  if err != nil {
    fmt.Fprintf(os.Stderr, "error in Commit()")
    return -1, err
  }

  // return id of newly inserted team and no error
  return projectId, nil
}
