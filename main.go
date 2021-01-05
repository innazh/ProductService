package main

import (
	"BetterProductService/data"
	"BetterProductService/handlers"
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/mongo"
)

/*There are potential issues with the way I handle the db connection if there are multiple routes,
ofc I can create a new client for another one, which might still create race condition for resources on the db but for my intentions, it'll do.
Next: create a file with params*/

func main() {
	var config *Config
	config = ParseConfigFile()
	//mb can move the code below to products file later
	l := log.New(os.Stdout, "product-api: ", log.LstdFlags|log.Lshortfile)
	db := data.GetNewClient(config.Database.ConnStr)
	ph := handlers.NewProducts(l, db)

	sm := mux.NewRouter()
	// sm.PathPrefix("/api/v1/")

	getRouter := sm.Methods("GET").Subrouter()
	getRouter.Handle("/", http.RedirectHandler("/products", 301))
	getRouter.Handle("/products", http.HandlerFunc(ph.GetProducts))
	getRouter.Handle("/products/{id}", http.HandlerFunc(ph.GetProduct))

	postRouter := sm.Methods("POST").Subrouter()
	postRouter.Use(ph.MiddlewareProductValidation)
	postRouter.Handle("/products", http.HandlerFunc(ph.AddProduct))

	putRouter := sm.Methods("PUT").Subrouter()
	putRouter.Use(ph.MiddlewareProductValidation)
	putRouter.Handle("/products/{id}", http.HandlerFunc(ph.UpdateProduct))

	deleteRouter := sm.Methods("DELETE").Subrouter()
	deleteRouter.Handle("/products/{id}", http.HandlerFunc(ph.DeleteProduct))

	//create a custom server to change the timeouts, port & assign the configured router sm to it
	s := &http.Server{
		Addr:         config.Server.Port,
		Handler:      sm,
		IdleTimeout:  120 * time.Second,
		WriteTimeout: 10 * time.Second,
		ReadTimeout:  10 * time.Second,
	}

	go func() {
		err := s.ListenAndServe()
		if err != nil {
			l.Fatal(err)
		}
	}()

	serverShutdown(l, s, db)
}

//close everything donw upon receiving a shut down command
func serverShutdown(l *log.Logger, s *http.Server, db *mongo.Client) {
	sigChannel := make(chan os.Signal)
	signal.Notify(sigChannel, os.Interrupt)
	signal.Notify(sigChannel, os.Kill)

	sig := <-sigChannel

	//l.Println("Disconnecting from the db", sig)
	//err := db.Disconnect(context.Background())
	//l.Println("Error disconnecting from the db:" + err.Error())

	l.Println("Shutting the server down->", sig)
	ctx, _ := context.WithTimeout(context.Background(), time.Second*30)
	s.Shutdown(ctx)
}
