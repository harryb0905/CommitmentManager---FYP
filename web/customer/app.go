package customer

import (
  "fmt"
  "net/http"
  "github.com/scc300/scc300-network/web/controllers"
)

// Creates a new web server for the customer application
func Serve(app *controllers.Application) {
  http.HandleFunc("/customer", app.HomeHandler)
  fmt.Println("Customer application listening (http://localhost:3000/) ...")
  http.ListenAndServe(":3000", nil)
}