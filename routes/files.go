package routes

import (
	"errors"
	"net/http"

	"github.com/julienschmidt/httprouter"
	"github.com/pquerna/ffjson/ffjson"

	"github.com/news-ai/tabulae-v1/controllers"
	// "github.com/news-ai/tabulae-v1/files"

	"github.com/news-ai/web/api"
	nError "github.com/news-ai/web/errors"
)

// func handleFileAction(r *http.Request, id string, action string) (interface{}, error) {
// 	switch r.Method {
// 	case "GET":
// 		switch action {
// 		case "headers":
// 			return api.BaseSingleResponseHandler(files.HandleFileGetHeaders(r, id))
// 		case "sheets":
// 			return api.BaseSingleResponseHandler(files.HandleFileGetSheets(r, id))
// 		}
// 	case "POST":
// 		switch action {
// 		case "headers":
// 			return api.BaseSingleResponseHandler(files.HandleFileUploadHeaders(r, id))
// 		}
// 	}
// 	return nil, errors.New("method not implemented")
// }

func handleFile(r *http.Request, id string) (interface{}, error) {
	switch r.Method {
	case "GET":
		return api.BaseSingleResponseHandler(controllers.GetFile(r, id))
	}
	return nil, errors.New("method not implemented")
}

func handleFiles(w http.ResponseWriter, r *http.Request) (interface{}, error) {
	switch r.Method {
	case "GET":
		val, included, count, total, err := controllers.GetFiles(r)
		return api.BaseResponseHandler(val, included, count, total, err, r)
	}
	return nil, errors.New("method not implemented")
}

// Handler for when the user wants all the files.
func FilesHandler(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	w.Header().Set("Content-Type", "application/json")
	val, err := handleFiles(w, r)

	if err == nil {
		err = ffjson.NewEncoder(w).Encode(val)
	}

	if err != nil {
		nError.ReturnError(w, http.StatusInternalServerError, "Files handling error", err.Error())
		return
	}
}

// Handler for when there is a key present after /files/<id> route.
func FileHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	w.Header().Set("Content-Type", "application/json")
	id := ps.ByName("id")
	val, err := handleFile(r, id)

	if err == nil {
		err = ffjson.NewEncoder(w).Encode(val)
	}

	if err != nil {
		nError.ReturnError(w, http.StatusInternalServerError, "File handling error", err.Error())
	}
	return
}

// func FileActionHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
// 	w.Header().Set("Content-Type", "application/json")
// 	id := ps.ByName("id")
// 	action := ps.ByName("action")
// 	val, err := handleFileAction(r, id, action)

// 	if err == nil {
// 		err = ffjson.NewEncoder(w).Encode(val)
// 	}

// 	if err != nil {
// 		nError.ReturnError(w, http.StatusInternalServerError, "File handling error", err.Error())
// 	}
// 	return
// }
