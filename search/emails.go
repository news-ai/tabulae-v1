package search

import (
	"log"
	"net/http"
	"strings"

	gcontext "github.com/gorilla/context"
	elastic "github.com/news-ai/elastic-appengine"

	apiModels "github.com/news-ai/api-v1/models"
	apiSearch "github.com/news-ai/api-v1/search"

	"github.com/news-ai/tabulae-v1/models"

	"github.com/news-ai/web/utilities"
)

var (
	elasticEmailLog        *elastic.Elastic
	elasticEmailTimeseries *elastic.Elastic
	elasticEmails          *elastic.Elastic
)

func searchEmail(elasticQuery interface{}) (interface{}, int, int, error) {
	hits, err := elasticEmailLog.QueryStruct(elasticQuery)
	if err != nil {
		log.Printf("%v", err)
		return nil, 0, 0, err
	}

	emailLogHits := []interface{}{}
	for i := 0; i < len(hits.Hits); i++ {
		emailLogHits = append(emailLogHits, hits.Hits[i].Source.Data)
	}

	return emailLogHits, len(emailLogHits), hits.Total, nil
}

func searchEmailTimeseries(elasticQuery interface{}) (interface{}, int, int, error) {
	hits, err := elasticEmailTimeseries.QueryStruct(elasticQuery)
	if err != nil {
		log.Printf("%v", err)
		return nil, 0, 0, err
	}

	emailTimeseriesHits := []interface{}{}
	for i := 0; i < len(hits.Hits); i++ {
		emailTimeseriesHits = append(emailTimeseriesHits, hits.Hits[i].Source.Data)
	}

	return emailTimeseriesHits, len(emailTimeseriesHits), hits.Total, nil
}

func searchEmailQuery(elasticQuery interface{}) ([]models.Email, int, int, error) {
	hits, err := elasticEmails.QueryStruct(elasticQuery)
	if err != nil {
		log.Printf("%v", err)
		return []models.Email{}, 0, 0, err
	}

	emailHits := hits.Hits
	emailLogHits := []models.Email{}
	for i := 0; i < len(emailHits); i++ {
		rawFeed := emailHits[i].Source.Data
		rawMap := rawFeed.(map[string]interface{})
		email := models.Email{}
		err := email.FillStruct(rawMap)
		if err != nil {
			log.Printf("%v", err)
		}

		if email.Opened == 0 {
			if email.Method == "sendgrid" && email.SendGridId == "" {
				continue
			} else if email.Method == "gmail" && email.GmailId == "" {
				continue
			}
		}

		email.Type = "emails"
		emailLogHits = append(emailLogHits, email)
	}

	return emailLogHits, len(emailLogHits), hits.Total, nil
}

func SearchEmailTimeseriesByUserId(r *http.Request, user apiModels.User) (interface{}, int, int, error) {
	offset := gcontext.Get(r, "offset").(int)
	limit := gcontext.Get(r, "limit").(int)

	elasticQuery := elastic.ElasticQueryWithSort{}
	elasticQuery.Size = limit
	elasticQuery.From = offset

	elasticUserIdQuery := apiSearch.ElasticUserIdQuery{}
	elasticUserIdQuery.Term.UserId = user.Id
	elasticQuery.Query.Bool.Must = append(elasticQuery.Query.Bool.Must, elasticUserIdQuery)

	elasticDateQuery := apiSearch.ElasticSortDataQuery{}
	elasticDateQuery.Date.Order = "desc"
	elasticDateQuery.Date.Mode = "avg"
	elasticQuery.Sort = append(elasticQuery.Sort, elasticDateQuery)

	return searchEmailTimeseries(elasticQuery)
}

func SearchEmailLogByEmailId(r *http.Request, user apiModels.User, emailId int64) (interface{}, int, int, error) {
	if emailId == 0 {
		return nil, 0, 0, nil
	}

	offset := gcontext.Get(r, "offset").(int)
	limit := gcontext.Get(r, "limit").(int)

	elasticQuery := elastic.ElasticQuery{}
	elasticQuery.Size = limit
	elasticQuery.From = offset

	elasticEmailIdQuery := apiSearch.ElasticEmailIdQuery{}
	elasticEmailIdQuery.Term.EmailId = emailId
	elasticQuery.Query.Bool.Must = append(elasticQuery.Query.Bool.Must, elasticEmailIdQuery)

	return searchEmail(elasticQuery)
}

func SearchEmailsByQuery(r *http.Request, user apiModels.User, searchQuery string) ([]models.Email, int, int, error) {
	if searchQuery == "" {
		return nil, 0, 0, nil
	}

	offset := gcontext.Get(r, "offset").(int)
	limit := gcontext.Get(r, "limit").(int)

	elasticQuery := elastic.ElasticQueryWithSort{}
	elasticQuery.Size = limit
	elasticQuery.From = offset

	elasticCreatedByQuery := apiSearch.ElasticCreatedByQuery{}
	elasticCreatedByQuery.Term.CreatedBy = user.Id

	elasticIsSentQuery := apiSearch.ElasticIsSentQuery{}
	elasticIsSentQuery.Term.IsSent = true

	elasticCancelQuery := apiSearch.ElasticCancelQuery{}
	elasticCancelQuery.Term.Cancel = false

	elasticQuery.Query.Bool.Must = append(elasticQuery.Query.Bool.Must, elasticCreatedByQuery)
	elasticQuery.Query.Bool.Must = append(elasticQuery.Query.Bool.Must, elasticIsSentQuery)
	elasticQuery.Query.Bool.Must = append(elasticQuery.Query.Bool.Must, elasticCancelQuery)

	email := strings.Replace(searchQuery, "\"", "", -1)
	if utilities.ValidateEmailFormat(email) {
		elasticEmailToQuery := apiSearch.ElasticEmailToQuery{}
		elasticEmailToQuery.Term.To = email
		elasticQuery.Query.Bool.Must = append(elasticQuery.Query.Bool.Must, elasticEmailToQuery)
	} else {
		elasticMatchQuery := elastic.ElasticMatchQuery{}
		elasticMatchQuery.Match.All = searchQuery
		elasticQuery.Query.Bool.Must = append(elasticQuery.Query.Bool.Must, elasticMatchQuery)
	}

	elasticCreatedQuery := apiSearch.ElasticSortDataCreatedQuery{}
	elasticCreatedQuery.DataCreated.Order = "desc"
	elasticCreatedQuery.DataCreated.Mode = "avg"
	elasticQuery.Sort = append(elasticQuery.Sort, elasticCreatedQuery)

	return searchEmailQuery(elasticQuery)
}

func SearchEmailsByQueryFields(r *http.Request, user apiModels.User, emailDate string, emailSubject string, emailBaseSubject string, filter string) ([]models.Email, int, int, error) {
	if emailDate == "" && emailSubject == "" && emailBaseSubject == "" && filter == "" {
		return nil, 0, 0, nil
	}

	offset := gcontext.Get(r, "offset").(int)
	limit := gcontext.Get(r, "limit").(int)

	elasticQuery := elastic.ElasticQueryWithSort{}
	elasticQuery.Size = limit
	elasticQuery.From = offset

	elasticCreatedByQuery := apiSearch.ElasticCreatedByQuery{}
	elasticCreatedByQuery.Term.CreatedBy = user.Id

	elasticIsSentQuery := apiSearch.ElasticIsSentQuery{}
	elasticIsSentQuery.Term.IsSent = true

	elasticCancelQuery := apiSearch.ElasticCancelQuery{}
	elasticCancelQuery.Term.Cancel = false

	elasticArchivedQuery := apiSearch.ElasticArchivedQuery{}
	elasticArchivedQuery.Term.Archived = false

	elasticQuery.Query.Bool.Must = append(elasticQuery.Query.Bool.Must, elasticCreatedByQuery)
	elasticQuery.Query.Bool.Must = append(elasticQuery.Query.Bool.Must, elasticIsSentQuery)
	elasticQuery.Query.Bool.Must = append(elasticQuery.Query.Bool.Must, elasticCancelQuery)
	elasticQuery.Query.Bool.Must = append(elasticQuery.Query.Bool.Must, elasticArchivedQuery)

	if emailDate != "" {
		elasticCreatedFilterQuery := apiSearch.ElasticCreatedRangeQuery{}
		elasticCreatedFilterQuery.Range.DataCreated.From = emailDate + "T00:00:00"
		elasticCreatedFilterQuery.Range.DataCreated.To = emailDate + "T23:59:59"
		elasticQuery.Query.Bool.Must = append(elasticQuery.Query.Bool.Must, elasticCreatedFilterQuery)
	}

	if emailBaseSubject != "" {
		elasticBaseSubjectQuery := apiSearch.ElasticBaseSubjectQuery{}
		elasticBaseSubjectQuery.Term.BaseSubject = emailBaseSubject
		elasticQuery.Query.Bool.Must = append(elasticQuery.Query.Bool.Must, elasticBaseSubjectQuery)
	} else if emailSubject != "" {
		elasticSubjectQuery := apiSearch.ElasticSubjectQuery{}
		elasticSubjectQuery.Term.Subject = emailSubject
		elasticQuery.Query.Bool.Must = append(elasticQuery.Query.Bool.Must, elasticSubjectQuery)
	}

	if filter == "open" {
		elasticOpenedRangeQuery := apiSearch.ElasticOpenedRangeQuery{}
		elasticOpenedRangeQuery.Range.DataOpened.GTE = 1
		elasticQuery.Query.Bool.Must = append(elasticQuery.Query.Bool.Must, elasticOpenedRangeQuery)
	} else if filter == "click" {
		elasticClickedRangeQuery := apiSearch.ElasticClickedRangeQuery{}
		elasticClickedRangeQuery.Range.DataClicked.GTE = 1
		elasticQuery.Query.Bool.Must = append(elasticQuery.Query.Bool.Must, elasticClickedRangeQuery)
	} else if filter == "bounce" {
		elasticBounceQuery := apiSearch.ElasticBounceQuery{}
		elasticBounceQuery.Term.BaseBounced = true
		elasticQuery.Query.Bool.Must = append(elasticQuery.Query.Bool.Must, elasticBounceQuery)
	} else if filter == "unopen" {
		elasticOpenedQuery := apiSearch.ElasticBaseOpenedQuery{}
		elasticOpenedQuery.Term.Opened = 0
		elasticQuery.Query.Bool.Must = append(elasticQuery.Query.Bool.Must, elasticOpenedQuery)
	} else if filter == "unclick" {
		elasticClickedQuery := apiSearch.ElasticBaseClickedQuery{}
		elasticClickedQuery.Term.Clicked = 0
		elasticQuery.Query.Bool.Must = append(elasticQuery.Query.Bool.Must, elasticClickedQuery)
	}

	elasticCreatedQuery := apiSearch.ElasticSortDataCreatedQuery{}
	elasticCreatedQuery.DataCreated.Order = "desc"
	elasticCreatedQuery.DataCreated.Mode = "avg"
	elasticQuery.Sort = append(elasticQuery.Sort, elasticCreatedQuery)

	return searchEmailQuery(elasticQuery)
}

func SearchEmailsByDateAndSubject(r *http.Request, user apiModels.User, emailDate string, subject string, baseSubject string, from, limit int) ([]models.Email, int, int, error) {
	if emailDate == "" {
		return nil, 0, 0, nil
	}

	elasticQuery := elastic.ElasticQueryWithMust{}
	elasticQuery.Size = limit
	elasticQuery.From = from

	elasticCreatedByQuery := apiSearch.ElasticCreatedByQuery{}
	elasticCreatedByQuery.Term.CreatedBy = user.Id

	elasticIsSentQuery := apiSearch.ElasticIsSentQuery{}
	elasticIsSentQuery.Term.IsSent = true

	elasticCancelQuery := apiSearch.ElasticCancelQuery{}
	elasticCancelQuery.Term.Cancel = false

	elasticDelieveredQuery := apiSearch.ElasticDelieveredQuery{}
	elasticDelieveredQuery.Term.Delievered = true

	elasticCreatedFilterQuery := apiSearch.ElasticCreatedRangeQuery{}
	elasticCreatedFilterQuery.Range.DataCreated.From = emailDate + "T00:00:00"
	elasticCreatedFilterQuery.Range.DataCreated.To = emailDate + "T23:59:59"

	elasticQuery.Query.Bool.Must = append(elasticQuery.Query.Bool.Must, elasticCreatedByQuery)

	if baseSubject == "" {
		elasticSubjectQuery := apiSearch.ElasticSubjectQuery{}
		elasticSubjectQuery.Term.Subject = subject
		elasticQuery.Query.Bool.Must = append(elasticQuery.Query.Bool.Must, elasticSubjectQuery)
	} else {
		elasticBaseSubjectQuery := apiSearch.ElasticBaseSubjectQuery{}
		elasticBaseSubjectQuery.Term.BaseSubject = baseSubject
		elasticQuery.Query.Bool.Must = append(elasticQuery.Query.Bool.Must, elasticBaseSubjectQuery)
	}

	elasticQuery.Query.Bool.Must = append(elasticQuery.Query.Bool.Must, elasticIsSentQuery)
	elasticQuery.Query.Bool.Must = append(elasticQuery.Query.Bool.Must, elasticCancelQuery)
	elasticQuery.Query.Bool.Must = append(elasticQuery.Query.Bool.Must, elasticCreatedFilterQuery)
	elasticQuery.Query.Bool.Must = append(elasticQuery.Query.Bool.Must, elasticDelieveredQuery)

	elasticCreatedQuery := apiSearch.ElasticSortDataCreatedQuery{}
	elasticCreatedQuery.DataCreated.Order = "desc"
	elasticCreatedQuery.DataCreated.Mode = "avg"
	elasticQuery.Sort = append(elasticQuery.Sort, elasticCreatedQuery)

	return searchEmailQuery(elasticQuery)
}
