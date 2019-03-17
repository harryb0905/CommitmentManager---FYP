package merchant

import (
  "net/http"
  "github.com/scc300/scc300-network/web/controllers"
)

// Creates a new web server for the merchant application
func InitServer(app *controllers.Application) (*http.ServeMux) {
  server := http.NewServeMux()
  server.HandleFunc("/", app.MerchantHandler)
  return server
}