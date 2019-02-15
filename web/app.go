package web

import (
	"fmt"
	"net/http"
	"github.com/scc300/scc300-network/web/controllers"
)

func Serve(app *controllers.Application) {
	http.HandleFunc("/home.html", app.HomeHandler)
	http.HandleFunc("/editor.html", app.EditorHandler)

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/home.html", http.StatusTemporaryRedirect)
	})

	fmt.Println("Listening (http://localhost:3000/) ...")
	http.ListenAndServe(":3000", nil)
}