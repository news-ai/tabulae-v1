package models

import (
	"errors"
	"log"
	"net/http"
	"time"

	"github.com/news-ai/api-v1/db"
	apiModels "github.com/news-ai/api-v1/models"

	"github.com/news-ai/web/utilities"
)

type Publication struct {
	apiModels.Base

	Name string `json:"name"`
	Url  string `json:"url"`

	// Social information
	LinkedIn  string `json:"linkedin"`
	Twitter   string `json:"twitter"`
	Instagram string `json:"instagram"`
	MuckRack  string `json:"-"`
	Website   string `json:"-"`
	Blog      string `json:"blog"`

	Verified bool `json:"verified"`
}

/*
* Public methods
 */

/*
* Create methods
 */

// Function to create a new publication into App Engine
func (p *Publication) Create(r *http.Request, currentUser apiModels.UserPostgres) (*Publication, error) {
	p.CreatedBy = currentUser.Id
	p.Created = time.Now()
	_, err := db.DB.Model(p).Returning("*").Insert()
	return p, err
}

/*
* Update methods
 */

// Function to save a new publication into App Engine
func (p *Publication) Save() (*Publication, error) {
	// Update the Updated time
	p.Updated = time.Now()
	_, err := db.DB.Model(p).Update()
	return p, err
}

func (p *Publication) Validate() (*Publication, error) {
	// Validate Fields
	if p.Name == "" {
		return p, errors.New("Missing fields")
	}

	// Format URL properly
	if p.Url != "" {
		normalizedUrl, err := utilities.NormalizeUrl(p.Url)
		if err != nil {
			log.Printf("%v", err)
			return p, err
		}
		p.Url = normalizedUrl
	}
	return p, nil
}

func (p *Publication) FillStruct(m map[string]interface{}) error {
	for k, v := range m {
		err := apiModels.SetField(p, k, v)
		if err != nil {
			return err
		}
	}
	return nil
}
