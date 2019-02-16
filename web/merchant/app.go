package merchant

import (
  "fmt"
  "net/http"
  "github.com/scc300/scc300-network/web/controllers"
)

// Creates a new web server for the merchant application
func Serve(app *controllers.Application) {
  http.HandleFunc("/merchant", app.MerchantHandler)
  fmt.Println("Merchant application listening (http://localhost:3001/) ...")
  http.ListenAndServe(":3001", nil)
}