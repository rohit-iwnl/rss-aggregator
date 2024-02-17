package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/go-chi/chi"
	"github.com/go-chi/cors"
	"github.com/joho/godotenv"
)

func main(){
	godotenv.Load(".env")

	portString := os.Getenv("PORT")
	if(portString == ""){
		log.Fatal("PORT is not set in .env file")
	}
	fmt.Println("PORT is set to", portString);

	router := chi.NewRouter()

	router.Use(cors.Handler(cors.Options{
		AllowedOrigins: []string{"https://*", "http://*"},
		AllowedMethods:  []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders: []string{"*"},
		ExposedHeaders: []string{"Link"},
		AllowCredentials: false,
		MaxAge: 300,
	}))

	v1Router := chi.NewRouter()

	v1Router.Get("/ready", http.HandlerFunc(handlerReadiness))
	v1Router.Get("/error", http.HandlerFunc(handleError))

	router.Mount("/v1", v1Router)

	srv := &http.Server{
		Handler : router,
		Addr: ":" + portString,
	}
	log.Printf("Server is running on port %s", portString)
	err := srv.ListenAndServe()
	if err!=nil{
		log.Fatal(err)
	}
}