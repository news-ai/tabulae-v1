package models

import (
	"net/http"
	"time"

	"github.com/news-ai/api-v1/db"
	apiModels "github.com/news-ai/api-v1/models"
)

type Template struct {
	apiModels.Base

	Name    string `json:"name"`
	Subject string `json:"subject" datastore:",noindex"`
	Body    string `json:"body" datastore:",noindex"`

	Archived bool `json:"archived"`
}

/*
* Public methods
 */

/*
* Create methods
 */

// Function to create a new team into App Engine
func (tpl *Template) Create(r *http.Request, currentUser apiModels.UserPostgres) (*Template, error) {
	tpl.CreatedBy = currentUser.Id
	tpl.Created = time.Now()
	_, err := db.DB.Model(tpl).Returning("*").Insert()
	return tpl, err
}

/*
* Update methods
 */

// Function to save a new team into App Engine
func (tpl *Template) Save() (*Template, error) {
	_, err := db.DB.Model(tpl).Update()
	return tpl, err
}
