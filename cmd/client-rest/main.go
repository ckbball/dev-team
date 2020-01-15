package main

import (
  "encoding/json"
  "flag"
  "fmt"
  "io/ioutil"
  "log"
  "net/http"
  "strings"
  "time"
)

func main() {
  // get configuration
  address := flag.String("server", "http://localhost:8082", "HTTP gateway url, e.g. http://localhost:8082")
  flag.Parse()

  t := time.Now().In(time.UTC)
  pfx := t.Format(time.RFC3339Nano)

  var body string
  log.Printf("\nAddress received: %s\n", *address)

  // Call CreateTeam
  resp, err := http.Post(*address+"/v1/teams", "application/json", strings.NewReader(fmt.Sprintf(`
    {
      "api":"v1",
      "team": {
        "leader": "loola@gmail.com",
        "name": "kappa",
        "open_roles": 3,
        "last_active": 0,
        "size": 3,
        "members": [{"email":"goja@yahoo.com", "id":1, "role":"backend"}, {"email":"linka@gmail.com", "id":2, "role":"backend"}],
        "skills": ["frontend", "design"]
      }
    }
  `, pfx, pfx, pfx)))
  if err != nil {
    log.Fatalf("failed to call CreateTeam method: %v", err)
  }
  bodyBytes, err := ioutil.ReadAll(resp.Body)
  resp.Body.Close()
  if err != nil {
    body = fmt.Sprintf("failed read CreateTeam response body: %v", err)
  } else {
    body = string(bodyBytes)
  }
  log.Printf("CreateTeam response: Code=%d, Body=%s\n\n", resp.StatusCode, body)

  var upsertTeam struct {
    Api    string `json:"api"`
    Status string `json:"status"`
    Id     string `json:"id"`
  }
  err = json.Unmarshal(bodyBytes, &upsertTeam)
  if err != nil {
    log.Fatalf("failed to unmarshal JSON response of CreateTeam method: %v", err)
    fmt.Println("error:", err)
  }
  log.Printf("upsertTeam struct: %s\n", upsertTeam)

  createdTeamId := upsertTeam.Id

  // Call CreateTeam
  resp, err = http.Post(*address+"/v1/teams", "application/json", strings.NewReader(fmt.Sprintf(`
    {
      "api":"v1",
      "team": {
        "leader": "geraswee@gmail.com",
        "name": "nadd",
        "open_roles": 2,
        "last_active": 1,
        "size": 2,
        "members": [{"email":"geraswee@gmail.com", "id":5, "role":"design"}, {"email":"linka@gmail.com", "id":2, "role":"backend"}],
        "skills": ["frontend", "devops"]
      }
    }
  `, pfx, pfx, pfx)))
  if err != nil {
    log.Fatalf("failed to call CreateTeam method: %v", err)
  }
  bodyBytes, err = ioutil.ReadAll(resp.Body)
  resp.Body.Close()
  if err != nil {
    body = fmt.Sprintf("failed read CreateTeam response body: %v", err)
  } else {
    body = string(bodyBytes)
  }
  log.Printf("CreateTeam2 response: Code=%d, Body=%s\n\n", resp.StatusCode, body)

  // Call AddMember
  resp, err = http.Post(*address+"/v1/teams/"+createdTeamId+"/members", "application/json", strings.NewReader(fmt.Sprintf(`
    {
      "api":"v1",
      "member_id": "3",
      "member_email": "freddy@yahoo.com",
      "role": "frontend"
    }
  `, pfx, pfx, pfx)))
  if err != nil {
    log.Fatalf("failed to call AddMember method: %v", err)
  }
  bodyBytes, err = ioutil.ReadAll(resp.Body)
  resp.Body.Close()
  if err != nil {
    body = fmt.Sprintf("failed read AddMember response body: %v", err)
  } else {
    body = string(bodyBytes)
  }
  log.Printf("AddMember response: Code=%d, Body=%s\n\n", resp.StatusCode, body)

  // Call GetTeamByTeamId
  req, err := http.NewRequest("GET", fmt.Sprintf("%s%s/%s", *address, "/v1/teams", createdTeamId), nil)
  resp, err = http.DefaultClient.Do(req)
  if err != nil {
    log.Fatalf("failed to call GetTeamByTeamId method: %v", err)
  }
  bodyBytes, err = ioutil.ReadAll(resp.Body)
  resp.Body.Close()
  if err != nil {
    body = fmt.Sprintf("failed read GetTeamByTeamId response body: %v", err)
  } else {
    body = string(bodyBytes)
  }
  log.Printf("GetTeamByTeamId response: Code=%d, Body=%s\n\n", resp.StatusCode, body)

  // Call UpsertProject
  resp, err = http.Post(*address+"/v1/teams/"+createdTeamId+"/project", "application/json", strings.NewReader(fmt.Sprintf(`
    {
      "api":"v1",
      "project": {
        "description": "Full stack app to track food",
        "languages": ["javascript", "html", "css", "python"],
        "name": "Food Track",
        "github_link": "github.com/foodtrack",
        "complexity": 3,
        "duration": 4
      }
    }
  `, pfx, pfx, pfx)))
  if err != nil {
    log.Fatalf("failed to call UpsertProject method: %v", err)
  }
  bodyBytes, err = ioutil.ReadAll(resp.Body)
  resp.Body.Close()
  if err != nil {
    body = fmt.Sprintf("failed read UpsertProject response body: %v", err)
  } else {
    body = string(bodyBytes)
  }
  log.Printf("UpsertProject response: Code=%d, Body=%s\n\n", resp.StatusCode, body)

  // Call Remove Member
  req, err = http.NewRequest("DELETE", fmt.Sprintf("%s%s/%s/%s/%s", *address, "/v1/teams", createdTeamId, "members", "2"), nil)
  resp, err = http.DefaultClient.Do(req)
  if err != nil {
    log.Fatalf("failed to call RemoveMember method: %v", err)
  }
  bodyBytes, err = ioutil.ReadAll(resp.Body)
  resp.Body.Close()
  if err != nil {
    body = fmt.Sprintf("failed read RemoveMember response body: %v", err)
  } else {
    body = string(bodyBytes)
  }
  log.Printf("RemoveMember response: Code=%d, Body=%s\n\n", resp.StatusCode, body)

  // Call GetTeamsByUserId
  req, err = http.NewRequest("GET", fmt.Sprintf("%s%s", *address, "/v1/teams/users/2"), nil)
  resp, err = http.DefaultClient.Do(req)
  if err != nil {
    log.Fatalf("failed to call GetTeamsByUserId method: %v", err)
  }
  bodyBytes, err = ioutil.ReadAll(resp.Body)
  resp.Body.Close()
  if err != nil {
    body = fmt.Sprintf("failed read GetTeamsByUserId response body: %v", err)
  } else {
    body = string(bodyBytes)
  }
  log.Printf("GetTeamsByUserId response: Code=%d, Body=%s\n\n", resp.StatusCode, body)

  // When implemented call getAllTeams()
  req, err = http.NewRequest("GET", fmt.Sprintf("%s%s", *address, "/v1/teams?page=0&limit=10"), nil)
  resp, err = http.DefaultClient.Do(req)
  if err != nil {
    log.Fatalf("failed to call GetTeams method: %v", err)
  }
  bodyBytes, err = ioutil.ReadAll(resp.Body)
  resp.Body.Close()
  if err != nil {
    body = fmt.Sprintf("failed read GetTeams response body: %v", err)
  } else {
    body = string(bodyBytes)
  }
  log.Printf("GetTeams response: Code=%d, Body=%s\n\n", resp.StatusCode, body)

  // Call DeleteTeam
  req, err = http.NewRequest("DELETE", fmt.Sprintf("%s%s/%s", *address, "/v1/teams", createdTeamId), nil)
  resp, err = http.DefaultClient.Do(req)
  if err != nil {
    log.Fatalf("failed to call DeleteTeam method: %v", err)
  }
  bodyBytes, err = ioutil.ReadAll(resp.Body)
  resp.Body.Close()
  if err != nil {
    body = fmt.Sprintf("failed read DeleteTeam response body: %v", err)
  } else {
    body = string(bodyBytes)
  }
  log.Printf("DeleteTeam response: Code=%d, Body=%s\n\n", resp.StatusCode, body)

}
