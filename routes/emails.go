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

// func handleEmailAction(c context.Context, r *http.Request, id string, action string) (interface{}, error) {
// 	switch r.Method {
// 	case "GET":
// 		switch action {
// 		case "send":
// 			return api.BaseSingleResponseHandler(controllers.SendEmail(c, r, id))
// 		case "cancel":
// 			return api.BaseSingleResponseHandler(controllers.CancelEmail(c, r, id))
// 		case "archive":
// 			return api.BaseSingleResponseHandler(controllers.ArchiveEmail(c, r, id))
// 		case "logs":
// 			return api.BaseSingleResponseHandler(controllers.GetEmailLogs(c, r, id))
// 		}
// 	case "POST":
// 		switch action {
// 		case "attach":
// 			return api.BaseSingleResponseHandler(files.HandleEmailAttachActionUpload(c, r, id))
// 		}
// 	}
// 	return nil, errors.New("method not implemented")
// }

// func handleEmail(c context.Context, r *http.Request, id string) (interface{}, error) {
// 	switch r.Method {
// 	case "GET":
// 		if id == "team" {
// 			val, included, count, total, err := controllers.GetTeamEmails(c, r)
// 			return api.BaseResponseHandler(val, included, count, total, err, r)
// 		} else if id == "scheduled" {
// 			val, included, count, total, err := controllers.GetScheduledEmails(c, r)
// 			return api.BaseResponseHandler(val, included, count, total, err, r)
// 		} else if id == "archived" {
// 			val, included, count, total, err := controllers.GetArchivedEmails(c, r)
// 			return api.BaseResponseHandler(val, included, count, total, err, r)
// 		} else if id == "cancelscheduled" {
// 			val, included, count, total, err := controllers.CancelAllScheduled(c, r)
// 			return api.BaseResponseHandler(val, included, count, total, err, r)
// 		} else if id == "sent" {
// 			val, included, count, total, err := controllers.GetSentEmails(c, r)
// 			return api.BaseResponseHandler(val, included, count, total, err, r)
// 		} else if id == "search" {
// 			val, included, count, total, err := controllers.GetEmailSearch(c, r)
// 			return api.BaseResponseHandler(val, included, count, total, err, r)
// 		} else if id == "stats" {
// 			val, included, count, total, err := controllers.GetEmailStats(c, r)
// 			return api.BaseResponseHandler(val, included, count, total, err, r)
// 		} else if id == "campaigns" {
// 			val, included, count, total, err := controllers.GetEmailCampaigns(c, r)
// 			return api.BaseResponseHandler(val, included, count, total, err, r)
// 		} else if id == "limits" {
// 			return api.BaseSingleResponseHandler(controllers.GetEmailProviderLimits(c, r))
// 		}
// 		return api.BaseSingleResponseHandler(controllers.GetEmail(c, r, id))
// 	case "PATCH":
// 		return api.BaseSingleResponseHandler(controllers.UpdateSingleEmail(c, r, id))
// 	case "POST":
// 		if id == "upload" {
// 			return api.BaseSingleResponseHandler(files.HandleEmailImageActionUpload(c, r))
// 		} else if id == "bulksend" {
// 			val, included, count, total, err := controllers.BulkSendEmail(c, r)
// 			return api.BaseResponseHandler(val, included, count, total, err, r)
// 		} else if id == "bulkcancel" {
// 			val, included, count, total, err := controllers.BulkCancelEmail(c, r)
// 			return api.BaseResponseHandler(val, included, count, total, err, r)
// 		} else if id == "bulkattach" {
// 			val, included, count, total, err := files.HandleBulkEmailAttachActionUpload(c, r)
// 			return api.BaseResponseHandler(val, included, count, total, err, r)
// 		}
// 	}
// 	return nil, errors.New("method not implemented")
// }

// func handleEmails(c context.Context, w http.ResponseWriter, r *http.Request) (interface{}, error) {
// 	switch r.Method {
// 	case "GET":
// 		val, included, count, total, err := controllers.GetEmails(c, r)
// 		return api.BaseResponseHandler(val, included, count, total, err, r)
// 	case "POST":
// 		return api.BaseSingleResponseHandler(controllers.CreateEmailTransition(c, r))
// 	case "PATCH":
// 		return api.BaseSingleResponseHandler(controllers.UpdateBatchEmail(c, r))
// 	}
// 	return nil, errors.New("method not implemented")
// }

// // Handler for when the user wants all the contacts.
// func EmailsHandler(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
// 	w.Header().Set("Content-Type", "application/json")
// 	c := appengine.NewContext(r)
// 	val, err := handleEmails(c, w, r)

// 	if err == nil {
// 		err = ffjson.NewEncoder(w).Encode(val)
// 	}

// 	if err != nil {
// 		nError.ReturnError(w, http.StatusInternalServerError, "Email handling error", err.Error())
// 	}
// 	return
// }

// // Handler for when there is a key present after /users/<id> route.
// func EmailHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
// 	w.Header().Set("Content-Type", "application/json")
// 	c := appengine.NewContext(r)
// 	id := ps.ByName("id")
// 	val, err := handleEmail(c, r, id)

// 	if err == nil {
// 		err = ffjson.NewEncoder(w).Encode(val)
// 	}

// 	if err != nil {
// 		nError.ReturnError(w, http.StatusInternalServerError, "Email handling error", err.Error())
// 	}
// 	return
// }

// func EmailActionHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
// 	w.Header().Set("Content-Type", "application/json")
// 	c := appengine.NewContext(r)
// 	id := ps.ByName("id")
// 	action := ps.ByName("action")
// 	val, err := handleEmailAction(c, r, id, action)

// 	if err == nil {
// 		err = ffjson.NewEncoder(w).Encode(val)
// 	}

// 	if err != nil {
// 		nError.ReturnError(w, http.StatusInternalServerError, "Email handling error", err.Error())
// 	}
// 	return
// }
