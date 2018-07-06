package controllers

import (
	"errors"
	"io/ioutil"
	"log"
	"net/http"
	"strings"

	"github.com/pquerna/ffjson/ffjson"

	"github.com/news-ai/api-v1/controllers"
	"github.com/news-ai/api-v1/db"

	"github.com/news-ai/tabulae-v1/models"

	"github.com/news-ai/web/utilities"
)

/*
* Private methods
 */

/*
* Get methods
 */

func getTemplate(id int64) (models.Template, error) {
	if id == 0 {
		return models.Template{}, errors.New("datastore: no such entity")
	}

	template := models.Template{}
	err := db.DB.Model(&template).Where("id = ?", id).Select()
	if err != nil {
		log.Printf("%v", err)
		return models.Template{}, err
	}

	if !template.Created.IsZero() {
		template.Type = "publications"
		return template, nil
	}

	return models.Template{}, errors.New("No template by this id")
}

/*
* Public methods
 */

/*
* Get methods
 */

func GetTemplate(r *http.Request, id string) (models.Template, interface{}, error) {
	// Get the details of the current user
	currentId, err := utilities.StringIdToInt(id)
	if err != nil {
		log.Printf("%v", err)
		return models.Template{}, nil, err
	}

	template, err := getTemplate(currentId)
	if err != nil {
		log.Printf("%v", err)
		return models.Template{}, nil, err
	}

	return template, nil, nil
}

func GetTemplates(r *http.Request) ([]models.Template, interface{}, int, int, error) {
	user, err := controllers.GetCurrentUser(r)
	if err != nil {
		log.Printf("%v", err)
		return []models.Template{}, nil, 0, 0, err
	}

	templates := []models.Template{}
	err = db.DB.Model(&templates).Where("created_by = ?", user.Id).Where("archived = ?", false).Select()
	if err != nil {
		log.Printf("%v", err)
		return []models.Template{}, nil, 0, 0, err
	}

	for i := 0; i < len(templates); i++ {
		templates[i].Type = "templates"
	}

	return templates, nil, len(templates), 0, nil
}

/*
* Create methods
 */

func CreateTemplate(r *http.Request) (models.Template, interface{}, error) {
	buf, _ := ioutil.ReadAll(r.Body)
	decoder := ffjson.NewDecoder()
	var template models.Template
	err := decoder.Decode(buf, &template)
	if err != nil {
		log.Printf("%v", err)
		return models.Template{}, nil, err
	}

	currentUser, err := controllers.GetCurrentUser(r)
	if err != nil {
		log.Printf("%v", err)
		return template, nil, err
	}

	if template.Name == "" {
		template.Name = "Sample"
	}

	if template.Subject == "" {
		template.Subject = template.Name
	}

	if strings.TrimSpace(template.Subject) == "" {
		template.Subject = template.Name
	}

	// Create template
	_, err = template.Create(r, currentUser)
	if err != nil {
		log.Printf("%v", err)
		return models.Template{}, nil, err
	}

	return template, nil, nil
}

/*
* Update methods
 */

func UpdateTemplate(r *http.Request, id string) (models.Template, interface{}, error) {
	// Get the details of the current template
	template, _, err := GetTemplate(r, id)
	if err != nil {
		log.Printf("%v", err)
		return models.Template{}, nil, err
	}

	// Checking if the current user logged in can edit this particular id
	user, err := controllers.GetCurrentUser(r)
	if err != nil {
		log.Printf("%v", err)
		return models.Template{}, nil, err
	}

	if template.CreatedBy != user.Id && !user.Data.IsAdmin {
		return models.Template{}, nil, errors.New("Forbidden")
	}

	decoder := ffjson.NewDecoder()
	buf, _ := ioutil.ReadAll(r.Body)
	var updatedTemplate models.Template
	err = decoder.Decode(buf, &updatedTemplate)
	if err != nil {
		log.Printf("%v", err)
		return models.Template{}, nil, err
	}

	utilities.UpdateIfNotBlank(&template.Name, updatedTemplate.Name)
	utilities.UpdateIfNotBlank(&template.Subject, updatedTemplate.Subject)
	utilities.UpdateIfNotBlank(&template.Body, updatedTemplate.Body)

	// If new template wants to be archived then archive it
	if updatedTemplate.Archived == true {
		template.Archived = true
	}

	// If they are already archived and you want to unarchive the media list
	if template.Archived == true && updatedTemplate.Archived == false {
		template.Archived = false
	}

	template.Save()
	return template, nil, nil
}
