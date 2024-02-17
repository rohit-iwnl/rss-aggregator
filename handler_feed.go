package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi"
	"github.com/google/uuid"
	"github.com/rohit-iwnl/rss-aggregator/internal/database"
)

func (apiCfg *apiConfig) handlerCreateFeed(w http.ResponseWriter, r *http.Request, user database.User){
	type parameters struct {
		Name string `json:"name"`
		URL string `json:"url"`
	}

	decoder := json.NewDecoder(r.Body)

	params := parameters{}
	err := decoder.Decode(&params)
	if err != nil {
		respondWithError(w, 400, fmt.Sprintf("Invalid request payload: %v", err))
		return
	}

	feed, err := apiCfg.DB.CreateFeed(r.Context(), database.CreateFeedParams{
		ID:        uuid.New(),
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
		Name:      params.Name,
		Url:       params.URL,
		UserID:    user.ID,
	})

	if err != nil {
		respondWithError(w, 400, fmt.Sprintf("Cannot create feed: %v", err))
		return
	}

	respondWithJSON(w,201,databaseFeedToFeed(feed))
}

func (apiCfg *apiConfig) handlerGetFeeds(w http.ResponseWriter, r *http.Request){
	feeds, err := apiCfg.DB.GetFeeds(r.Context())
	if err!=nil{
		respondWithError(w,403,fmt.Sprintf("Cannot get feeds: %v", err))
		return
	}

	respondWithJSON(w,201,databaseFeedsToFeedsList(feeds))
}


func (apiCfg *apiConfig) handlerDeleteFeedsFollow(w http.ResponseWriter, r *http.Request,user database.User){
	feedFollowIdString := chi.URLParam(r, "feedFollowID")

	feedFollowID, err := uuid.Parse(feedFollowIdString)
	if err != nil {
		respondWithError(w, 400, fmt.Sprintf("Invalid feed follow id: %v", err))
		return
	}

	err = apiCfg.DB.DeleteFeedFollows(r.Context(), database.DeleteFeedFollowsParams{
		ID: feedFollowID,
		UserID: user.ID,
	})
	if err != nil {
		respondWithError(w, 400, fmt.Sprintf("Cannot delete feed follow: %v", err))
		return
	}
	respondWithJSON(w,200,struct{}{})
}