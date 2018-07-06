package controllers

import (
	"bytes"
	"encoding/json"
	"errors"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"

	"github.com/pquerna/ffjson/ffjson"

	"github.com/news-ai/api-v1/controllers"
	"github.com/news-ai/api-v1/db"
	apiModels "github.com/news-ai/api-v1/models"

	"github.com/news-ai/tabulae-v1/models"

	"github.com/news-ai/web/encrypt"
	"github.com/news-ai/web/permissions"
	"github.com/news-ai/web/utilities"
)

type SMTPEmailResponse struct {
	Status bool   `json:"status"`
	Error  string `json:"error"`
}

func getEmailSetting(r *http.Request, id int64) (models.EmailSetting, error) {
	if id == 0 {
		return models.EmailSetting{}, errors.New("datastore: no such entity")
	}

	emailSetting := models.EmailSetting{}
	err := db.DB.Model(&emailSetting).Where("id = ?", id).Select()
	if err != nil {
		log.Printf("%v", err)
		return models.EmailSetting{}, err
	}

	if !emailSetting.Created.IsZero() {
		emailSetting.Type = "emailsettings"

		user, err := controllers.GetCurrentUser(r)
		if err != nil {
			log.Printf("%v", err)
			return models.EmailSetting{}, errors.New("Could not get user")
		}

		if !permissions.AccessToObject(emailSetting.CreatedBy, user.Id) && !user.Data.IsAdmin {
			return models.EmailSetting{}, errors.New("Forbidden")
		}

		return emailSetting, nil
	}

	return models.EmailSetting{}, errors.New("No email setting by this id")
}

func GetEmailSetting(r *http.Request, id string) (models.EmailSetting, interface{}, error) {
	// Get the details of the current user
	currentId, err := utilities.StringIdToInt(id)
	if err != nil {
		log.Printf("%v", err)
		return models.EmailSetting{}, nil, err
	}

	emailSetting, err := getEmailSetting(r, currentId)
	if err != nil {
		log.Printf("%v", err)
		return models.EmailSetting{}, nil, err
	}

	return emailSetting, nil, nil
}

// To get details without having to authenticate (when sending scheduled emails)
func GetEmailSettingById(r *http.Request, id int64) (models.EmailSetting, error) {
	if id == 0 {
		return models.EmailSetting{}, errors.New("datastore: no such entity")
	}

	emailSetting := models.EmailSetting{}
	err := db.DB.Model(&emailSetting).Where("id = ?", id).Select()
	if err != nil {
		log.Printf("%v", err)
		return models.EmailSetting{}, err
	}

	if !emailSetting.Created.IsZero() {
		emailSetting.Type = "emailsettings"
		return emailSetting, nil
	}

	return models.EmailSetting{}, errors.New("No email setting by this id")
}

func GetEmailSettings(r *http.Request) ([]models.EmailSetting, interface{}, int, int, error) {
	user, err := controllers.GetCurrentUser(r)
	if err != nil {
		log.Printf("%v", err)
		return []models.EmailSetting{}, nil, 0, 0, err
	}

	emailSettings := []models.EmailSetting{}
	err = db.DB.Model(&emailSettings).Where("created_by = ?", user.Id).Select()
	if err != nil {
		log.Printf("%v", err)
		return []models.EmailSetting{}, nil, 0, 0, err
	}

	for i := 0; i < len(emailSettings); i++ {
		emailSettings[i].Type = "emailsettings"
	}

	return emailSettings, nil, len(emailSettings), 0, nil
}

/*
* Create methods
 */

func AddUserEmail(r *http.Request) (apiModels.User, interface{}, error) {
	buf, _ := ioutil.ReadAll(r.Body)

	currentUser, err := controllers.GetCurrentUser(r)
	if err != nil {
		log.Printf("%v", err)
		return apiModels.User{}, nil, err
	}

	decoder := ffjson.NewDecoder()
	var userEmailSettings models.UserEmailSetting
	err = decoder.Decode(buf, &userEmailSettings)
	if err != nil {
		return apiModels.User{}, nil, err
	}

	userPw, err := encrypt.EncryptString(userEmailSettings.SMTPPassword)
	if err != nil {
		return apiModels.User{}, nil, err
	}

	currentUser.Data.SMTPUsername = userEmailSettings.SMTPUsername
	currentUser.Data.SMTPPassword = []byte(userPw)
	currentUser.Data.SMTPValid = false
	controllers.SaveUser(r, &currentUser)

	return currentUser.Data, nil, nil
}

func CreateEmailSettings(r *http.Request) (models.EmailSetting, interface{}, error) {
	buf, _ := ioutil.ReadAll(r.Body)

	currentUser, err := controllers.GetCurrentUser(r)
	if err != nil {
		log.Printf("%v", err)
		return models.EmailSetting{}, nil, err
	}

	decoder := ffjson.NewDecoder()
	var emailSettings models.EmailSetting
	err = decoder.Decode(buf, &emailSettings)
	if err != nil {
		return models.EmailSetting{}, nil, err
	}

	// Create email setting
	_, err = emailSettings.Create(r, currentUser)
	if err != nil {
		log.Printf("%v", err)
		return models.EmailSetting{}, nil, err
	}

	emailSettings.Type = "emailsettings"

	currentUser.Data.EmailSetting = emailSettings.Id
	currentUser.Data.SMTPValid = false
	controllers.SaveUser(r, &currentUser)

	return emailSettings, nil, nil
}

func VerifyEmailSetting(r *http.Request, id string) (SMTPEmailResponse, interface{}, error) {
	emailSetting, _, err := GetEmailSetting(r, id)
	if err != nil {
		log.Printf("%v", err)
		return SMTPEmailResponse{}, nil, err
	}

	currentUser, err := controllers.GetCurrentUser(r)
	if err != nil {
		log.Printf("%v", err)
		return SMTPEmailResponse{}, nil, err
	}

	smtpUser, _, err := controllers.GetUserById(r, emailSetting.CreatedBy)
	if err != nil {
		log.Printf("%v", err)
		return SMTPEmailResponse{}, nil, err
	}

	if !permissions.AccessToObject(emailSetting.CreatedBy, smtpUser.Id) && !currentUser.Data.IsAdmin {
		log.Printf("%v", err)
		return SMTPEmailResponse{}, nil, err
	}

	SMTPPassword := string(smtpUser.Data.SMTPPassword[:])

	client := http.Client{}
	getUrl := "https://tabulae-smtp.newsai.org/verify"

	verifyEmailRequest := models.SMTPSettings{}

	verifyEmailRequest.Servername = emailSetting.SMTPServer + ":" + strconv.Itoa(emailSetting.SMTPPortSSL)
	verifyEmailRequest.EmailUser = smtpUser.Data.SMTPUsername
	verifyEmailRequest.EmailPassword = SMTPPassword

	VerifyEmailRequest, err := json.Marshal(verifyEmailRequest)
	if err != nil {
		log.Printf("%v", err)
		return SMTPEmailResponse{}, nil, err
	}

	log.Printf("%v", string(VerifyEmailRequest))
	verifyEmailQuery := bytes.NewReader(VerifyEmailRequest)

	req, _ := http.NewRequest("POST", getUrl, verifyEmailQuery)
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("%v", err)
		return SMTPEmailResponse{}, nil, err
	}
	defer resp.Body.Close()

	decoder := json.NewDecoder(resp.Body)
	var verifyResponse SMTPEmailResponse
	err = decoder.Decode(&verifyResponse)
	if err != nil {
		log.Printf("%v", err)
		return SMTPEmailResponse{}, nil, err
	}

	if verifyResponse.Status {
		smtpUser.Data.SMTPValid = true
		controllers.SaveUser(r, &smtpUser)
	}

	return verifyResponse, nil, nil
}

func GetEmailSettingDetails(r *http.Request, id string) (SMTPEmailResponse, interface{}, error) {
	emailSetting, _, err := GetEmailSetting(r, id)
	if err != nil {
		log.Printf("%v", err)
		return SMTPEmailResponse{}, nil, err
	}

	currentUser, err := controllers.GetCurrentUser(r)
	if err != nil {
		log.Printf("%v", err)
		return SMTPEmailResponse{}, nil, err
	}

	smtpUser, _, err := controllers.GetUserById(r, emailSetting.CreatedBy)
	if err != nil {
		log.Printf("%v", err)
		return SMTPEmailResponse{}, nil, err
	}

	if !permissions.AccessToObject(emailSetting.CreatedBy, smtpUser.Id) && !currentUser.Data.IsAdmin {
		log.Printf("%v", err)
		return SMTPEmailResponse{}, nil, err
	}

	SMTPPassword := string(smtpUser.Data.SMTPPassword[:])
	userPassword, err := encrypt.DecryptString(SMTPPassword)
	if err != nil {
		return SMTPEmailResponse{}, nil, err
	}

	log.Printf("%v", smtpUser.Data.SMTPUsername)
	log.Printf("%v", userPassword)

	return SMTPEmailResponse{}, nil, err
}
