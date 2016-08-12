package main

import (
	"html/template"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/jetbasrawi/go.cqrs"
	"github.com/jetbasrawi/go.cqrs/examples/simplecqrs/simplecqrs"
)

var t = template.Must(template.ParseGlob("templates/*"))

func setupHandlers() *http.ServeMux {
	mux := http.NewServeMux()

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		Items := readModel.GetInventoryItems()

		t.ExecuteTemplate(w, "index", Items)
	})

	mux.HandleFunc("/add", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			r.ParseForm()

			id := ycq.NewUUID()
			em := ycq.NewCommandMessage(id, &simplecqrs.CreateInventoryItem{
				Name: r.Form.Get("name"),
			})

			err := dispatcher.Dispatch(em)
			if err != nil {
				log.Println(err)
			}

			http.Redirect(w, r, "/", http.StatusSeeOther)
		}

		t.ExecuteTemplate(w, "add", nil)

	})

	mux.HandleFunc("/details/", func(w http.ResponseWriter, r *http.Request) {
		p := strings.Split(r.URL.Path, "/")
		id := p[len(p)-1]
		Item := readModel.GetInventoryItemDetails(id)
		t.ExecuteTemplate(w, "details", Item)
	})

	mux.HandleFunc("/deactivate/", func(w http.ResponseWriter, r *http.Request) {

		// v := r.URL.Query().Get("version")
		// ver, err := strconv.Atoi(v)
		// if err != nil {
		// 	log.Println(err)
		// }

		p := strings.Split(r.URL.Path, "/")
		id := p[len(p)-1]
		Item := readModel.GetInventoryItemDetails(id)

		if r.Method == http.MethodPost {
			em := ycq.NewCommandMessage(id, &simplecqrs.DeactivateInventoryItem{OriginalVersion: 0})
			err := dispatcher.Dispatch(em)
			if err != nil {
				log.Println(err)
			}

			redirectURL := "/"
			http.Redirect(w, r, redirectURL, http.StatusSeeOther)
		}

		t.ExecuteTemplate(w, "deactivate", Item)
	})

	mux.HandleFunc("/changename/", func(w http.ResponseWriter, r *http.Request) {

		p := strings.Split(r.URL.Path, "/")
		id := p[len(p)-1]
		Item := readModel.GetInventoryItemDetails(id)

		if r.Method == http.MethodPost {
			r.ParseForm()
			em := ycq.NewCommandMessage(id, &simplecqrs.RenameInventoryItem{NewName: r.Form.Get("name")})
			err := dispatcher.Dispatch(em)
			if err != nil {
				log.Println(err)
			}

			redirectURL := "/"
			http.Redirect(w, r, redirectURL, http.StatusSeeOther)
		}

		t.ExecuteTemplate(w, "changename", Item)

	})

	mux.HandleFunc("/checkin/", func(w http.ResponseWriter, r *http.Request) {

		p := strings.Split(r.URL.Path, "/")
		id := p[len(p)-1]
		Item := readModel.GetInventoryItemDetails(id)

		if r.Method == http.MethodPost {
			r.ParseForm()
			num, err := strconv.Atoi(r.Form.Get("number"))
			if err != nil {
				http.Error(w, "Unable to read number.", http.StatusInternalServerError)
			}

			em := ycq.NewCommandMessage(id, &simplecqrs.CheckInItemsToInventory{Count: num})
			err = dispatcher.Dispatch(em)
			if err != nil {
				log.Println(err)
			}

			redirectURL := "/"
			http.Redirect(w, r, redirectURL, http.StatusSeeOther)
		}

		t.ExecuteTemplate(w, "checkin", Item)

	})

	mux.HandleFunc("/remove/", func(w http.ResponseWriter, r *http.Request) {

		p := strings.Split(r.URL.Path, "/")
		id := p[len(p)-1]
		Item := readModel.GetInventoryItemDetails(id)

		if r.Method == http.MethodPost {
			r.ParseForm()
			num, err := strconv.Atoi(r.Form.Get("number"))
			if err != nil {
				http.Error(w, "Unable to read number.", http.StatusInternalServerError)
			}

			em := ycq.NewCommandMessage(id, &simplecqrs.RemoveItemsFromInventory{Count: num})
			err = dispatcher.Dispatch(em)
			if err != nil {
				log.Println(err)
			}
			redirectURL := "/"
			http.Redirect(w, r, redirectURL, http.StatusSeeOther)
		}

		t.ExecuteTemplate(w, "remove", Item)

	})

	mux.HandleFunc("/assets/", func(w http.ResponseWriter, r *http.Request) {
		staticFile := r.URL.Path[len("/assets/"):]
		if len(staticFile) != 0 {
			f, err := http.Dir("assets/").Open(staticFile)
			if err == nil {
				content := io.ReadSeeker(f)
				http.ServeContent(w, r, staticFile, time.Now(), content)
				return
			}
		}
		http.NotFound(w, r)
	})

	return mux
}
