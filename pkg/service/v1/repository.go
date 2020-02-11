package v1

import (
  "context"
  "database/sql"
  "errors"
  "fmt"
  "os"
  "strconv"
  "strings"
  // "time"

  v1 "github.com/ckbball/dev-team/pkg/api/v1"
)

type repository interface {
  CreateTeam(context.Context, *v1.Team) (string, error)
  DeleteTeam(context.Context, string) (int64, int64, int64, error)
  GetTeamByTeamId(context.Context, string) (*v1.Team, error)
  GetTeamByTeamName(context.Context, string) (*v1.Team, error)
  GetTeamsByUserId(context.Context, string) ([]*v1.Team, error)
  AddMember(context.Context, *v1.MemberUpsertRequest) (string, error)
  RemoveMember(context.Context, string, string) (int64, error)
  UpsertProject(context.Context, string, *v1.Project) (int64, error)
  GetTeams(context.Context, *v1.GetTeamsRequest) ([]*v1.Team, error)
  CountUserTeams(context.Context, string) (int, error) // in: userId as string || out: Number of teams user owns as int, error
  CheckUserOwnsTeam(context.Context, string, string) (bool, error)
  CheckMemberExists(context.Context, string, string) (bool, error)
  CheckTeamSize(context.Context, string) (bool, error)
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

  //

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

  // decide whether to add owner to member list in initial creation
  if len(team.Members) > 0 {
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

    fmt.Fprintf(os.Stderr, "memberStmt in CreateTeam: %v\n", memberStmt)

    // insert members into members table including team_id field
    _, err = tx.Exec(memberStmt, memberArgs...)
    if err != nil {
      tx.Rollback()
      return "Exec member stmt", err
    }
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
  projStmt := `DELETE FROM projects WHERE team_id=?`
  langStmt := `DELETE FROM languages WHERE team_id=?`

  idAsInt, _ := strconv.Atoi(id)

  // start transaction
  tx, err := r.db.BeginTx(ctx, &sql.TxOptions{})
  if err != nil {
    return -1, -1, -1, err
  }

  // delete all languages of a specific team
  langResult, err := tx.Exec(langStmt, idAsInt)
  if err != nil {
    tx.Rollback()
    return -1, -1, -1, err
  }
  // gather the number of rows deleted
  _, err = langResult.RowsAffected()
  if err != nil {
    tx.Rollback()
    return -1, -1, -1, err
  }

  // delete all projects of a specific team
  projResult, err := tx.Exec(projStmt, idAsInt)
  if err != nil {
    tx.Rollback()
    return -1, -1, -1, err
  }
  // gather the number of rows deleted
  _, err = projResult.RowsAffected()
  if err != nil {
    tx.Rollback()
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
  teamStmt := `UPDATE teams SET open_roles = open_roles - 1 WHERE id=?`

  // start transaction
  tx, err := r.db.BeginTx(ctx, &sql.TxOptions{})
  if err != nil {
    return "", err
  }

  convert, _ := strconv.ParseInt(req.MemberId, 10, 64)

  // insert team into teams table capturing the id
  memResult, err := tx.Exec(memberStmt, convert, req.TeamId, req.MemberEmail, req.Role)
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

  // decrement number of open roles on specific team
  _, err = tx.Exec(teamStmt, req.TeamId)
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
  langStmt := `INSERT INTO languages (lang_name, team_id) VALUES %s`

  projDel := `DELETE FROM projects WHERE team_id=? `
  langDel := `DELETE FROM languages WHERE team_id=? `

  // start transaction
  tx, err := r.db.BeginTx(ctx, &sql.TxOptions{})
  if err != nil {
    fmt.Fprintf(os.Stderr, "error in BeginTx")
    return -1, err
  }

  // if team has project already delete it and its languages
  // delete languages from specified team
  _, err = tx.Exec(langDel, teamId)
  if err != nil {
    tx.Rollback()
    return -1, err
  }

  // delete projects from specified team
  _, err = tx.Exec(projDel, teamId)
  if err != nil {
    tx.Rollback()
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
    langArgs = append(langArgs, teamId)
  }

  // create lang sql statement
  langStmt = fmt.Sprintf(langStmt, strings.Join(langStrings, ","))

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

func (r *teamRepository) GetTeamByTeamName(ctx context.Context, name string) (*v1.Team, error) {
  // prepare sql statements for team, member, skills, project, languages
  teamStmt := `SELECT leader, team_name, open_roles, size, last_active, id FROM teams WHERE team_name=?`
  memberStmt := `SELECT user_id, member_email, member_role FROM members WHERE team_id=?`
  skillStmt := `SELECT skill_name FROM skills WHERE team_id=?`
  projStmt := `SELECT goal, project_name, github_link, complexity, duration FROM projects WHERE team_id=?`
  langStmt := `SELECT lang_name FROM languages WHERE team_id=?`

  team := &v1.Team{}
  members := []*v1.Member{}
  skills := []string{}
  project := &v1.Project{}
  languages := []string{}

  // start transaction
  tx, err := r.db.BeginTx(ctx, &sql.TxOptions{})
  if err != nil {
    fmt.Fprintf(os.Stderr, "error in BeginTx")
    return nil, err
  }

  name = strings.ToLower(name)

  // execute team sql statement
  row := tx.QueryRow(teamStmt, name)

  // scan fields into team
  err = row.Scan(&team.Leader, &team.Name, &team.OpenRoles, &team.Size, &team.LastActive, &team.Id)
  if err == sql.ErrNoRows {
    return nil, errors.New("team Query: no matching record found")
  } else if err != nil {
    return nil, err
  }

  // execute member statement
  memberRows, err := tx.Query(memberStmt, team.Id)
  if err != nil {
    fmt.Fprintf(os.Stderr, "error in members Query")
    return nil, err
  }

  defer memberRows.Close()

  // scan each member row into members variable
  for memberRows.Next() {
    s := &v1.Member{}

    err = memberRows.Scan(&s.Id, &s.Email, &s.Role)
    if err != nil {
      return nil, err
    }
    members = append(members, s)
  }

  if err = memberRows.Err(); err != nil {
    return nil, err
  }

  // add retrieved members to team
  team.Members = members

  // execute skills statement query
  skillRows, err := tx.Query(skillStmt, team.Id)
  if err != nil {
    fmt.Fprintf(os.Stderr, "error in skill Query")
    return nil, err
  }

  //scan skills
  defer skillRows.Close()

  // scan each member row into skills variable
  for skillRows.Next() {
    s := ""

    err = skillRows.Scan(&s)
    if err != nil {
      return nil, err
    }
    skills = append(skills, s)
  }

  if err = skillRows.Err(); err != nil {
    return nil, err
  }

  // add retrieved skills to team
  team.Skills = skills

  // execute project statement query
  projectRow := tx.QueryRow(projStmt, team.Id)

  // scan project
  err = projectRow.Scan(&project.Description, &project.Name, &project.GithubLink, &project.Complexity, &project.Duration)
  if err == sql.ErrNoRows {
    // for development this is fine
    // fmt.Fprintf(os.Stderr, "team has no project")
  } else if err != nil {
    return nil, err
  }

  // execute languages statement query
  languagesRows, err := tx.Query(langStmt, team.Id)
  if err != nil {
    fmt.Fprintf(os.Stderr, "error in languages Query")
    return nil, err
  }
  // scan languages
  defer languagesRows.Close()

  // scan each languages row into languages variable
  for languagesRows.Next() {
    s := ""

    err = languagesRows.Scan(&s)
    if err != nil {
      return nil, err
    }
    languages = append(languages, s)
  }

  if err = languagesRows.Err(); err != nil {
    return nil, err
  }

  // add retrieved languagess to team
  project.Languages = languages

  // commit transaction
  err = tx.Commit()
  if err != nil {
    fmt.Fprintf(os.Stderr, "error in Commit()")
    return nil, err
  }

  team.Project = project

  // return id of newly inserted team and no error
  return team, nil
}

func (r *teamRepository) GetTeamByTeamId(ctx context.Context, id string) (*v1.Team, error) {
  // prepare sql statements for team, member, skills, project, languages
  teamStmt := `SELECT leader, team_name, open_roles, size, last_active, id FROM teams WHERE id=?`
  memberStmt := `SELECT user_id, member_email, member_role FROM members WHERE team_id=?`
  skillStmt := `SELECT skill_name FROM skills WHERE team_id=?`
  projStmt := `SELECT goal, project_name, github_link, complexity, duration FROM projects WHERE team_id=?`
  langStmt := `SELECT lang_name FROM languages WHERE team_id=?`

  team := &v1.Team{}
  members := []*v1.Member{}
  skills := []string{}
  project := &v1.Project{}
  languages := []string{}

  // start transaction
  tx, err := r.db.BeginTx(ctx, &sql.TxOptions{})
  if err != nil {
    fmt.Fprintf(os.Stderr, "error in BeginTx")
    return nil, err
  }

  // execute team sql statement
  row := tx.QueryRow(teamStmt, id)

  // scan fields into team
  err = row.Scan(&team.Leader, &team.Name, &team.OpenRoles, &team.Size, &team.LastActive, &team.Id)
  if err == sql.ErrNoRows {
    return nil, errors.New("team Query: no matching record found")
  } else if err != nil {
    return nil, err
  }

  // execute member statement
  memberRows, err := tx.Query(memberStmt, id)
  if err != nil {
    fmt.Fprintf(os.Stderr, "error in members Query")
    return nil, err
  }

  defer memberRows.Close()

  // scan each member row into members variable
  for memberRows.Next() {
    s := &v1.Member{}

    err = memberRows.Scan(&s.Id, &s.Email, &s.Role)
    if err != nil {
      return nil, err
    }
    members = append(members, s)
  }

  if err = memberRows.Err(); err != nil {
    return nil, err
  }

  // add retrieved members to team
  team.Members = members

  // execute skills statement query
  skillRows, err := tx.Query(skillStmt, id)
  if err != nil {
    fmt.Fprintf(os.Stderr, "error in skill Query")
    return nil, err
  }

  //scan skills
  defer skillRows.Close()

  // scan each member row into skills variable
  for skillRows.Next() {
    s := ""

    err = skillRows.Scan(&s)
    if err != nil {
      return nil, err
    }
    skills = append(skills, s)
  }

  if err = skillRows.Err(); err != nil {
    return nil, err
  }

  // add retrieved skills to team
  team.Skills = skills

  // execute project statement query
  projectRow := tx.QueryRow(projStmt, id)

  // scan project
  err = projectRow.Scan(&project.Description, &project.Name, &project.GithubLink, &project.Complexity, &project.Duration)
  if err == sql.ErrNoRows {
    // for development this is fine
    // fmt.Fprintf(os.Stderr, "team has no project")
  } else if err != nil {
    return nil, err
  }

  // execute languages statement query
  languagesRows, err := tx.Query(langStmt, id)
  if err != nil {
    fmt.Fprintf(os.Stderr, "error in languages Query")
    return nil, err
  }
  // scan languages
  defer languagesRows.Close()

  // scan each languages row into languages variable
  for languagesRows.Next() {
    s := ""

    err = languagesRows.Scan(&s)
    if err != nil {
      return nil, err
    }
    languages = append(languages, s)
  }

  if err = languagesRows.Err(); err != nil {
    return nil, err
  }

  // add retrieved languagess to team
  project.Languages = languages

  // commit transaction
  err = tx.Commit()
  if err != nil {
    fmt.Fprintf(os.Stderr, "error in Commit()")
    return nil, err
  }

  team.Project = project

  // return id of newly inserted team and no error
  return team, nil
}

func (r *teamRepository) GetTeamsByUserId(ctx context.Context, id string) ([]*v1.Team, error) {
  // prepare sql statements for team, member, skills, project, languages
  memberStmt := `SELECT team_id FROM members WHERE user_id=?`

  // prepare necessary variables
  teams := []*v1.Team{}
  members := []int{}

  fmt.Fprintf(os.Stderr, "id: %v\n")

  // select all member rows where user_id = id
  // for each row, call GetTeamByTeamId append response to teams var
  // return teams

  // execute member statement
  memberRows, err := r.db.Query(memberStmt, id)
  if err != nil {
    fmt.Fprintf(os.Stderr, "error in members Query")
    return nil, err
  }

  defer memberRows.Close()

  // scan each member row into members variable
  for memberRows.Next() {
    var s int

    // scan members.team_id into member's Id field
    err = memberRows.Scan(&s)
    if err != nil {
      return nil, err
    }
    members = append(members, s)
  }

  fmt.Fprintf(os.Stderr, "members: %v\n", members)

  if err = memberRows.Err(); err != nil {
    return nil, err
  }

  // iterate over each member, calling GetTeamByTeamId() with team_id of each team user is from
  for _, mem := range members {
    team := &v1.Team{}

    team, err = r.GetTeamByTeamId(ctx, strconv.Itoa(int(mem)))
    fmt.Fprintf(os.Stderr, "team: %v\n", team)
    if err != nil {
      return teams, err
    }

    teams = append(teams, team)
  }

  // return list of teams that user is in
  return teams, nil
}

func (r *teamRepository) GetTeams(ctx context.Context, req *v1.GetTeamsRequest) ([]*v1.Team, error) {
  // calculate page id
  // range over id's from page_id to page_id + limit calling GetTeamByTeamId and appending to teams var
  // return teams
  var item_id int64
  item_id = 0
  if req.Page > 1 {
    item_id = req.Limit * (req.Page - 1)
  }

  // time_diff := time.Now().Unix() - 2592000

  // query for future in production
  // `SELECT id FROM teams WHERE id > ? AND last_active > ? ORDER BY id ASC LIMIT ?`

  // technology query param is frozen for now

  teamStmt := ""
  var teamRows *sql.Rows
  var err error

  if req.Role != "" {
    teamStmt = `SELECT team_id FROM skills WHERE skill_name=? ORDER BY id ASC LIMIT ?`
    teamRows, err = r.db.Query(teamStmt, req.Role, req.Limit*10)
  } else if req.Level != 0 {
    teamStmt = `SELECT team_id FROM projects WHERE complexity=? ORDER BY id ASC LIMIT ?`
    teamRows, err = r.db.Query(teamStmt, req.Level, req.Limit)
  } else {
    teamStmt = `SELECT id FROM teams WHERE id > ? ORDER BY id ASC LIMIT ?`
    teamRows, err = r.db.Query(teamStmt, item_id, req.Limit)
  }

  teams := []*v1.Team{}
  teamsOut := []*v1.Team{}

  fmt.Fprintf(os.Stderr, "teamStmt: %v", teamStmt)
  fmt.Fprintf(os.Stderr, "teamsOut: %v", teamsOut)

  if err != nil {
    fmt.Fprintf(os.Stderr, "error in GetTeams Query")
    return nil, err
  }

  defer teamRows.Close()

  // scan each team row into teams variable
  for teamRows.Next() {
    s := &v1.Team{}

    // scan teams.team_id into team's Id field
    err = teamRows.Scan(&s.Id)
    if err != nil {
      return nil, err
    }
    teams = append(teams, s)
  }

  if err = teamRows.Err(); err != nil {
    return nil, err
  }

  // range over list of ids calling GetTeamByTeamId
  for _, mem := range teams {
    team := &v1.Team{}

    team, err = r.GetTeamByTeamId(ctx, mem.Id)
    if err != nil {
      return teamsOut, err
    }

    teamsOut = append(teamsOut, team)
  }

  // return list of teams that user is in
  return teamsOut, nil
  //return teams, nil
}

// takes a userId and searches db for how many teams this user owns
// returns number of teams owned and an error
func (r *teamRepository) CountUserTeams(ctx context.Context, userId string) (int, error) {
  countStmt := `SELECT COUNT(*) FROM teams WHERE leader=?`

  fmt.Fprintf(os.Stderr, "request userID: %v\n", userId)
  rows, err := r.db.Query(countStmt, userId)
  if err != nil {
    fmt.Fprintf(os.Stderr, "error in CountUserTeams Query")
    return -1, err
  }
  defer rows.Close()
  rows.Next()
  var count int
  err = rows.Scan(&count)
  fmt.Fprintf(os.Stderr, "count: %v\n", count)
  if err != nil {
    return -1, err
  }
  return count, nil
}

// takes a userId and searches db for how many teams this user owns
// returns true if user owns team false if not
func (r *teamRepository) CheckUserOwnsTeam(ctx context.Context, userId, teamId string) (bool, error) {
  // select leader FROM teams WHERE leader=userId AND id=teamId
  checkStmt := `SELECT leader FROM teams WHERE leader=? AND id=?`
  // get row
  row := r.db.QueryRow(checkStmt, userId, teamId)
  var leader string
  // scan fields into team
  err := row.Scan(&leader)
  if err == sql.ErrNoRows {
    return false, errors.New("CheckUserOwnsTeam Query: no matching record found")
  } else if err != nil {
    return false, err
  }

  // if row exists user owns the team and return true else return false
  return true, nil
}

// takes a userId and teamId and searches db if user is already on team
// returns true if user on team false if not
func (r *teamRepository) CheckMemberExists(ctx context.Context, userId, teamId string) (bool, error) {
  // select role FROM members WHERE user_id=userId AND team_id=teamId
  checkStmt := `SELECT member_role FROM members WHERE user_id=? AND team_id=?`
  // get row
  row := r.db.QueryRow(checkStmt, userId, teamId)
  var role string
  // scan fields into team
  err := row.Scan(&role)
  if err == sql.ErrNoRows {
    fmt.Fprintf(os.Stderr, "rows checkMember exists: \n")
    return false, nil
  } else if err != nil {
    return false, err
  }

  // if row exists user is a member on team
  return true, nil
}

// takes a userId and teamId and searches db if user is already on team
// returns true if user on team false if not
func (r *teamRepository) CheckTeamSize(ctx context.Context, teamId string) (bool, error) {
  // select role FROM members WHERE user_id=userId AND team_id=teamId
  checkStmt := `SELECT open_roles FROM teams WHERE id=?`
  // get row
  row := r.db.QueryRow(checkStmt, teamId)
  var spots int
  // scan fields into team
  err := row.Scan(&spots)
  if err == sql.ErrNoRows {
    return false, nil
  } else if err != nil {
    return false, err
  }

  // if open_roles is less than zero that means team has reached max size
  if spots < 1 {
    return true, nil
  }

  return false, nil
}

// ---------------------------- HELPER FUNCTIONS -------------------------------

func (r *teamRepository) checkCount(rows *sql.Rows) (int, error) {
  var count int
  for rows.Next() {
    err := rows.Scan(&count)
    if err != nil {
      return -1, err
    }
  }
  return count, nil
}
