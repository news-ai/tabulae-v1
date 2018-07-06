package routes

import (
	"errors"
	"net/http"

	"github.com/julienschmidt/httprouter"
	"github.com/pquerna/ffjson/ffjson"

	"github.com/news-ai/tabulae-v1/controllers"

	"github.com/news-ai/web/api"
	nError "github.com/news-ai/web/errors"
)

func handleFeed(r *http.Request, id string) (interface{}, error) {
	switch r.Method {
	case "GET":
		return api.BaseSingleResponseHandler(controllers.GetFeed(r, id))
	case "DELETE":
		return api.BaseSingleResponseHandler(controllers.DeleteFeed(r, id))
	}
	return nil, errors.New("method not implemented")
}

func handleFeeds(w http.ResponseWriter, r *http.Request) (interface{}, error) {
	switch r.Method {
	case "GET":
		val, included, count, total, err := controllers.GetFeeds(r)
		return api.BaseResponseHandler(val, included, count, total, err, r)
	case "POST":
		return api.BaseSingleResponseHandler(controllers.CreateFeed(r))
	}
	return nil, errors.New("method not implemented")
}

// Handler for when the user wants all the contacts.
func FeedsHandler(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	w.Header().Set("Content-Type", "application/json")
	val, err := handleFeeds(w, r)

	if err == nil {
		err = ffjson.NewEncoder(w).Encode(val)
	}

	if err != nil {
		nError.ReturnError(w, http.StatusInternalServerError, "Feed handling error", err.Error())
	}
	return
}

// Handler for when there is a key present after /users/<id> route.
func FeedHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	w.Header().Set("Content-Type", "application/json")
	id := ps.ByName("id")
	val, err := handleFeed(r, id)

	if err == nil {
		err = ffjson.NewEncoder(w).Encode(val)
	}

	if err != nil {
		nError.ReturnError(w, http.StatusInternalServerError, "Feed handling error", err.Error())
	}
	return
}
