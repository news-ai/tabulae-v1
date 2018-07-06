package controllers

import (
	"errors"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/pquerna/ffjson/ffjson"

	"github.com/news-ai/api-v1/controllers"
	"github.com/news-ai/api-v1/db"

	"github.com/news-ai/tabulae-v1/models"
	// "github.com/news-ai/tabulae-v1/sync"

	"github.com/news-ai/web/utilities"
)

/*
* Private methods
 */

/*
* Get methods
 */

func getFeed(id int64) (models.Feed, error) {
	if id == 0 {
		return models.Feed{}, errors.New("datastore: no such entity")
	}

	feed := models.Feed{}
	err := db.DB.Model(&feed).Where("id = ?", id).Select()
	if err != nil {
		log.Printf("%v", err)
		return models.Feed{}, err
	}

	if !feed.Created.IsZero() {
		feed.Type = "feeds"
		return feed, nil
	}

	return models.Feed{}, errors.New("No feed by this id")
}

func filterFeeds(r *http.Request, queryType, query string) ([]models.Feed, error) {
	feeds := []models.Feed{}
	err := db.DB.Model(&feeds).Where(queryType+" = ?", query).Select()
	if err != nil {
		log.Printf("%v", err)
		return []models.Feed{}, err
	}

	for i := 0; i < len(feeds); i++ {
		feeds[i].Type = "feeds"
	}

	return feeds, nil
}

/*
* Public methods
 */

/*
* Get methods
 */

func GetFeed(r *http.Request, id string) (models.Feed, interface{}, error) {
	// Get the details of the current user
	currentId, err := utilities.StringIdToInt(id)
	if err != nil {
		log.Printf("%v", err)
		return models.Feed{}, nil, err
	}

	feed, err := getFeed(currentId)
	if err != nil {
		log.Printf("%v", err)
		return models.Feed{}, nil, err
	}

	return feed, nil, nil
}

func GetFeedById(r *http.Request, id int64) (models.Feed, interface{}, error) {
	feed, err := getFeed(id)
	if err != nil {
		log.Printf("%v", err)
		return models.Feed{}, nil, err
	}
	return feed, nil, nil
}

func GetFeeds(r *http.Request) ([]models.Feed, interface{}, int, int, error) {
	user, err := controllers.GetCurrentUser(r)
	if err != nil {
		log.Printf("%v", err)
		return []models.Feed{}, nil, 0, 0, err
	}

	feeds := []models.Feed{}
	err = db.DB.Model(&feeds).Where("created_by = ?", user.Id).Select()
	if err != nil {
		log.Printf("%v", err)
		return []models.Feed{}, nil, 0, 0, err
	}

	for i := 0; i < len(feeds); i++ {
		feeds[i].Type = "feeds"
	}

	return feeds, nil, len(feeds), 0, nil
}

func GetFeedsByResourceId(r *http.Request, resouceName string, resourceId int64) ([]models.Feed, error) {
	feeds := []models.Feed{}
	err := db.DB.Model(&feeds).Where(resouceName+" = ?", resourceId).Select()
	if err != nil {
		log.Printf("%v", err)
		return []models.Feed{}, err
	}

	for i := 0; i < len(feeds); i++ {
		feeds[i].Type = "feeds"
	}

	return feeds, nil
}

func FilterFeeds(r *http.Request, queryType, query string) ([]models.Feed, error) {
	// User has to be logged in
	_, err := controllers.GetCurrentUser(r)
	if err != nil {
		return []models.Feed{}, err
	}

	return filterFeeds(r, queryType, query)
}

/*
* Create methods
 */

func CreateFeed(r *http.Request) (models.Feed, interface{}, error) {
	buf, _ := ioutil.ReadAll(r.Body)
	decoder := ffjson.NewDecoder()
	var feed models.Feed
	err := decoder.Decode(buf, &feed)
	if err != nil {
		log.Printf("%v", err)
		return models.Feed{}, nil, err
	}

	currentUser, err := controllers.GetCurrentUser(r)
	if err != nil {
		log.Printf("%v", err)
		return feed, nil, err
	}

	feeds := []models.Feed{}
	err = db.DB.Model(&feeds).Where("feed_url = ?", feed.FeedURL).Where("created_by = ?", currentUser.Id).Where("contact_id = ?", feed.ContactId).Select()
	if err != nil {
		log.Printf("%v", err)
		return models.Feed{}, nil, err
	}

	if len(feeds) > 0 {
		return models.Feed{}, nil, errors.New("Feed already exits for the contact")
	}

	baseDomain, err := utilities.NormalizeUrl(feed.FeedURL)
	if err != nil {
		log.Printf("%v", err)
		return models.Feed{}, nil, err
	}

	publicationName, err := utilities.GetTitleFromHTTPRequest(baseDomain)
	if err != nil {
		publicationName, err = utilities.GetDomainName(baseDomain)
		if err != nil {
			log.Printf("%v", err)
			return models.Feed{}, nil, err
		}
	}

	publication, err := FindOrCreatePublication(r, publicationName, baseDomain)
	if err != nil {
		log.Printf("%v", err)
		return models.Feed{}, nil, err
	}

	publication.Url = baseDomain
	publication.Save()

	feed.PublicationId = publication.Id

	// Create feed
	_, err = feed.Create(r, currentUser)
	if err != nil {
		log.Printf("%v", err)
		return models.Feed{}, nil, err
	}

	// Run new feed through pub/sub
	// sync.NewRSSFeedSync(r, feed.FeedURL, feed.PublicationId)

	return feed, nil, nil
}

/*
* Delete methods
 */

func DeleteFeed(r *http.Request, id string) (interface{}, interface{}, error) {
	// Get the details of the current user
	currentId, err := utilities.StringIdToInt(id)
	if err != nil {
		log.Printf("%v", err)
		return models.Feed{}, nil, err
	}

	feed, err := getFeed(currentId)
	if err != nil {
		log.Printf("%v", err)
		return models.Feed{}, nil, err
	}

	_, err = feed.Delete()
	if err != nil {
		log.Printf("%v", err)
		return models.Feed{}, nil, err
	}

	return feed, nil, nil
}
