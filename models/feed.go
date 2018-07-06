package models

import (
	"net/http"
	"time"

	"github.com/news-ai/api-v1/db"
	apiModels "github.com/news-ai/api-v1/models"
)

type Feed struct {
	apiModels.Base

	FeedURL string `json:"url"`

	ContactId     int64 `json:"contactid" apiModel:"Contact"`
	ListId        int64 `json:"listid" apiModel:"MediaList"`
	PublicationId int64 `json:"publicationid" apiModel:"Publication"`

	ValidFeed bool `json:"validfeed"`
	Running   bool `json:"running"`
}

/*
* Private methods
 */

/*
* Create methods
 */

func (f *Feed) Create(r *http.Request, currentUser apiModels.UserPostgres) (*Feed, error) {
	f.CreatedBy = currentUser.Id
	f.Created = time.Now()

	// Initially the feed is both running and valid
	f.Running = true
	f.ValidFeed = true
	_, err := db.DB.Model(f).Returning("*").Insert()
	return f, err
}

/*
* Update methods
 */

// Function to save a new email into App Engine
func (f *Feed) Save() (*Feed, error) {
	_, err := db.DB.Model(f).Update()
	return f, err
}

// Function to save a new user into App Engine
func (f *Feed) Delete() (*Feed, error) {
	err := db.DB.Delete(f)
	return f, err
}
