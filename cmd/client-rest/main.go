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
  address := flag.String("server", "http://localhost:8080", "HTTP gateway url, e.g. http://localhost:8080")
  flag.Parse()

  t := time.Now().In(time.UTC)
  pfx := t.Format(time.RFC3339Nano)

  var body string

}
