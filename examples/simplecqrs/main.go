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

var (
	readModel  simplecqrs.ReadModelFacade
	dispatcher ycq.Dispatcher
)

var t = template.Must(template.ParseGlob("templates/*"))

func init() {

	// CQRS Infrastructure configuration

	// Configure the read model

	// Create a readModel instance
	readModel = simplecqrs.NewReadModel()

	// Create a InventoryListView
	listView := simplecqrs.NewInventoryListView()
	// Create a InventoryItemDetailView
	detailView := simplecqrs.NewInventoryItemDetailView()

	// Create an EventBus
	eventBus := ycq.NewInternalEventBus()
	// Register the listView as an event handler on the event bus
	// for the events specified.
	eventBus.AddHandler(listView,
		&simplecqrs.InventoryItemCreated{},
		&simplecqrs.InventoryItemRenamed{},
		&simplecqrs.InventoryItemDeactivated{},
	)
	// Register the detail view as an event handler on the event bus
	// for the events specified.
	eventBus.AddHandler(detailView,
		&simplecqrs.InventoryItemCreated{},
		&simplecqrs.InventoryItemRenamed{},
		&simplecqrs.InventoryItemDeactivated{},
		&simplecqrs.ItemsRemovedFromInventory{},
		&simplecqrs.ItemsCheckedIntoInventory{},
	)

	// Here we use an in memory event repository.
	repo := simplecqrs.NewInMemoryRepo(eventBus)

	// Here we use geteventstore with the go.geteventstore client
	// https://github.com/jetbasrawi/go.geteventstore
	// Uncomment the following code and comment out the previous in memory repository
	// to use geteventstore

	//client, err := goes.NewClient(nil, "http://localhost:2113")
	//if err != nil {
	//	log.Fatal(err)
	//}
	//repo, err := simplecqrs.NewInventoryItemRepo(client, eventBus)
	//if err != nil {
	//	log.Fatal(err)
	//}

	// Create an InventoryCommandHandlers instance
	inventoryCommandHandler := simplecqrs.NewInventoryCommandHandlers(repo)

	// Create a dispatcher
	dispatcher = ycq.NewInMemoryDispatcher()
	// Register the inventory command handlers instance as a command handler
	// for the events specified.
	err := dispatcher.RegisterHandler(inventoryCommandHandler,
		&simplecqrs.CreateInventoryItem{},
		&simplecqrs.DeactivateInventoryItem{},
		&simplecqrs.RenameInventoryItem{},
		&simplecqrs.CheckInItemsToInventory{},
		&simplecqrs.RemoveItemsFromInventory{},
	)
	if err != nil {
		log.Fatal(err)
	}
}

func main() {

	mux := setupHandlers()
	if err := http.ListenAndServe(":8088", mux); err != nil {
		log.Fatal(err)
	}

}

func setupHandlers() *http.ServeMux {
	mux := http.NewServeMux()

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		Items := readModel.GetInventoryItems()

		err := t.ExecuteTemplate(w, "index", Items)
		if err != nil {
			log.Fatal(err)
		}
	})

	mux.HandleFunc("/add", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			err := r.ParseForm()
			if err != nil {
				log.Fatal(err)
			}

			id := ycq.NewUUID()
			em := ycq.NewCommandMessage(id, &simplecqrs.CreateInventoryItem{
				Name: r.Form.Get("name"),
			})

			err = dispatcher.Dispatch(em)
			if err != nil {
				log.Println(err)
			}

			http.Redirect(w, r, "/", http.StatusSeeOther)
		}

		err := t.ExecuteTemplate(w, "add", nil)
		if err != nil {
			log.Fatal(err)
		}
	})

	mux.HandleFunc("/details/", func(w http.ResponseWriter, r *http.Request) {
		p := strings.Split(r.URL.Path, "/")
		id := p[len(p)-1]
		Item := readModel.GetInventoryItemDetails(id)
		err := t.ExecuteTemplate(w, "details", Item)
		if err != nil {
			log.Fatal(err)
		}
	})

	mux.HandleFunc("/deactivate/", func(w http.ResponseWriter, r *http.Request) {

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

		err := t.ExecuteTemplate(w, "deactivate", Item)
		if err != nil {
			log.Fatal(err)
		}
	})

	mux.HandleFunc("/changename/", func(w http.ResponseWriter, r *http.Request) {

		p := strings.Split(r.URL.Path, "/")
		id := p[len(p)-1]
		Item := readModel.GetInventoryItemDetails(id)

		if r.Method == http.MethodPost {
			err := r.ParseForm()
			if err != nil {
				log.Fatal(err)
			}

			em := ycq.NewCommandMessage(id, &simplecqrs.RenameInventoryItem{NewName: r.Form.Get("name")})
			err = dispatcher.Dispatch(em)
			if err != nil {
				log.Println(err)
			}

			redirectURL := "/"
			http.Redirect(w, r, redirectURL, http.StatusSeeOther)
		}

		err := t.ExecuteTemplate(w, "changename", Item)
		if err != nil {
			log.Fatal(err)
		}

	})

	mux.HandleFunc("/checkin/", func(w http.ResponseWriter, r *http.Request) {

		p := strings.Split(r.URL.Path, "/")
		id := p[len(p)-1]
		Item := readModel.GetInventoryItemDetails(id)

		if r.Method == http.MethodPost {
			err := r.ParseForm()
			if err != nil {
				log.Fatal(err)
			}

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

		err := t.ExecuteTemplate(w, "checkin", Item)
		if err != nil {
			log.Fatal(err)
		}

	})

	mux.HandleFunc("/remove/", func(w http.ResponseWriter, r *http.Request) {

		p := strings.Split(r.URL.Path, "/")
		id := p[len(p)-1]
		Item := readModel.GetInventoryItemDetails(id)

		if r.Method == http.MethodPost {
			err := r.ParseForm()
			if err != nil {
				log.Fatal(err)
			}

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

		err := t.ExecuteTemplate(w, "remove", Item)
		if err != nil {
			log.Fatal(err)
		}
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

