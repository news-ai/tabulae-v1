package controllers

import (
	"errors"
	"log"
	"net/http"

	"github.com/news-ai/api-v1/controllers"
	"github.com/news-ai/api-v1/db"

	"github.com/news-ai/tabulae-v1/models"
)

/*
* Private methods
 */

/*
* Get methods
 */

func getUnsubscribedContact(id int64) (models.ContactUnsubscribe, error) {
	if id == 0 {
		return models.ContactUnsubscribe{}, errors.New("datastore: no such entity")
	}

	contactUnsubscribe := models.ContactUnsubscribe{}
	err := db.DB.Model(&contactUnsubscribe).Where("id = ?", id).Select()
	if err != nil {
		log.Printf("%v", err)
		return models.ContactUnsubscribe{}, err
	}

	if !contactUnsubscribe.Created.IsZero() {
		contactUnsubscribe.Type = "unsubscribedcontacts"
		return contactUnsubscribe, nil
	}

	return models.ContactUnsubscribe{}, errors.New("No contact unsubscribed by this id")
}

/*
* Public methods
 */

/*
* Get methods
 */

// Gets every single agency
func GetUnsubscribedContacts(r *http.Request) ([]models.ContactUnsubscribe, interface{}, int, int, error) {
	// Now if user is not querying then check
	user, err := controllers.GetCurrentUser(r)
	if err != nil {
		log.Printf("%v", err)
		return []models.ContactUnsubscribe{}, nil, 0, 0, err
	}

	if !user.Data.IsAdmin {
		return []models.ContactUnsubscribe{}, nil, 0, 0, errors.New("Forbidden")
	}

	contactUnsubscribes := []models.ContactUnsubscribe{}
	err = db.DB.Model(&contactUnsubscribes).Where("created_by = ?", user.Id).Select()
	if err != nil {
		log.Printf("%v", err)
		return []models.ContactUnsubscribe{}, nil, 0, 0, err
	}

	for i := 0; i < len(contactUnsubscribes); i++ {
		contactUnsubscribes[i].Type = "unsubscribedcontacts"
	}

	return contactUnsubscribes, nil, len(contactUnsubscribes), 0, nil
}
