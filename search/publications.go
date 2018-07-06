package search

import (
	"log"
	"net/http"
	"net/url"

	gcontext "github.com/gorilla/context"

	elastic "github.com/news-ai/elastic-appengine"
	"github.com/news-ai/tabulae-v1/models"
)

var (
	elasticPublication *elastic.Elastic
)

func SearchPublication(r *http.Request, search string) ([]models.Publication, int, error) {
	search = url.QueryEscape(search)
	search = "q=data.Name:" + search

	offset := gcontext.Get(r, "offset").(int)
	limit := gcontext.Get(r, "limit").(int)

	hits, err := elasticPublication.Query(offset, limit, search)
	if err != nil {
		log.Printf("%v", err)
		return []models.Publication{}, 0, err
	}

	publicationHits := hits.Hits
	publications := []models.Publication{}
	for i := 0; i < len(publicationHits); i++ {
		rawPublication := publicationHits[i].Source.Data
		rawMap := rawPublication.(map[string]interface{})
		publication := models.Publication{}
		err := publication.FillStruct(rawMap)
		if err != nil {
			log.Printf("%v", err)
		}

		publication.Type = "publications"
		publications = append(publications, publication)
	}

	return publications, hits.Total, nil
}
