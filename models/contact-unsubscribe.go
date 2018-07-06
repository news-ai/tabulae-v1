package models

import (
	"net/http"
	"time"

	"github.com/news-ai/api-v1/db"
	apiModels "github.com/news-ai/api-v1/models"
)

type ContactUnsubscribe struct {
	apiModels.Base

	ListId    int64 `json:"listid"`
	ContactId int64 `json:"contactid"`
	EmailId   int64 `json:"emailid"`

	Email        string `json:"email"`
	Unsubscribed bool   `json:"unsubscribed"`
}

/*
* Public methods
 */

/*
* Create methods
 */

func (cu *ContactUnsubscribe) Create(r *http.Request) (*ContactUnsubscribe, error) {
	cu.Created = time.Now()
	_, err := db.DB.Model(cu).Returning("*").Insert()
	return cu, err
}

/*
* Update methods
 */

// Function to save a new contact into App Engine
func (cu *ContactUnsubscribe) Save(r *http.Request) (*ContactUnsubscribe, error) {
	// Update the Updated time
	cu.Updated = time.Now()
	_, err := db.DB.Model(cu).Update()
	return cu, err
}
