package controllers

import (
	// "fmt"
	"net/http"
	// q "github.com/chainHero/heroes-service/quark"
)

func (app *Application) EditorHandler(w http.ResponseWriter, r *http.Request) {
	data := &struct {
		TransactionId string
		Success       bool
		Response      bool
	}{
		TransactionId: "",
		Success:       false,
		Response:      false,
	}
	// if r.FormValue("submitted") == "true" {
	// 	// This will search for commitment with that name, get the source and return
	// 	// Parse the returned result, get first relation name (Offer)
	// 	// Perform below with that found first relation
	// 	// Output results to HTML
	// 	created, err := q.GetCreatedCommitments("SellItem", app.Fabric);
	// 	if err != nil {
  //     fmt.Println(err)
  //   } else {
  //     for _, res := range created {
  //       fmt.Println("CREATED:", res.Value)
  //     }
  //   }
	//
	// 	// helloValue := r.FormValue("hello")
	// 	// txid, err := app.Fabric.InvokeHello(helloValue)
	// 	// if err != nil {
	// 	// 	http.Error(w, "Unable to invoke hello in the blockchain", 500)
	// 	// }
	// 	// data.TransactionId = txid
	// 	// data.Success = true
	// 	// data.Response = true
	// }
	renderTemplate(w, r, "editor.html", data)
}
