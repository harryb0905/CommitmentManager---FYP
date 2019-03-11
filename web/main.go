package web

import (
  "net/http"
  "github.com/scc300/scc300-network/web/controllers"
)

// Create 2 servers using a channel
// 2 apps; merchant app and customer app (listening on different ports)
func StartServers(app *controllers.Application) {
  finish := make(chan bool)
  
  server3000 := http.NewServeMux()
  server3000.HandleFunc("/", app.MerchantHandler)

  server3001 := http.NewServeMux()
  server3001.HandleFunc("/", app.CustomerHandler)

  go func() {
    http.ListenAndServe(":3000", server3000)
  }()

  go func() {
    http.ListenAndServe(":3001", server3001)
  }()

  <-finish
}