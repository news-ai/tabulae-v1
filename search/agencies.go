package search

import (
	"log"
	"net/http"
	"net/url"

	gcontext "github.com/gorilla/context"

	"github.com/news-ai/api-v1/models"

	elastic "github.com/news-ai/elastic-appengine"
)

var (
	elasticAgency *elastic.Elastic
)

func SearchAgency(r *http.Request, search string) ([]models.Agency, int, error) {
	search = url.QueryEscape(search)
	search = "q=data.Name:" + search

	offset := gcontext.Get(r, "offset").(int)
	limit := gcontext.Get(r, "limit").(int)

	hits, err := elasticAgency.Query(offset, limit, search)
	if err != nil {
		log.Printf("%v", err)
		return []models.Agency{}, 0, err
	}

	agencyHits := hits.Hits
	agencies := []models.Agency{}
	for i := 0; i < len(agencyHits); i++ {
		rawAgency := agencyHits[i].Source.Data
		rawMap := rawAgency.(map[string]interface{})
		agency := models.Agency{}
		err := agency.FillStruct(rawMap)
		if err != nil {
			log.Printf("%v", err)
		}

		agency.Type = "agencies"
		agencies = append(agencies, agency)
	}

	return agencies, hits.Total, nil
}
