package web

import (
  "net/http"
  "github.com/scc300/scc300-network/web/controllers"
  merchant "github.com/scc300/scc300-network/web/merchant"
  customer "github.com/scc300/scc300-network/web/customer"
)

// Create 2 servers using a channel
// 2 apps; merchant app and customer app (listening on different ports)
func StartServers(app *controllers.Application) {
  finish := make(chan bool)

  go func() {
    http.ListenAndServe(":3000", merchant.InitServer(app))
  }()

  go func() {
    http.ListenAndServe(":3001", customer.InitServer(app))
  }()

  <-finish
}