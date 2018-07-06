package routes

// import (
// 	"errors"
// 	"net/http"

// 	"github.com/julienschmidt/httprouter"
// 	"github.com/pquerna/ffjson/ffjson"

// 	"github.com/news-ai/tabulae-v1/controllers"

// 	"github.com/news-ai/web/api"
// 	nError "github.com/news-ai/web/errors"
// )

// func handleContactAction(r *http.Request, id string, action string) (interface{}, error) {
// 	switch r.Method {
// 	case "GET":
// 		switch action {
// 		case "feed":
// 			val, included, count, total, err := controllers.GetFeedForContact(r, id)
// 			return api.BaseResponseHandler(val, included, count, total, err, r)
// 		case "headlines":
// 			val, included, count, total, err := controllers.GetHeadlinesForContact(r, id)
// 			return api.BaseResponseHandler(val, included, count, total, err, r)
// 		case "tweets":
// 			val, included, count, total, err := controllers.GetTweetsForContact(r, id)
// 			return api.BaseResponseHandler(val, included, count, total, err, r)
// 		case "twitterprofile":
// 			return api.BaseSingleResponseHandler(controllers.GetTwitterProfileForContact(r, id))
// 		case "twittertimeseries":
// 			return api.BaseSingleResponseHandler(controllers.GetTwitterTimeseriesForContact(r, id))
// 		case "instagrams":
// 			val, included, count, total, err := controllers.GetInstagramPostsForContact(r, id)
// 			return api.BaseResponseHandler(val, included, count, total, err, r)
// 		case "instagramprofile":
// 			return api.BaseSingleResponseHandler(controllers.GetInstagramProfileForContact(r, id))
// 		case "instagramtimeseries":
// 			return api.BaseSingleResponseHandler(controllers.GetInstagramTimeseriesForContact(r, id))
// 		case "feeds":
// 			val, included, count, total, err := controllers.GetFeedsForContact(r, id)
// 			return api.BaseResponseHandler(val, included, count, total, err, r)
// 		case "emails":
// 			val, included, count, total, err := controllers.GetEmailsForContact(r, id)
// 			return api.BaseResponseHandler(val, included, count, total, err, r)
// 		case "lists":
// 			val, included, count, total, err := controllers.GetListsForContact(r, id)
// 			return api.BaseResponseHandler(val, included, count, total, err, r)
// 		case "enrich":
// 			return api.BaseSingleResponseHandler(controllers.EnrichContact(r, id))
// 		}
// 	}
// 	return nil, errors.New("method not implemented")
// }

// func handleContact(r *http.Request, id string) (interface{}, error) {
// 	switch r.Method {
// 	case "GET":
// 		return api.BaseSingleResponseHandler(controllers.GetContact(r, id))
// 	case "PATCH":
// 		return api.BaseSingleResponseHandler(controllers.UpdateSingleContact(r, id))
// 	case "POST":
// 		if id == "copy" {
// 			val, included, count, total, err := controllers.CopyContacts(r)
// 			return api.BaseResponseHandler(val, included, count, total, err, r)
// 		} else if id == "bulkdelete" {
// 			val, included, count, total, err := controllers.BulkDeleteContacts(r)
// 			return api.BaseResponseHandler(val, included, count, total, err, r)
// 		}
// 	case "DELETE":
// 		return api.BaseSingleResponseHandler(controllers.DeleteContact(r, id))
// 	}
// 	return nil, errors.New("method not implemented")
// }

// func handleContacts(w http.ResponseWriter, r *http.Request) (interface{}, error) {
// 	switch r.Method {
// 	case "GET":
// 		val, included, count, total, err := controllers.GetContacts(r)
// 		return api.BaseResponseHandler(val, included, count, total, err, r)
// 	case "POST":
// 		val, included, count, total, err := controllers.CreateContact(r)
// 		return api.BaseResponseHandler(val, included, count, total, err, r)
// 	case "PATCH":
// 		val, included, count, total, err := controllers.UpdateBatchContact(r)
// 		return api.BaseResponseHandler(val, included, count, total, err, r)
// 	}
// 	return nil, errors.New("method not implemented")
// }

// // Handler for when the user wants all the contacts.
// func ContactsHandler(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
// 	w.Header().Set("Content-Type", "application/json")
// 	val, err := handleContacts(w, r)

// 	if err == nil {
// 		err = ffjson.NewEncoder(w).Encode(val)
// 	}

// 	if err != nil {
// 		nError.ReturnError(w, http.StatusInternalServerError, "Contact handling error", err.Error())
// 	}
// 	return
// }

// // Handler for when there is a key present after /users/<id> route.
// func ContactHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
// 	w.Header().Set("Content-Type", "application/json")
// 	id := ps.ByName("id")
// 	val, err := handleContact(r, id)

// 	if err == nil {
// 		err = ffjson.NewEncoder(w).Encode(val)
// 	}

// 	if err != nil {
// 		nError.ReturnError(w, http.StatusInternalServerError, "Contact handling error", err.Error())
// 	}
// 	return
// }

// func ContactActionHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
// 	w.Header().Set("Content-Type", "application/json")
// 	id := ps.ByName("id")
// 	action := ps.ByName("action")
// 	val, err := handleContactAction(r, id, action)

// 	if err == nil {
// 		err = ffjson.NewEncoder(w).Encode(val)
// 	}

// 	if err != nil {
// 		nError.ReturnError(w, http.StatusInternalServerError, "Contact handling error", err.Error())
// 	}
// 	return
// }
