package search

import (
	"log"
	"net/http"

	gcontext "github.com/gorilla/context"

	apiSearch "github.com/news-ai/api-v1/search"

	tabulaeModels "github.com/news-ai/tabulae-v1/models"

	elastic "github.com/news-ai/elastic-appengine"
)

var (
	elasticList *elastic.Elastic
)

func searchList(elasticQuery interface{}) ([]tabulaeModels.MediaList, int, error) {
	hits, err := elasticList.QueryStruct(elasticQuery)
	if err != nil {
		log.Printf("%v", err)
		return []tabulaeModels.MediaList{}, 0, err
	}

	listHits := hits.Hits
	lists := []tabulaeModels.MediaList{}
	for i := 0; i < len(listHits); i++ {
		rawList := listHits[i].Source.Data
		rawMap := rawList.(map[string]interface{})
		contact := tabulaeModels.MediaList{}
		err := contact.FillStruct(rawMap)
		if err != nil {
			log.Printf("%v", err)
		}

		contact.Type = "lists"
		lists = append(lists, contact)
	}

	return lists, hits.Total, nil
}

func SearchListsByClientName(r *http.Request, clientName string, userId int64) ([]tabulaeModels.MediaList, int, error) {
	if clientName == "" {
		return []tabulaeModels.MediaList{}, 0, nil
	}

	offset := gcontext.Get(r, "offset").(int)
	limit := gcontext.Get(r, "limit").(int)

	elasticQuery := elastic.ElasticQueryMust{}
	elasticQuery.Size = limit
	elasticQuery.From = offset

	elasticCreatedByQuery := apiSearch.ElasticCreatedByQuery{}
	elasticCreatedByQuery.Term.CreatedBy = userId
	elasticQuery.Query.Bool.Must = append(elasticQuery.Query.Bool.Must, elasticCreatedByQuery)

	elasticArchivedQuery := apiSearch.ElasticArchivedQuery{}
	elasticArchivedQuery.Term.Archived = false
	elasticQuery.Query.Bool.Must = append(elasticQuery.Query.Bool.Must, elasticArchivedQuery)

	elasticClientQuery := apiSearch.ElasticClientQuery{}
	elasticClientQuery.Term.Client = clientName
	elasticQuery.Query.Bool.Must = append(elasticQuery.Query.Bool.Must, elasticClientQuery)

	elasticQuery.MinScore = float32(0.5)
	return searchList(elasticQuery)
}

func SearchListsByTag(r *http.Request, tag string, userId int64) ([]tabulaeModels.MediaList, int, error) {
	if tag == "" {
		return []tabulaeModels.MediaList{}, 0, nil
	}

	offset := gcontext.Get(r, "offset").(int)
	limit := gcontext.Get(r, "limit").(int)

	elasticQuery := elastic.ElasticQueryMustWithSort{}
	elasticQuery.Size = limit
	elasticQuery.From = offset

	elasticCreatedByQuery := apiSearch.ElasticCreatedByQuery{}
	elasticCreatedByQuery.Term.CreatedBy = userId
	elasticQuery.Query.Bool.Must = append(elasticQuery.Query.Bool.Must, elasticCreatedByQuery)

	elasticArchivedQuery := apiSearch.ElasticArchivedQuery{}
	elasticArchivedQuery.Term.Archived = false
	elasticQuery.Query.Bool.Must = append(elasticQuery.Query.Bool.Must, elasticArchivedQuery)

	elasticTagQuery := apiSearch.ElasticTagQuery{}
	elasticTagQuery.Term.Tag = tag
	elasticQuery.Query.Bool.Must = append(elasticQuery.Query.Bool.Must, elasticTagQuery)

	elasticCreatedQuery := apiSearch.ElasticSortDataCreatedQuery{}
	elasticCreatedQuery.DataCreated.Order = "desc"
	elasticCreatedQuery.DataCreated.Mode = "avg"
	elasticQuery.Sort = append(elasticQuery.Sort, elasticCreatedQuery)

	elasticQuery.MinScore = float32(0.5)
	return searchList(elasticQuery)
}

func SearchListsByAll(r *http.Request, query string, userId int64) ([]tabulaeModels.MediaList, int, error) {
	if query == "" {
		return []tabulaeModels.MediaList{}, 0, nil
	}

	offset := gcontext.Get(r, "offset").(int)
	limit := gcontext.Get(r, "limit").(int)

	elasticQuery := elastic.ElasticQueryMust{}
	elasticQuery.Size = limit
	elasticQuery.From = offset

	elasticCreatedByQuery := apiSearch.ElasticCreatedByQuery{}
	elasticCreatedByQuery.Term.CreatedBy = userId
	elasticQuery.Query.Bool.Must = append(elasticQuery.Query.Bool.Must, elasticCreatedByQuery)

	elasticArchivedQuery := apiSearch.ElasticArchivedQuery{}
	elasticArchivedQuery.Term.Archived = false
	elasticQuery.Query.Bool.Must = append(elasticQuery.Query.Bool.Must, elasticArchivedQuery)

	elasticAllQuery := apiSearch.ElasticAllQuery{}
	elasticAllQuery.Term.All = query
	elasticQuery.Query.Bool.Must = append(elasticQuery.Query.Bool.Must, elasticAllQuery)

	elasticQuery.MinScore = float32(0.5)
	return searchList(elasticQuery)
}

func SearchListsByFieldSelector(r *http.Request, fieldSelector string, query string, userId int64) ([]tabulaeModels.MediaList, int, error) {
	if fieldSelector == "client" {
		return SearchListsByClientName(r, query, userId)
	}
	return SearchListsByTag(r, query, userId)
}
