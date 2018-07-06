package controllers

import (
	"errors"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strings"

	"github.com/news-ai/api-v1/controllers"
	"github.com/news-ai/api-v1/db"
	apiSearch "github.com/news-ai/api-v1/search"

	gcontext "github.com/gorilla/context"
	"github.com/pquerna/ffjson/ffjson"

	"github.com/news-ai/tabulae-v1/models"
	"github.com/news-ai/tabulae-v1/search"
	// 	"github.com/news-ai/tabulae-v1/sync"
	"github.com/news-ai/web/utilities"
)

// /*
// * Private methods
//  */

// /*
// * Get methods
//  */

func getPublication(id int64) (models.Publication, error) {
	if id == 0 {
		return models.Publication{}, errors.New("datastore: no such entity")
	}

	publication := models.Publication{}
	err := db.DB.Model(&publication).Where("id = ?", id).Select()
	if err != nil {
		log.Printf("%v", err)
		return models.Publication{}, err
	}

	if !publication.Created.IsZero() {
		publication.Type = "publications"
		return publication, nil
	}

	return models.Publication{}, errors.New("No publication by this id")
}

// /*
// * Filter methods
//  */

func filterPublication(queryType, query string) (models.Publication, error) {
	// Get a publication by the URL
	publication := models.Publication{}
	err := db.DB.Model(&publication).Where(queryType+" = ?", query).Select()
	if err != nil {
		log.Printf("%v", err)
		return models.Publication{}, err
	}

	if publication.Created.IsZero() {
		publication.Type = "publications"
		return publication, nil
	}

	return models.Publication{}, errors.New("No publication by this " + queryType)
}

/*
* Public methods
 */

/*
* Get methods
 */

func GetPublications(r *http.Request) ([]models.Publication, interface{}, int, int, error) {
	// If user is querying then it is not denied by the server
	queryField := gcontext.Get(r, "q").(string)
	if queryField != "" {
		publications, total, err := search.SearchPublication(r, queryField)
		if err != nil {
			return []models.Publication{}, nil, 0, 0, err
		}
		return publications, nil, len(publications), total, nil
	}

	// Now if user is not querying then check
	user, err := controllers.GetCurrentUser(r)
	if err != nil {
		log.Printf("%v", err)
		return []models.Publication{}, nil, 0, 0, err
	}

	if !user.Data.IsAdmin {
		return []models.Publication{}, nil, 0, 0, errors.New("Forbidden")
	}

	publications := []models.Publication{}
	err = db.DB.Model(&publications).Select()
	if err != nil {
		log.Printf("%v", err)
		return []models.Publication{}, nil, 0, 0, err
	}

	for i := 0; i < len(publications); i++ {
		publications[i].Type = "publications"
	}

	return publications, nil, len(publications), 0, nil
}

func GetPublication(id string) (models.Publication, interface{}, error) {
	// Get a publication by id
	currentId, err := utilities.StringIdToInt(id)
	if err != nil {
		log.Printf("%v", err)
		return models.Publication{}, nil, err
	}

	publication, err := getPublication(currentId)
	if err != nil {
		log.Printf("%v", err)
		return models.Publication{}, nil, err
	}
	return publication, nil, nil
}

func GetHeadlinesForPublication(r *http.Request, id string) (interface{}, interface{}, int, int, error) {
	// Get the details of the current user
	currentId, err := utilities.StringIdToInt(id)
	if err != nil {
		log.Printf("%v", err)
		return nil, nil, 0, 0, err
	}

	headlines, total, err := apiSearch.SearchHeadlinesByPublicationId(r, currentId)
	if err != nil {
		log.Printf("%v", err)
		return nil, nil, 0, 0, err
	}

	return headlines, nil, len(headlines), total, nil
}

func GetEnrichCompanyProfile(r *http.Request, id string) (interface{}, interface{}, error) {
	// Get the details of the current user
	currentId, err := utilities.StringIdToInt(id)
	if err != nil {
		log.Printf("%v", err)
		return nil, nil, err
	}

	publication, err := getPublication(currentId)
	if err != nil {
		log.Printf("%v", err)
		return nil, nil, err
	}

	if publication.Url == "" {
		return nil, nil, errors.New("Publication has no URL")
	}

	publicationUrl, err := url.Parse(publication.Url)
	if err != nil {
		return nil, nil, err
	}

	publicationDetail, err := apiSearch.SearchCompanyDatabase(r, publicationUrl.Host)
	if err != nil {
		log.Printf("%v", err)
		return nil, nil, err
	}

	return publicationDetail.Data, nil, nil
}

// /*
// * Create methods
//  */

func CreatePublication(w http.ResponseWriter, r *http.Request) (interface{}, interface{}, int, int, error) {
	// Parse JSON
	buf, _ := ioutil.ReadAll(r.Body)

	decoder := ffjson.NewDecoder()
	var publication models.Publication
	err := decoder.Decode(buf, &publication)

	if err != nil {
		currentUser, err := controllers.GetCurrentUser(r)
		if err != nil {
			log.Printf("%v", err)
			return []models.Publication{}, nil, 0, 0, err
		}

		var publications []models.Publication
		arrayDecoder := ffjson.NewDecoder()
		err = arrayDecoder.Decode(buf, &publications)

		if err != nil {
			log.Printf("%v", err)
			return []models.Publication{}, nil, 0, 0, err
		}

		newPublications := []models.Publication{}
		for i := 0; i < len(publications); i++ {
			_, err = publications[i].Validate()
			if err != nil {
				log.Printf("%v", err)
				return []models.Publication{}, nil, 0, 0, err
			}

			presentPublication, _, err := FilterPublicationByNameAndUrl(publications[i].Name, publications[i].Url)
			if err != nil {
				_, err = publications[i].Create(r, currentUser)
				if err != nil {
					log.Printf("%v", err)
					return []models.Publication{}, nil, 0, 0, err
				}
				// sync.ResourceSync(r, publications[i].Id, "Publication", "create")
				newPublications = append(newPublications, publications[i])
			} else {
				newPublications = append(newPublications, presentPublication)
			}
		}
		return newPublications, nil, len(newPublications), 0, err
	}

	_, err = publication.Validate()
	if err != nil {
		log.Printf("%v", err)
		return models.Publication{}, nil, 0, 0, err
	}

	presentPublication, _, err := FilterPublicationByNameAndUrl(publication.Name, publication.Url)
	if err != nil {
		currentUser, err := controllers.GetCurrentUser(r)
		if err != nil {
			log.Printf("%v", err)
			return models.Publication{}, nil, 0, 0, err
		}
		// Create publication
		_, err = publication.Create(r, currentUser)
		if err != nil {
			log.Printf("%v", err)
			return models.Publication{}, nil, 0, 0, err
		}
		// sync.ResourceSync(r, publication.Id, "Publication", "create")
		return publication, nil, 1, 0, nil
	}
	return presentPublication, nil, 1, 0, nil
}

// /*
// * Update methods
//  */

func UpdatePublication(r *http.Request, id string) (models.Publication, interface{}, error) {
	// Get the details of the current user
	currentId, err := utilities.StringIdToInt(id)
	if err != nil {
		log.Printf("%v", err)
		return models.Publication{}, nil, err
	}

	publication, err := getPublication(currentId)
	if err != nil {
		log.Printf("%v", err)
		return models.Publication{}, nil, err
	}

	decoder := ffjson.NewDecoder()
	buf, _ := ioutil.ReadAll(r.Body)
	var updatedPublication models.Publication
	err = decoder.Decode(buf, &updatedPublication)
	if err != nil {
		log.Printf("%v", err)
		return models.Publication{}, nil, err
	}

	if publication.Verified {
		return models.Publication{}, nil, errors.New("Url of a verified publication can not be changed")
	}

	// If updated publication url is empty and the publication has not been verified
	if updatedPublication.Url != "" && !publication.Verified {
		publication.Url = updatedPublication.Url
		publication.Save()
	}

	// sync.ResourceSync(r, publication.Id, "Publication", "create")
	return publication, nil, nil
}

func VerifyPublication(r *http.Request, id string) (models.Publication, interface{}, error) {
	// Get the details of the current user
	currentId, err := utilities.StringIdToInt(id)
	if err != nil {
		log.Printf("%v", err)
		return models.Publication{}, nil, err
	}

	publication, err := getPublication(currentId)
	if err != nil {
		log.Printf("%v", err)
		return models.Publication{}, nil, err
	}

	currentUser, err := controllers.GetCurrentUser(r)
	if err != nil {
		log.Printf("%v", err)
		return models.Publication{}, nil, err
	}

	if !currentUser.Data.IsAdmin {
		return models.Publication{}, nil, errors.New("Forbidden")
	}

	publication.Verified = true
	publication.Save()

	// sync.ResourceSync(r, publication.Id, "Publication", "create")
	return publication, nil, nil
}

func UploadFindOrCreatePublication(r *http.Request, name string, url string) (models.Publication, error) {
	name = strings.Trim(name, " ")
	publication, _, err := FilterPublicationByNameAndUrl(name, url)
	if err != nil {
		currentUser, err := controllers.GetCurrentUser(r)
		if err != nil {
			log.Printf("%v", err)
			return models.Publication{}, err
		}

		var newPublication models.Publication
		newPublication.Name = name
		newPublication.Url = url

		_, err = newPublication.Create(r, currentUser)
		if err != nil {
			log.Printf("%v", err)
			return models.Publication{}, err
		}

		return newPublication, nil
	}

	return publication, nil
}

func FindOrCreatePublication(r *http.Request, name string, url string) (models.Publication, error) {
	name = strings.Trim(name, " ")
	publication, _, err := FilterPublicationByNameAndUrl(name, url)
	if err != nil {
		currentUser, err := controllers.GetCurrentUser(r)
		if err != nil {
			log.Printf("%v", err)
			return models.Publication{}, err
		}

		var newPublication models.Publication
		newPublication.Name = name
		_, err = newPublication.Create(r, currentUser)
		if err != nil {
			log.Printf("%v", err)
			return models.Publication{}, err
		}

		// sync.ResourceSync(r, newPublication.Id, "Publication", "create")
		return newPublication, nil
	}

	return publication, nil
}

// /*
// * Filter methods
//  */

func FilterPublicationByNameAndUrl(name string, url string) (models.Publication, interface{}, error) {
	// If the url is not empty then we try and compare it to ones that already exist
	if url != "" {
		publication, err := filterPublication("Url", url)

		// If it does exist then return it
		if err == nil {
			return publication, nil, nil
		}
	}

	// If the url is empty or it doesn't exist then we try search by name
	publication, err := filterPublication("Name", name)
	if err != nil {
		// This means there's no name or url that matches that publication object
		log.Printf("%v", err)
		return models.Publication{}, nil, err
	}

	// If the name does exist then we return it
	// Here we can be a little clever
	if url != "" && publication.Url == "" {
		// If there is a url present in the search object but not in the publication
		// object then we add it and save it.
		publication.Url = url
		publication.Save()
	}

	return publication, nil, nil
}
