package controllers

import (
	"errors"
	"fmt"
	"log"
	"net/http"

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

func getFile(r *http.Request, id int64) (models.File, error) {
	if id == 0 {
		return models.File{}, errors.New("datastore: no such entity")
	}

	file := models.File{}
	err := db.DB.Model(&file).Where("id = ?", id).Select()
	if err != nil {
		log.Printf("%v", err)
		return models.File{}, err
	}

	if !file.Created.IsZero() {
		file.Type = "feeds"

		user, err := controllers.GetCurrentUser(r)
		if err != nil {
			log.Printf("%v", err)
			return models.File{}, errors.New("Could not get user")
		}

		if file.CreatedBy != user.Id && !user.Data.IsAdmin {
			return models.File{}, errors.New("Forbidden")
		}

		return file, nil
	}

	return models.File{}, errors.New("No file by this id")
}

func getFileUnauthorized(r *http.Request, id int64) (models.File, error) {
	if id == 0 {
		return models.File{}, errors.New("datastore: no such entity")
	}

	file := models.File{}
	err := db.DB.Model(&file).Where("id = ?", id).Select()
	if err != nil {
		log.Printf("%v", err)
		return models.File{}, err
	}

	if !file.Created.IsZero() {
		file.Type = "feeds"
		return file, nil
	}

	return models.File{}, errors.New("No file by this id")
}

/*
* Public methods
 */

/*
* Get methods
 */

// Gets every single file by the user
func GetFiles(r *http.Request) ([]models.File, interface{}, int, int, error) {
	user, err := controllers.GetCurrentUser(r)
	if err != nil {
		log.Printf("%v", err)
		return []models.File{}, nil, 0, 0, err
	}

	files := []models.File{}
	err = db.DB.Model(&files).Where("created_by = ?", user.Id).Select()
	if err != nil {
		log.Printf("%v", err)
		return []models.File{}, nil, 0, 0, err
	}

	for i := 0; i < len(files); i++ {
		files[i].Type = "files"
	}

	return files, nil, len(files), 0, nil
}

func GetFile(r *http.Request, id string) (models.File, interface{}, error) {
	// Get the details of the current user
	currentId, err := utilities.StringIdToInt(id)
	if err != nil {
		log.Printf("%v", err)
		return models.File{}, nil, err
	}

	file, err := getFile(r, currentId)
	if err != nil {
		log.Printf("%v", err)
		return models.File{}, nil, err
	}

	return file, nil, nil
}

func GetFileById(r *http.Request, id int64) (models.File, interface{}, error) {
	file, err := getFile(r, id)
	if err != nil {
		log.Printf("%v", err)
		return models.File{}, nil, err
	}

	return file, nil, nil
}

func GetFileByIdUnauthorized(r *http.Request, id int64) (models.File, interface{}, error) {
	file, err := getFileUnauthorized(r, id)
	if err != nil {
		log.Printf("%v", err)
		return models.File{}, nil, err
	}

	return file, nil, nil
}

func FilterFileByImported(r *http.Request) ([]models.File, error) {
	files := []models.File{}
	err := db.DB.Model(&files).Where("imported = ?", true).Select()
	if err != nil {
		log.Printf("%v", err)
		return []models.File{}, err
	}

	if len(files) == 0 {
		return []models.File{}, errors.New("No files by the field Imported")
	}

	for i := 0; i < len(files); i++ {
		files[i].Type = "files"
	}

	nonImageFiles := []models.File{}
	for i := 0; i < len(files); i++ {
		if files[i].Url == "" {
			files[i].Type = "files"
			nonImageFiles = append(nonImageFiles, files[i])
		}
	}

	return nonImageFiles, nil
}

/*
* Create methods
 */

func CreateFile(r *http.Request, fileName string, listid string, createdby string) (models.File, error) {
	// Convert listId and createdById from string to int64
	listId, err := utilities.StringIdToInt(listid)
	if err != nil {
		log.Printf("%v", err)
		return models.File{}, err
	}
	createdBy, err := utilities.StringIdToInt(createdby)
	if err != nil {
		log.Printf("%v", err)
		return models.File{}, err
	}

	// Initialize file
	file := models.File{}
	file.FileName = fileName
	file.ListId = listId
	file.CreatedBy = createdBy
	file.FileExists = true

	currentUser, err := controllers.GetCurrentUser(r)
	if err != nil {
		log.Printf("%v", err)
		return file, err
	}

	// Create file
	_, err = file.Create(r, currentUser)
	if err != nil {
		log.Printf("%v", err)
		return models.File{}, err
	}

	// Attach the fileId to the media list associated to it
	// mediaList, _, err := GetMediaList(r, listid)
	// if err != nil {
	// 	log.Printf("%v", err)
	// 	return models.File{}, err
	// }
	// mediaList.FileUpload = file.Id
	// mediaList.Save()

	return file, nil
}

func CreateImageFile(r *http.Request, originalFilename string, fileName string, createdby string, bucket string) (models.File, error) {
	// Convert listId and createdById from string to int64
	createdBy, err := utilities.StringIdToInt(createdby)
	if err != nil {
		log.Printf("%v", err)
		return models.File{}, err
	}

	publicURL := "https://storage.googleapis.com/%s/%s"

	// Initialize file
	file := models.File{}
	file.OriginalName = originalFilename
	file.FileName = fileName
	file.CreatedBy = createdBy
	file.FileExists = true
	file.Url = fmt.Sprintf(publicURL, bucket, fileName)

	currentUser, err := controllers.GetCurrentUser(r)
	if err != nil {
		log.Printf("%v", err)
		return file, err
	}

	// Create file
	_, err = file.Create(r, currentUser)
	if err != nil {
		log.Printf("%v", err)
		return models.File{}, err
	}

	return file, nil
}

func CreateAttachmentFile(r *http.Request, originalFilename string, fileName string, emailid string, createdby string) (models.File, error) {
	// Convert listId and createdById from string to int64
	emailId, err := utilities.StringIdToInt(emailid)
	if err != nil {
		log.Printf("%v", err)
		return models.File{}, err
	}
	createdBy, err := utilities.StringIdToInt(createdby)
	if err != nil {
		log.Printf("%v", err)
		return models.File{}, err
	}

	// Initialize file
	file := models.File{}
	file.OriginalName = originalFilename
	file.FileName = fileName
	file.EmailId = emailId
	file.CreatedBy = createdBy
	file.FileExists = true

	currentUser, err := controllers.GetCurrentUser(r)
	if err != nil {
		log.Printf("%v", err)
		return file, err
	}

	// Create file
	_, err = file.Create(r, currentUser)
	if err != nil {
		log.Printf("%v", err)
		return models.File{}, err
	}

	// Attach attachment to email
	// if emailId != 0 {
	// 	email, err := getEmail(r, emailId)
	// 	if err != nil {
	// 		log.Printf("%v", err)
	// 		return models.File{}, err
	// 	}

	// 	email.Attachments = append(email.Attachments, file.Id)
	// 	email.Save()
	// }

	return file, nil
}

/*
* XLSX -> API methods
 */
