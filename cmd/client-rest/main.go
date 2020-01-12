package main

import (
  //"encoding/json"
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
        "name": "loola",
        "open_roles": 3,
        "last_active": 0,
        "size": 3,
        "members": [{"name":"goja@yahoo.com", "id":1}, {"name":"linka@gmail.com", "id":2}],
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

  /*
     // Call UpdateUser with correct info
     resp, err = http.Post(*address+"/v1/users/"+created.Id, "application/json", strings.NewReader(fmt.Sprintf(`
          {
            "api":"v1",
            "user": {
              "email": "loola@gmail.com",
              "password": "haha",
              "username": "brewhaha",
              "last_active": 100,
              "experience": "senior",
              "languages": ["haskell", "python", "csharp"]
            }
          }
        `, pfx, pfx, pfx)))
     if err != nil {
       log.Fatalf("failed to call UpdateUser method: %v", err)
     }
     bodyBytes, err = ioutil.ReadAll(resp.Body)
     resp.Body.Close()
     if err != nil {
       body = fmt.Sprintf("failed read UpdateUser response body: %v", err)
     } else {
       body = string(bodyBytes)
     }
     log.Printf("UpdateUser response: Code=%d, Body=%s\n\n", resp.StatusCode, body)

     // parse status of UpdateUser
     var updated struct {
       API      string `json:"api"`
       Status   string `json:"status"`
       Matched  string `json:"matched"`
       Modified string `json:"modified"`
     }
     err = json.Unmarshal(bodyBytes, &updated)
     if err != nil {
       log.Fatalf("failed to unmarshal JSON response of UpdateUser method: %v", err)
       fmt.Println("error:", err)
     }
     log.Printf("updated struct: %s\n", updated)

     // Call FilterUsers
     resp, err = http.Post(*address+"/v1/search", "application/json", strings.NewReader(fmt.Sprintf(`
             {
               "api":"v1",
               "experience": "beginner",
               "page": 1,
               "limit": 20
             }
           `, pfx, pfx, pfx)))
     if err != nil {
       log.Fatalf("failed to call FilterUsers method: %v", err)
     }
     bodyBytes, err = ioutil.ReadAll(resp.Body)
     resp.Body.Close()
     if err != nil {
       body = fmt.Sprintf("failed read FilterUsers response body: %v", err)
     } else {
       body = string(bodyBytes)
     }
     log.Printf("FilterUsers searching for users who know java\n")
     log.Printf("FilterUsers response: Code=%d, Body=%s\n\n", resp.StatusCode, body)

     var users struct {
       LastActive string `json:"last_active"`
       Experience string `json:"experience"`
       Languages  string `json:"languages"`
       Username   string `json:"username"`
     }
     bobby := users
     err = json.Unmarshal(bodyBytes, &bobby)
     if err != nil {
       log.Fatalf("failed to unmarshal JSON response of FilterUsers method: %v", err)
       fmt.Println("error:", err)
     }

     /*
        resp, err = http.Post(*address+"/v1/users/search", "application/json", strings.NewReader(fmt.Sprintf(`
          {
            "api":"v1",
            "experience": "middle",
            "page": 1,
            "limit": 20
          }
        `, pfx, pfx, pfx)))
        if err != nil {
          log.Fatalf("failed to call FilterUsers method: %v", err)
        }
        bodyBytes, err = ioutil.ReadAll(resp.Body)
        resp.Body.Close()
        if err != nil {
          body = fmt.Sprintf("failed read FilterUsers response body: %v", err)
        } else {
          body = string(bodyBytes)
        }
        log.Printf("FilterUsers searching for users with middle experience\n")
        log.Printf("FilterUsers response: Code=%d, Body=%s\n\n", resp.StatusCode, body)
  */
  /*
     // Call GetById
     req, err := http.NewRequest("GET", fmt.Sprintf("%s%s/%s", *address, "/v1/users", created.Id), nil)
     resp, err = http.DefaultClient.Do(req)
     if err != nil {
       log.Fatalf("failed to call GetById method: %v", err)
     }
     bodyBytes, err = ioutil.ReadAll(resp.Body)
     resp.Body.Close()
     if err != nil {
       body = fmt.Sprintf("failed read GetById response body: %v", err)
     } else {
       body = string(bodyBytes)
     }
     log.Printf("GetById response: Code=%d, Body=%s\n\n", resp.StatusCode, body)
  */
  // Call DeleteTeam
  req, err := http.NewRequest("DELETE", fmt.Sprintf("%s%s/%s", *address, "/v1/teams", "1"), nil)
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

  // Call Remove Member
  req, err = http.NewRequest("DELETE", fmt.Sprintf("%s%s/%s/%s/%s", *address, "/v1/teams", "1", "members", "2"), nil)
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

}
