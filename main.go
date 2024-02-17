package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/go-chi/chi"
	"github.com/go-chi/cors"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"github.com/rohit-iwnl/rss-aggregator/internal/database"
)

type apiConfig struct{
	DB *database.Queries
}

func main(){
	godotenv.Load(".env")

	portString := os.Getenv("PORT")
	if(portString == ""){
		log.Fatal("PORT is not set in .env file")
	}
	fmt.Println("PORT is set to", portString);

	dbUrl := os.Getenv("DB_URL")
	if dbUrl == "" {
		log.Fatal("DATABASE_URL is not set in .env file")
	}

	conn,err := sql.Open("postgres", dbUrl)

	if err!=nil{
		log.Fatal("Cannot connect to database", err)
	}

	db := database.New(conn)
	
	apiCfg := apiConfig{
		DB : db,
	}

	go startScraping(db,10,time.Minute)

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
	v1Router.Post("/users", apiCfg.handlerCreateUser)
	v1Router.Get("/users", apiCfg.middleWareAuth(apiCfg.handlerGetUser))
	v1Router.Post("/feeds", apiCfg.middleWareAuth(apiCfg.handlerCreateFeed))
	v1Router.Get("/feeds",apiCfg.handlerGetFeeds)
	v1Router.Post("/feed_follows", apiCfg.middleWareAuth(apiCfg.handlerCreateFeedFollow))
	v1Router.Get("/feed_follows",apiCfg.middleWareAuth(apiCfg.handlerGetFeedFollows))
	v1Router.Delete("/feed_follows/{feedFollowID}", apiCfg.middleWareAuth(apiCfg.handlerDeleteFeedsFollow))
	v1Router.Get("/posts", apiCfg.middleWareAuth(apiCfg.handlerGetPostsForUser))

	router.Mount("/v1", v1Router)

	srv := &http.Server{
		Handler : router,
		Addr: ":" + portString,
	}
	log.Printf("Server is running on port %s", portString)
	err = srv.ListenAndServe()
	if err!=nil{
		log.Fatal(err)
	}
}