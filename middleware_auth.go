package main

import (
	"fmt"
	"net/http"

	"github.com/rohit-iwnl/rss-aggregator/internal/auth"
	"github.com/rohit-iwnl/rss-aggregator/internal/database"
)

type authedHandler func(http.ResponseWriter, *http.Request, database.User)

func (cfg *apiConfig) middleWareAuth(handler authedHandler) http.HandlerFunc{
	return func(w http.ResponseWriter, r *http.Request){
		apiKey, err := auth.GetApiKey(r.Header)
		if err != nil {
			respondWithError(w,403,fmt.Sprintf("Auth Error: %v", err))
			return
		}

		user, err := cfg.DB.GetUserByApiKey(r.Context(),apiKey)
		if err != nil {
			respondWithError(w,403,fmt.Sprintf("Couldn't get user: %v", err))
			return
		}

		handler(w,r,user)
	}
}