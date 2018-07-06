package models

import (
	"net/http"
	"strings"
	"time"

	"github.com/news-ai/web/utilities"

	"github.com/news-ai/api-v1/db"
	apiModels "github.com/news-ai/api-v1/models"
)

type ContactIdsArray struct {
	ContactIds []int64 `json:"ids"`
	Days       int     `json:"days"`
}

type CustomContactField struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type Contact struct {
	apiModels.Base

	// Contact information
	FirstName string `json:"firstname"`
	LastName  string `json:"lastname"`
	Email     string `json:"email"`

	ListId int64 `json:"listid" apiModel:"MediaList"`

	// Notes on a particular contact
	Notes string `json:"notes"`

	// Publications this contact works for now and before
	Employers     []int64 `json:"employers" apiModel:"Publication"`
	PastEmployers []int64 `json:"pastemployers" apiModel:"Publication"`

	// Social information
	LinkedIn  string `json:"linkedin"`
	Twitter   string `json:"twitter"`
	Instagram string `json:"instagram"`
	MuckRack  string `json:"-"`
	Website   string `json:"website"`
	Blog      string `json:"blog"`

	TwitterInvalid   bool `json:"twitterinvalid"`
	InstagramInvalid bool `json:"instagraminvalid"`

	TwitterPrivate   bool `json:"twitterprivate"`
	InstagramPrivate bool `json:"instagramprivate"`

	Location    string `json:"location"`
	PhoneNumber string `json:"phonenumber"`

	// Custom fields
	CustomFields []CustomContactField `json:"customfields" datastore:",noindex"`

	// Is information outdated
	IsOutdated   bool `json:"isoutdated"`
	EmailBounced bool `json:"emailbounced"`

	// Parent contact
	IsMasterContact bool  `json:"ismastercontact"`
	ParentContact   int64 `json:"parent" apiModel:"Contact"`

	IsDeleted bool `json:"isdeleted"`
	ReadOnly  bool `json:"readonly" datastore:"-"`

	ImageURL string `json:"imageurl"`

	Tags []string `json:"tags"`

	TeamId   int64 `json:"teamid"`
	ClientId int64 `json:"clientid"`

	LinkedInUpdated time.Time `json:"linkedinupdated"`
}

/*
* Public methods
 */

/*
* Create methods
 */

func (ct *Contact) Create(r *http.Request, currentUser apiModels.UserPostgres) (*Contact, error) {
	ct.CreatedBy = currentUser.Id
	ct.Created = time.Now()
	ct.Normalize()
	ct.FormatName()
	_, err := db.DB.Model(ct).Returning("*").Insert()
	return ct, err
}

/*
* Update methods
 */

// Function to save a new contact into App Engine
func (ct *Contact) Save(r *http.Request) (*Contact, error) {
	// Update the Updated time
	ct.Updated = time.Now()
	ct.Normalize()
	ct.FormatName()
	_, err := db.DB.Model(ct).Update()
	return ct, err
}

/*
* Normalization methods
 */

func (ct *Contact) FormatName() (*Contact, error) {
	if len(ct.FirstName) > 0 && len(ct.LastName) == 0 {
		if strings.Contains(ct.FirstName, ",") {
			nameSplit := strings.Split(ct.FirstName, ", ")
			if len(nameSplit) > 1 {
				// Agarwal, Abhi as First Name
				ct.FirstName = nameSplit[1]
				ct.LastName = nameSplit[0]
			}
		} else if strings.Contains(ct.FirstName, " ") {
			// Abhi Agarwal as First Name
			nameSplit := strings.Split(ct.FirstName, " ")
			if len(nameSplit) > 1 {
				ct.FirstName = nameSplit[0]
				ct.LastName = strings.Join(nameSplit[1:], " ")
			}
		}
	}

	return ct, nil
}

func (ct *Contact) Normalize() (*Contact, error) {
	ct.LinkedIn = strings.ToLower(utilities.StripQueryString(ct.LinkedIn))
	ct.Twitter = utilities.StripQueryString(ct.Twitter)
	ct.Twitter = strings.ToLower(utilities.NormalizeUrlToUsername(ct.Twitter, "twitter.com"))
	ct.Instagram = utilities.StripQueryString(ct.Instagram)
	ct.Instagram = strings.ToLower(utilities.NormalizeUrlToUsername(ct.Instagram, "instagram.com"))
	ct.MuckRack = strings.ToLower(utilities.StripQueryString(ct.MuckRack))

	ct.Website = utilities.StripQueryStringForWebsite(ct.Website)
	ct.Blog = utilities.StripQueryStringForWebsite(ct.Blog)

	ct.Email = strings.ToLower(ct.Email)
	ct.Email = strings.TrimSpace(ct.Email)

	ct.FirstName = strings.TrimSpace(ct.FirstName)
	ct.LastName = strings.TrimSpace(ct.LastName)

	return ct, nil
}

/*
* Action methods
 */

func (c *Contact) FillStruct(m map[string]interface{}) error {
	for k, v := range m {
		err := apiModels.SetField(c, k, v)
		if err != nil {
			return err
		}
	}
	return nil
}
