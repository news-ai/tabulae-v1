package models

import (
	"net/http"
	"time"

	"github.com/news-ai/api-v1/db"
	apiModels "github.com/news-ai/api-v1/models"
)

type File struct {
	apiModels.Base

	OriginalName string `json:"originalname"`
	FileName     string `json:"filename"`
	ListId       int64  `json:"listid" apiModel:"MediaList"`
	EmailId      int64  `json:"emailid" apiModel:"Email"`

	Url string `json:"url"`

	HeaderNames []string `json:"headernames" datastore:",noindex"`
	Order       []string `json:"order" datastore:",noindex"`

	Imported   bool `json:"imported"`
	FileExists bool `json:"fileexists"`
}

type FileOrder struct {
	HeaderNames []string `json:"headernames"`
	Order       []string `json:"order"`
	Sheet       string   `json:"string"`
}

/*
* Public methods
 */

/*
* Create methods
 */

func (f *File) Create(r *http.Request, currentUser apiModels.UserPostgres) (*File, error) {
	f.CreatedBy = currentUser.Id
	f.Created = time.Now()
	_, err := db.DB.Model(f).Returning("*").Insert()
	return f, err
}

/*
* Update methods
 */

// Function to save a new file into App Engine
func (f *File) Save() (*File, error) {
	// Update the Updated time
	f.Updated = time.Now()
	_, err := db.DB.Model(f).Update()
	return f, err
}
