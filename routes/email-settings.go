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

func handleEmailSettingAction(r *http.Request, id string, action string) (interface{}, error) {
	switch r.Method {
	case "GET":
		switch action {
		case "verify":
			return api.BaseSingleResponseHandler(controllers.VerifyEmailSetting(r, id))
		case "details":
			return api.BaseSingleResponseHandler(controllers.GetEmailSettingDetails(r, id))
		}
	}
	return nil, errors.New("method not implemented")
}

func handleEmailSetting(r *http.Request, id string) (interface{}, error) {
	switch r.Method {
	case "GET":
		return api.BaseSingleResponseHandler(controllers.GetEmailSetting(r, id))
	case "POST":
		switch id {
		case "add-email":
			return api.BaseSingleResponseHandler(controllers.AddUserEmail(r))
		}
	}
	return nil, errors.New("method not implemented")
}

func handleEmailSettings(w http.ResponseWriter, r *http.Request) (interface{}, error) {
	switch r.Method {
	case "GET":
		val, included, count, total, err := controllers.GetEmailSettings(r)
		return api.BaseResponseHandler(val, included, count, total, err, r)
	case "POST":
		return api.BaseSingleResponseHandler(controllers.CreateEmailSettings(r))
	}
	return nil, errors.New("method not implemented")
}

// Handler for when the user wants all the contacts.
func EmailSettingsHandler(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	w.Header().Set("Content-Type", "application/json")
	val, err := handleEmailSettings(w, r)

	if err == nil {
		err = ffjson.NewEncoder(w).Encode(val)
	}

	if err != nil {
		nError.ReturnError(w, http.StatusInternalServerError, "Email setting handling error", err.Error())
	}
	return
}

// Handler for when there is a key present after /emailsettings/<id> route.
func EmailSettingHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	w.Header().Set("Content-Type", "application/json")
	id := ps.ByName("id")
	val, err := handleEmailSetting(r, id)

	if err == nil {
		err = ffjson.NewEncoder(w).Encode(val)
	}

	if err != nil {
		nError.ReturnError(w, http.StatusInternalServerError, "Email setting handling error", err.Error())
	}
	return
}

func EmailSettingActionHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	w.Header().Set("Content-Type", "application/json")
	id := ps.ByName("id")
	action := ps.ByName("action")
	val, err := handleEmailSettingAction(r, id, action)

	if err == nil {
		err = ffjson.NewEncoder(w).Encode(val)
	}

	if err != nil {
		nError.ReturnError(w, http.StatusInternalServerError, "Email setting handling error", err.Error())
	}
	return
}
