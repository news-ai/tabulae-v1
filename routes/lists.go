package routes

// import (
// 	"errors"
// 	"net/http"

// 	"github.com/julienschmidt/httprouter"
// 	"github.com/pquerna/ffjson/ffjson"

// 	"github.com/news-ai/tabulae-v1/controllers"
// 	"github.com/news-ai/tabulae-v1/files"

// 	"github.com/news-ai/web/api"
// 	nError "github.com/news-ai/web/errors"
// )

// var (
// 	errMediaListHandling = "Media List handling error"
// )

// func handleMediaListActions(c context.Context, r *http.Request, id string, action string) (interface{}, error) {
// 	switch r.Method {
// 	case "GET":
// 		switch action {
// 		case "contacts":
// 			val, included, count, total, err := controllers.GetContactsForList(c, r, id)
// 			return api.BaseResponseHandler(val, included, count, total, err, r)
// 		case "headlines":
// 			val, included, count, total, err := controllers.GetHeadlinesForList(c, r, id)
// 			return api.BaseResponseHandler(val, included, count, total, err, r)
// 		case "tweets":
// 			val, included, count, total, err := controllers.GetTweetsForList(c, r, id)
// 			return api.BaseResponseHandler(val, included, count, total, err, r)
// 		case "feed":
// 			val, included, count, total, err := controllers.GetFeedForList(c, r, id)
// 			return api.BaseResponseHandler(val, included, count, total, err, r)
// 		case "emails":
// 			val, included, count, total, err := controllers.GetEmailsForList(c, r, id)
// 			return api.BaseResponseHandler(val, included, count, total, err, r)
// 		case "public":
// 			return api.BaseSingleResponseHandler(controllers.UpdateMediaListToPublic(c, r, id))
// 		case "resync":
// 			return api.BaseSingleResponseHandler(controllers.ReSyncMediaList(c, r, id))
// 		}
// 	case "POST":
// 		switch action {
// 		case "upload":
// 			return api.BaseSingleResponseHandler(files.HandleMediaListActionUpload(c, r, id))
// 		case "twittertimeseries":
// 			return api.BaseSingleResponseHandler(controllers.GetTwitterTimeseriesForList(c, r, id))
// 		case "instagramtimeseries":
// 			return api.BaseSingleResponseHandler(controllers.GetInstagramTimeseriesForList(c, r, id))
// 		case "duplicate":
// 			return api.BaseSingleResponseHandler(controllers.DuplicateList(c, r, id))
// 		}
// 	}
// 	return nil, errors.New("method not implemented")
// }

// func handleMediaList(c context.Context, r *http.Request, id string) (interface{}, error) {
// 	switch r.Method {
// 	case "GET":
// 		if id == "archived" {
// 			val, included, count, total, err := controllers.GetMediaLists(c, r, true)
// 			return api.BaseResponseHandler(val, included, count, total, err, r)
// 		} else if id == "clients" {
// 			val, included, count, total, err := controllers.GetMediaListsClients(c, r)
// 			return api.BaseResponseHandler(val, included, count, total, err, r)
// 		} else if id == "public" {
// 			val, included, count, total, err := controllers.GetPublicMediaLists(c, r)
// 			return api.BaseResponseHandler(val, included, count, total, err, r)
// 		} else if id == "team" {
// 			val, included, count, total, err := controllers.GetTeamMediaLists(c, r)
// 			return api.BaseResponseHandler(val, included, count, total, err, r)
// 		}
// 		return api.BaseSingleResponseHandler(controllers.GetMediaList(c, r, id))
// 	case "PATCH":
// 		return api.BaseSingleResponseHandler(controllers.UpdateMediaList(c, r, id))
// 	case "DELETE":
// 		return api.BaseSingleResponseHandler(controllers.DeleteMediaList(c, r, id))
// 	}
// 	return nil, errors.New("method not implemented")
// }

// func handleMediaLists(c context.Context, w http.ResponseWriter, r *http.Request) (interface{}, error) {
// 	switch r.Method {
// 	case "GET":
// 		val, included, count, total, err := controllers.GetMediaLists(c, r, false)
// 		return api.BaseResponseHandler(val, included, count, total, err, r)
// 	case "POST":
// 		return api.BaseSingleResponseHandler(controllers.CreateMediaList(c, w, r))
// 	}
// 	return nil, errors.New("method not implemented")
// }

// // Handler for when the user wants all the agencies.
// func MediaListsHandler(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
// 	w.Header().Set("Content-Type", "application/json")
// 	c := appengine.NewContext(r)
// 	val, err := handleMediaLists(c, w, r)

// 	if err == nil {
// 		err = ffjson.NewEncoder(w).Encode(val)
// 	}

// 	if err != nil {
// 		nError.ReturnError(w, http.StatusInternalServerError, errMediaListHandling, err.Error())
// 	}
// 	return
// }

// // Handler for when there is a key present after /users/<id> route.
// func MediaListHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
// 	w.Header().Set("Content-Type", "application/json")
// 	c := appengine.NewContext(r)
// 	id := ps.ByName("id")
// 	val, err := handleMediaList(c, r, id)

// 	if err == nil {
// 		err = ffjson.NewEncoder(w).Encode(val)
// 	}

// 	if err != nil {
// 		nError.ReturnError(w, http.StatusInternalServerError, errMediaListHandling, err.Error())
// 	}
// 	return
// }

// // Handler for when the user wants to perform an action on the lists
// func MediaListActionHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
// 	w.Header().Set("Content-Type", "application/json")
// 	c := appengine.NewContext(r)
// 	id := ps.ByName("id")
// 	action := ps.ByName("action")

// 	val, err := handleMediaListActions(c, r, id, action)
// 	if err == nil {
// 		err = ffjson.NewEncoder(w).Encode(val)
// 	}

// 	if err != nil {
// 		nError.ReturnError(w, http.StatusInternalServerError, errMediaListHandling, err.Error())
// 	}
// 	return
// }
