package parse

import (
	"errors"
	"log"
	"net/http"

	"github.com/news-ai/tabulae/controllers"
	"github.com/news-ai/tabulae/models"
	// "github.com/news-ai/tabulae/sync"

	"github.com/news-ai/goexcel"
	"github.com/news-ai/web/utilities"
)

func FileToExcelSheets(r *http.Request, file []byte, contentType string) (goexcel.Sheet, error) {
	return goexcel.FileToExcelSheets(r, file, contentType)
}

func FileToExcelHeader(r *http.Request, file []byte, contentType string) ([]goexcel.Column, error) {
	return goexcel.FileToExcelHeader(r, file, contentType)
}

func ExcelHeadersToListModel(r *http.Request, file []byte, fileName string, headerNames []string, headers []string, mediaListid int64, contentType string) (models.MediaList, error) {
	// Batch get all the contacts
	contacts, customFields, err := goexcel.HeadersToListModel(r, file, headers, contentType)
	if err != nil {
		log.Printf("%v", err)
		return models.MediaList{}, err
	}

	// Batch create all the contact
	contactIds, publicationIds, err := controllers.BatchCreateContactsForExcelUpload(r, contacts, mediaListid)
	if err != nil {
		log.Printf("%v", err)
		return models.MediaList{}, err
	}

	if len(headers) != len(headerNames) {
		log.Infof(c, "%v", headers)
		log.Infof(c, "%v", headerNames)

		headerError := errors.New("Length of headers does not match length of header names")
		return models.MediaList{}, headerError
	}

	// Create a media list
	mediaListId := utilities.IntIdToString(mediaListid)
	mediaList, _, err := controllers.GetMediaList(r, mediaListId)
	mediaList.Contacts = contactIds
	for i := 0; i < len(headers); i++ {
		if _, ok := customFields[headers[i]]; ok {
			if headers[i] != "ignore_column" {
				customField := models.CustomFieldsMap{}
				customField.Name = headerNames[i]
				customField.Value = headers[i]
				customField.CustomField = true
				customField.Hidden = false
				mediaList.FieldsMap = append(mediaList.FieldsMap, customField)
			}
		}
	}

	// Save the media list
	mediaList.Save()
	// sync.ListUploadResourceBulkSync(r, mediaList.Id, mediaList.Contacts, publicationIds)

	return mediaList, nil
}
