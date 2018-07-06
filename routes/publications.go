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

func handlePublicationActions(r *http.Request, id string, action string) (interface{}, error) {
	switch r.Method {
	case "GET":
		switch action {
		case "headlines":
			val, included, count, total, err := controllers.GetHeadlinesForPublication(r, id)
			return api.BaseResponseHandler(val, included, count, total, err, r)
		case "database-profile":
			return api.BaseSingleResponseHandler(controllers.GetEnrichCompanyProfile(r, id))
		case "verify":
			return api.BaseSingleResponseHandler(controllers.VerifyPublication(r, id))
		}
	}
	return nil, errors.New("method not implemented")
}

func handlePublication(r *http.Request, id string) (interface{}, error) {
	switch r.Method {
	case "GET":
		return api.BaseSingleResponseHandler(controllers.GetPublication(id))
	case "PATCH":
		return api.BaseSingleResponseHandler(controllers.UpdatePublication(r, id))
	}
	return nil, errors.New("method not implemented")
}

func handlePublications(w http.ResponseWriter, r *http.Request) (interface{}, error) {
	switch r.Method {
	case "GET":
		if len(r.URL.Query()) > 0 {
			if val, ok := r.URL.Query()["name"]; ok && len(val) > 0 {
				return api.BaseSingleResponseHandler(controllers.FilterPublicationByNameAndUrl(val[0], ""))
			}
		}
		val, included, count, total, err := controllers.GetPublications(r)
		return api.BaseResponseHandler(val, included, count, total, err, r)
	case "POST":
		val, included, count, total, err := controllers.CreatePublication(w, r)
		if count == 1 {
			return api.BaseSingleResponseHandler(val, included, err)
		}
		return api.BaseResponseHandler(val, included, count, total, err, r)
	}
	return nil, errors.New("method not implemented")
}

// Handler for when the user wants all the agencies.
func PublicationsHandler(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	w.Header().Set("Content-Type", "application/json")
	val, err := handlePublications(w, r)

	if err == nil {
		err = ffjson.NewEncoder(w).Encode(val)
	}

	if err != nil {
		nError.ReturnError(w, http.StatusInternalServerError, "Publication handling error", err.Error())
	}
	return
}

// Handler for when there is a key present after /users/<id> route.
func PublicationHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	w.Header().Set("Content-Type", "application/json")
	id := ps.ByName("id")
	val, err := handlePublication(r, id)

	if err == nil {
		err = ffjson.NewEncoder(w).Encode(val)
	}

	if err != nil {
		nError.ReturnError(w, http.StatusInternalServerError, "Publication handling error", err.Error())
	}
	return
}

// Handler for when the user wants to perform an action on the publications
func PublicationActionHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	w.Header().Set("Content-Type", "application/json")
	id := ps.ByName("id")
	action := ps.ByName("action")

	val, err := handlePublicationActions(r, id, action)
	if err == nil {
		err = ffjson.NewEncoder(w).Encode(val)
	}

	if err != nil {
		nError.ReturnError(w, http.StatusInternalServerError, "Publication handling error", err.Error())
	}
	return
}
