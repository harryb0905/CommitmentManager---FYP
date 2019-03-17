package customer

import (
  "net/http"
  "github.com/scc300/scc300-network/web/controllers"
)

// Creates a new web server for the customer application
func InitServer(app *controllers.Application) (*http.ServeMux) {
  server := http.NewServeMux()
  server.HandleFunc("/", app.CustomerHandler)
  return server
}