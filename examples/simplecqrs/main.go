package main

import (
	"html/template"
	"io"
	"log"
	"net/http"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/jetbasrawi/go.cqrs"
	"github.com/jetbasrawi/go.cqrs/examples/simplecqrs/simplecqrs"
	"github.com/jetbasrawi/go.geteventstore"
)

var (
	readModel  simplecqrs.ReadModelFacade
	dispatcher ycq.Dispatcher
)

func setupHandlers() *http.ServeMux {
	p, err := filepath.Abs(path.Join("templates", "*"))
	if err != nil {
		log.Fatal(err)
	}
	t := template.Must(template.ParseGlob(p))

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

	// Create an in memory repository
	//repo := simplecqrs.NewInMemoryRepo(eventBus)
	client, err := goes.NewClient(nil, "http://localhost:2113")
	if err != nil {
		log.Fatal(err)
	}
	repo, err := simplecqrs.NewInventoryItemRepo(client, eventBus)
	if err != nil {
		log.Fatal(err)
	}

	// Create an InventoryCommandHandlers instance
	inventoryCommandHandler := simplecqrs.NewInventoryCommandHandlers(repo)

	// Create a dispatcher
	dispatcher = ycq.NewInMemoryDispatcher()
	// Register the inventory command handlers instance as a command handler
	// for the events specified.
	dispatcher.RegisterHandler(inventoryCommandHandler,
		&simplecqrs.CreateInventoryItem{},
		&simplecqrs.DeactivateInventoryItem{},
		&simplecqrs.RenameInventoryItem{},
		&simplecqrs.CheckInItemsToInventory{},
		&simplecqrs.RemoveItemsFromInventory{},
	)

}

func main() {

	mux := setupHandlers()
	if err := http.ListenAndServe(":8088", mux); err != nil {
		log.Fatal(err)
	}

}
