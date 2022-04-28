package main

import (
	_ "embed"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/patrickdappollonio/redis-todo-list/storage"
)

//go:embed template.html
var frontend string

var (
	key  = strings.ToLower(envdefault("KEY", "guestbook"))
	port = ":" + envdefault("PORT", "80")
)

func main() {
	var client storage.Storage
	var clienterr error

	switch {
	case allSet("REDIS_HOST", "REDIS_PASS"):
		client, clienterr = storage.NewRedis(
			os.Getenv("REDIS_HOST"),
			os.Getenv("REDIS_PASS"),
			os.Getenv("REDIS_SERVER_NAME"),
			boolEnv("REDIS_USE_TLS", false),
		)
	case allSet("SQL_USERNAME", "SQL_PASSWORD", "SQL_HOST", "SQL_DATABASE"):
		client, clienterr = storage.NewMSSQL(
			os.Getenv("SQL_USERNAME"),
			os.Getenv("SQL_PASSWORD"),
			os.Getenv("SQL_HOST"),
			os.Getenv("SQL_INSTANCE"),
			os.Getenv("SQL_DATABASE"),
		)
	default:
		log.Fatal("Error: No storage backend configured")
	}

	if clienterr != nil {
		log.Fatalf("Error while setting up client: %s", clienterr)
	}

	if err := client.IsValidKey(key); err != nil {
		log.Fatalf("Error on value for $KEY: %s", err)
	}

	if err := client.Bootstrap(key); err != nil {
		log.Fatalf("Error while bootstrapping client: %s", err)
	}

	log.Printf("Client boostrapped -- Backend configuration: %s", client.ConfigString())

	r := chi.NewRouter()
	r.Use(middleware.Logger)

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write(strToBytes(frontend))
	})

	r.Get("/guestbook", func(w http.ResponseWriter, r *http.Request) {
		var (
			cmd   = r.URL.Query().Get("cmd")
			value = r.URL.Query().Get("value")
		)

		switch cmd {
		case "set":
			if err := client.Set(key, value); err != nil {
				log.Printf("Error: %s", err)
				httpError(w, http.StatusInternalServerError, "%s", err)
				return
			}

		case "get":
			val, err := client.Get(key)
			if err != nil {
				log.Printf("Error: %s", err)
				httpError(w, http.StatusInternalServerError, "%s", err)
				return
			}

			asJSON(w, map[string]interface{}{"data": val})

		default:
			httpError(w, http.StatusBadRequest, "invalid command: %s", cmd)
		}
	})

	log.Println("Starting HTTP server on port", port)
	log.Fatal(http.ListenAndServe(port, r))
}
