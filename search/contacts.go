package search

import (
	"log"
	"net/http"
	// "net/url"
	// "strconv"

	gcontext "github.com/gorilla/context"
	elastic "github.com/news-ai/elastic-appengine"

	apiModels "github.com/news-ai/api-v1/models"
	apiSearch "github.com/news-ai/api-v1/search"

	"github.com/news-ai/tabulae-v1/models"
)

var (
	elasticContact *elastic.Elastic
)

func searchContact(elasticQuery elastic.ElasticQuery) ([]models.Contact, int, error) {
	hits, err := elasticContact.QueryStruct(elasticQuery)
	if err != nil {
		log.Printf("%v", err)
		return []models.Contact{}, 0, err
	}

	contactHits := hits.Hits
	contacts := []models.Contact{}
	for i := 0; i < len(contactHits); i++ {
		rawContact := contactHits[i].Source.Data
		rawMap := rawContact.(map[string]interface{})
		contact := models.Contact{}
		err := contact.FillStruct(rawMap)
		if err != nil {
			log.Printf("%v", err)
		}

		contact.Type = "contacts"
		contacts = append(contacts, contact)
	}

	return contacts, hits.Total, nil
}

func SearchContacts(r *http.Request, search string, userId int64) ([]models.Contact, int, error) {
	if userId == 0 || search == "" {
		return []models.Contact{}, 0, nil
	}

	offset := gcontext.Get(r, "offset").(int)
	limit := gcontext.Get(r, "limit").(int)

	elasticQuery := elastic.ElasticQuery{}
	elasticQuery.Size = limit
	elasticQuery.From = offset

	elasticCreatedByQuery := apiSearch.ElasticCreatedByQuery{}
	elasticCreatedByQuery.Term.CreatedBy = userId

	elasticMatchQuery := elastic.ElasticMatchQuery{}
	elasticMatchQuery.Match.All = search

	elasticQuery.Query.Bool.Must = append(elasticQuery.Query.Bool.Must, elasticCreatedByQuery)
	elasticQuery.Query.Bool.Must = append(elasticQuery.Query.Bool.Must, elasticMatchQuery)

	return searchContact(elasticQuery)
}

func SearchContactsByList(r *http.Request, search string, user apiModels.User, userId int64, listId int64) ([]models.Contact, int, error) {
	if listId == 0 || search == "" {
		return []models.Contact{}, 0, nil
	}

	offset := gcontext.Get(r, "offset").(int)
	limit := gcontext.Get(r, "limit").(int)

	elasticQuery := elastic.ElasticQuery{}
	elasticQuery.Size = limit
	elasticQuery.From = offset

	if !user.IsAdmin {
		elasticCreatedByQuery := apiSearch.ElasticCreatedByQuery{}
		elasticCreatedByQuery.Term.CreatedBy = userId
		elasticQuery.Query.Bool.Must = append(elasticQuery.Query.Bool.Must, elasticCreatedByQuery)
	}

	elasticListIdQuery := apiSearch.ElasticListIdQuery{}
	elasticListIdQuery.Term.ListId = listId

	elasticIsDeletedQuery := apiSearch.ElasticIsDeletedQuery{}
	elasticIsDeletedQuery.Term.IsDeleted = false

	elasticMatchQuery := elastic.ElasticMatchQuery{}
	elasticMatchQuery.Match.All = search

	elasticQuery.Query.Bool.Must = append(elasticQuery.Query.Bool.Must, elasticListIdQuery)
	elasticQuery.Query.Bool.Must = append(elasticQuery.Query.Bool.Must, elasticIsDeletedQuery)
	elasticQuery.Query.Bool.Must = append(elasticQuery.Query.Bool.Must, elasticMatchQuery)

	return searchContact(elasticQuery)
}

func SearchContactsByTag(r *http.Request, tag string, userId int64) ([]models.Contact, int, error) {
	if tag == "" {
		return []models.Contact{}, 0, nil
	}

	offset := gcontext.Get(r, "offset").(int)
	limit := gcontext.Get(r, "limit").(int)

	elasticQuery := elastic.ElasticQuery{}
	elasticQuery.Size = limit
	elasticQuery.From = offset

	elasticCreatedByQuery := apiSearch.ElasticCreatedByQuery{}
	elasticCreatedByQuery.Term.CreatedBy = userId
	elasticQuery.Query.Bool.Must = append(elasticQuery.Query.Bool.Must, elasticCreatedByQuery)

	elasticTagQuery := apiSearch.ElasticTagQuery{}
	elasticTagQuery.Term.Tag = tag
	elasticQuery.Query.Bool.Must = append(elasticQuery.Query.Bool.Must, elasticTagQuery)

	return searchContact(elasticQuery)
}

func SearchContactsByPublicationId(r *http.Request, publicationId string, userId int64) ([]models.Contact, int, error) {
	if publicationId == "" {
		return []models.Contact{}, 0, nil
	}

	offset := gcontext.Get(r, "offset").(int)
	limit := gcontext.Get(r, "limit").(int)

	elasticQuery := elastic.ElasticQuery{}
	elasticQuery.Size = limit
	elasticQuery.From = offset

	elasticCreatedByQuery := apiSearch.ElasticCreatedByQuery{}
	elasticCreatedByQuery.Term.CreatedBy = userId
	elasticQuery.Query.Bool.Must = append(elasticQuery.Query.Bool.Must, elasticCreatedByQuery)

	elasticEmployersQuery := apiSearch.ElasticEmployersQuery{}
	elasticEmployersQuery.Term.Employers = publicationId
	elasticQuery.Query.Bool.Must = append(elasticQuery.Query.Bool.Must, elasticEmployersQuery)

	return searchContact(elasticQuery)
}

func SearchContactsByFieldSelector(r *http.Request, fieldSelector string, query string, userId int64) ([]models.Contact, int, error) {
	if fieldSelector == "tag" {
		return SearchContactsByTag(r, query, userId)
	} else if fieldSelector == "publication" {
		return SearchContactsByPublicationId(r, query, userId)
	}

	return []models.Contact{}, 0, nil
}
