package controllers

import (
	"errors"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"time"

	gcontext "github.com/gorilla/context"
	"github.com/pquerna/ffjson/ffjson"

	"github.com/news-ai/api-v1/controllers"
	apiModels "github.com/news-ai/api-v1/models"

	"github.com/news-ai/tabulae-v1/models"
	"github.com/news-ai/tabulae-v1/search"
	"github.com/news-ai/tabulae-v1/sync"

	"github.com/news-ai/web/permissions"
	"github.com/news-ai/web/utilities"
)

type cancelEmailsBulk struct {
	Emails []int64 `json:"emails"`
}

/*
* Private methods
 */

/*
* Get methods
 */

func getEmail(r *http.Request, id int64) (models.Email, error) {
	if id == 0 {
		return models.Email{}, errors.New("datastore: no such entity")
	}
	// Get the email by id
	var email models.Email
	emailId := datastore.NewKey("Email", "", id, nil)
	err := nds.Get(emailId, &email)
	if err != nil {
		log.Errorf("%v", err)
		return models.Email{}, err
	}

	if !email.Created.IsZero() {
		email.Format(emailId, "emails")

		user, err := controllers.GetCurrentUser(r)
		if err != nil {
			log.Errorf("%v", err)
			return models.Email{}, errors.New("Could not get user")
		}

		if !permissions.AccessToObject(email.CreatedBy, user.Id) && !user.IsAdmin {
			return models.Email{}, errors.New("Forbidden")
		}

		return email, nil
	}

	return models.Email{}, errors.New("No email by this id")
}

func getEmailUnauthorized(r *http.Request, id int64) (models.Email, error) {
	if id == 0 {
		return models.Email{}, errors.New("datastore: no such entity")
	}
	// Get the email by id
	var email models.Email
	emailId := datastore.NewKey("Email", "", id, nil)
	err := nds.Get(emailId, &email)
	if err != nil {
		log.Errorf("%v", err)
		return models.Email{}, err
	}

	if !email.Created.IsZero() {
		email.Format(emailId, "emails")
		return email, nil
	}

	return models.Email{}, errors.New("No email by this id")
}

func getEmailUnauthorizedBulk(r *http.Request, ids []int64) ([]models.Email, error) {
	var ks []*datastore.Key

	for i := 0; i < len(ids); i++ {
		emailKey := datastore.NewKey("Email", "", ids[i], nil)
		ks = append(ks, emailKey)
	}

	var emails []models.Email
	emails = make([]models.Email, len(ks))
	err := nds.GetMulti(ks, emails)
	if err != nil {
		log.Errorf("%v", err)
		return []models.Email{}, err
	}

	for i := 0; i < len(emails); i++ {
		emails[i].Format(ks[i], "emails")
	}

	return emails, nil
}

/*
* Filter methods
 */

func filterEmail(queryType, query string) (models.Email, error) {
	// Get a publication by the URL
	ks, err := datastore.NewQuery("Email").Filter(queryType+" =", query).KeysOnly().GetAll(nil)
	if err != nil {
		log.Errorf("%v", err)
		return models.Email{}, err
	}

	if len(ks) == 0 {
		return models.Email{}, errors.New("No email by the field " + queryType)
	}

	var emails []models.Email
	emails = make([]models.Email, len(ks))
	err = nds.GetMulti(ks, emails)
	if err != nil {
		log.Errorf("%v", err)
		return models.Email{}, err
	}

	if len(emails) > 0 {
		emails[0].Format(ks[0], "emails")
		return emails[0], nil
	}

	return models.Email{}, errors.New("No email by this " + queryType)
}

func filterEmailbyListId(r *http.Request, listId int64) ([]models.Email, int, error) {
	emails := []models.Email{}

	user, err := controllers.GetCurrentUser(r)
	if err != nil {
		log.Errorf("%v", err)
		return []models.Email{}, 0, err
	}

	query := datastore.NewQuery("Email").Filter("CreatedBy =", user.Id).Filter("ListId =", listId)
	query = controllers.ConstructQuery(query, r)
	ks, err := query.KeysOnly().GetAll(nil)
	if err != nil {
		log.Errorf("%v", err)
		return []models.Email{}, 0, err
	}

	emails = make([]models.Email, len(ks))
	err = nds.GetMulti(ks, emails)
	if err != nil {
		log.Errorf("%v", err)
		return []models.Email{}, 0, err
	}

	for i := 0; i < len(emails); i++ {
		emails[i].Format(ks[i], "emails")
	}

	return emails, len(emails), nil
}

func filterOrderedEmailbyContactId(r *http.Request, contact models.Contact) ([]models.Email, error) {
	emails := []models.Email{}

	query := datastore.NewQuery("Email").Filter("CreatedBy =", contact.CreatedBy).Filter("To =", contact.Email).Filter("IsSent =", true).Filter("Cancel =", false).Filter("Archived =", false).Order("-Created")
	ks, err := query.KeysOnly().GetAll(nil)
	if err != nil {
		log.Errorf("%v", err)
		return []models.Email{}, err
	}

	emails = make([]models.Email, len(ks))
	err = nds.GetMulti(ks, emails)
	if err != nil {
		log.Errorf("%v", err)
		return []models.Email{}, err
	}

	for i := 0; i < len(emails); i++ {
		emails[i].Format(ks[i], "emails")
	}

	return emails, nil
}

func filterEmailbyContactId(r *http.Request, contactId int64) ([]models.Email, error) {
	emails := []models.Email{}

	user, err := controllers.GetCurrentUser(r)
	if err != nil {
		log.Errorf("%v", err)
		return []models.Email{}, err
	}

	query := datastore.NewQuery("Email").Filter("CreatedBy =", user.Id).Filter("ContactId =", contactId).Filter("IsSent =", true)
	query = controllers.ConstructQuery(query, r)
	ks, err := query.KeysOnly().GetAll(nil)
	if err != nil {
		log.Errorf("%v", err)
		return []models.Email{}, err
	}

	emails = make([]models.Email, len(ks))
	err = nds.GetMulti(ks, emails)
	if err != nil {
		log.Errorf("%v", err)
		return []models.Email{}, err
	}

	for i := 0; i < len(emails); i++ {
		emails[i].Format(ks[i], "emails")
	}

	return emails, nil
}

func filterEmailbyContactEmail(r *http.Request, email string) ([]models.Email, error) {
	emails := []models.Email{}

	user, err := controllers.GetCurrentUser(r)
	if err != nil {
		log.Errorf("%v", err)
		return []models.Email{}, err
	}

	query := datastore.NewQuery("Email").Filter("CreatedBy =", user.Id).Filter("To =", email).Filter("IsSent =", true)
	query = controllers.ConstructQuery(query, r)
	ks, err := query.KeysOnly().GetAll(nil)
	if err != nil {
		log.Errorf("%v", err)
		return []models.Email{}, err
	}

	emails = make([]models.Email, len(ks))
	err = nds.GetMulti(ks, emails)
	if err != nil {
		log.Errorf("%v", err)
		return []models.Email{}, err
	}

	for i := 0; i < len(emails); i++ {
		emails[i].Format(ks[i], "emails")
	}

	return emails, nil
}

func emailsToLists(r *http.Request, emails []models.Email) []models.MediaList {
	mediaListIds := []int64{}

	for i := 0; i < len(emails); i++ {
		if emails[i].ListId != 0 {
			mediaListIds = append(mediaListIds, emails[i].ListId)
		}
	}

	// Work on includes
	mediaLists := []models.MediaList{}
	mediaListExists := map[int64]bool{}
	mediaListExists = make(map[int64]bool)

	for i := 0; i < len(mediaListIds); i++ {
		if _, ok := mediaListExists[mediaListIds[i]]; !ok {
			if mediaListIds[i] != 0 {
				mediaList, _ := getMediaList(r, mediaListIds[i])
				mediaLists = append(mediaLists, mediaList)
				mediaListExists[mediaListIds[i]] = true
			}
		}
	}

	return mediaLists
}

func emailsToContacts(r *http.Request, emails []models.Email) []models.Contact {
	contactIds := []int64{}

	for i := 0; i < len(emails); i++ {
		if emails[i].ContactId != 0 {
			contactIds = append(contactIds, emails[i].ContactId)
		}
	}

	// Work on includes
	contacts := []models.Contact{}
	contactExists := map[int64]bool{}
	contactExists = make(map[int64]bool)

	for i := 0; i < len(contactIds); i++ {
		if _, ok := contactExists[contactIds[i]]; !ok {
			if contactIds[i] != 0 {
				contact, _ := getContact(r, contactIds[i])
				contacts = append(contacts, contact)
				contactExists[contactIds[i]] = true
			}
		}
	}

	return contacts
}

func sendEmail(r *http.Request, email models.Email) (models.Email, error) {
	user, err := controllers.GetCurrentUser(r)
	if err != nil {
		log.Errorf("%v", err)
		return email, err
	}

	if !user.EmailConfirmed {
		return email, errors.New("Users email is not confirmed - the user cannot send emails.")
	}

	// Check if email is already sent
	if email.IsSent {
		return email, errors.New("Email has already been sent.")
	}

	// Validate if HTML is valid
	validHTML := utilities.ValidateHTML(email.Body)
	if !validHTML {
		return email, errors.New("Invalid HTML")
	}

	if email.Subject == "" {
		email.Subject = "(no subject)"
	}

	userEmails := map[string]bool{}
	for i := 0; i < len(user.Emails); i++ {
		userEmails[user.Emails[i]] = true
	}

	emailId := strconv.FormatInt(email.Id, 10)
	email.Body = utilities.AppendHrefWithLink(email.Body, emailId, "https://email2.newsai.co/a")
	email.Body += "<img src=\"https://email2.newsai.co/?id=" + emailId + "\" alt=\"NewsAI\" />"
	email.IsSent = true

	// Check if the user's email is valid for sending
	if email.Method == "sendgrid" && email.FromEmail != "" {
		userEmailValid := false
		if user.Email == email.FromEmail {
			userEmailValid = true
		}

		if _, ok := userEmails[email.FromEmail]; ok {
			userEmailValid = true
		}

		if !userEmailValid {
			return models.Email{}, errors.New("The email requested is not confirmed by the user yet")
		}
	}

	return email, nil
}

/*
* Public methods
 */

/*
* Get methods
 */

func GetEmails(r *http.Request) ([]models.Email, interface{}, int, int, error) {
	emails := []models.Email{}

	user, err := controllers.GetCurrentUser(r)
	if err != nil {
		log.Errorf("%v", err)
		return []models.Email{}, nil, 0, 0, err
	}

	query := datastore.NewQuery("Email").Filter("CreatedBy =", user.Id)
	query = controllers.ConstructQuery(query, r)
	ks, err := query.KeysOnly().GetAll(nil)
	if err != nil {
		log.Errorf("%v", err)
		return []models.Email{}, nil, 0, 0, err
	}

	emails = make([]models.Email, len(ks))
	err = nds.GetMulti(ks, emails)
	if err != nil {
		log.Errorf("%v", err)
		return []models.Email{}, nil, 0, 0, err
	}

	for i := 0; i < len(emails); i++ {
		emails[i].Format(ks[i], "emails")
	}

	// Add includes
	mediaLists := emailsToLists(r, emails)
	contacts := emailsToContacts(r, emails)
	includes := make([]interface{}, len(mediaLists)+len(contacts))
	for i := 0; i < len(mediaLists); i++ {
		includes[i] = mediaLists[i]
	}

	for i := 0; i < len(contacts); i++ {
		includes[i+len(mediaLists)] = contacts[i]
	}

	return emails, includes, len(emails), 0, nil
}

func GetSentEmails(r *http.Request) ([]models.Email, interface{}, int, int, error) {
	emails := []models.Email{}

	user, err := controllers.GetCurrentUser(r)
	if err != nil {
		log.Errorf("%v", err)
		return []models.Email{}, nil, 0, 0, err
	}

	// Check if in the memcache there is a userid_emailAddress => timeLatsEmailSent

	// Filter all emails that are in the future (scheduled for later)
	query := datastore.NewQuery("Email").Filter("CreatedBy =", user.Id).Filter("IsSent =", true).Filter("Cancel =", false).Filter("Delievered =", true).Filter("Archived =", false)
	query = controllers.ConstructQuery(query, r)
	ks, err := query.KeysOnly().GetAll(nil)
	if err != nil {
		log.Errorf("%v", err)
		return []models.Email{}, nil, 0, 0, err
	}

	emails = make([]models.Email, len(ks))
	err = nds.GetMulti(ks, emails)
	if err != nil {
		log.Errorf("%v", err)
		return []models.Email{}, nil, 0, 0, err
	}

	for i := 0; i < len(emails); i++ {
		emails[i].Format(ks[i], "emails")
	}

	// Add includes
	mediaLists := emailsToLists(r, emails)
	contacts := emailsToContacts(r, emails)
	includes := make([]interface{}, len(mediaLists)+len(contacts))
	for i := 0; i < len(mediaLists); i++ {
		includes[i] = mediaLists[i]
	}

	for i := 0; i < len(contacts); i++ {
		includes[i+len(mediaLists)] = contacts[i]
	}

	return emails, includes, len(emails), 0, nil
}

func GetEmailStats(r *http.Request) (interface{}, interface{}, int, int, error) {
	user, err := controllers.GetCurrentUser(r)
	if err != nil {
		log.Errorf("%v", err)
		return nil, nil, 0, 0, err
	}

	timeseriesData, count, total, err := search.SearchEmailTimeseriesByUserId(r, user)
	return timeseriesData, nil, count, total, err
}

func GetScheduledEmails(r *http.Request) ([]models.Email, interface{}, int, int, error) {
	emails := []models.Email{}

	user, err := controllers.GetCurrentUser(r)
	if err != nil {
		log.Errorf("%v", err)
		return []models.Email{}, nil, 0, 0, err
	}

	// Filter all emails that are in the future (scheduled for later)
	query := datastore.NewQuery("Email").Filter("CreatedBy =", user.Id).Filter("SendAt >=", time.Now()).Filter("Cancel =", false).Filter("IsSent =", true)
	query = controllers.ConstructQuery(query, r)
	ks, err := query.KeysOnly().GetAll(nil)
	if err != nil {
		log.Errorf("%v", err)
		return []models.Email{}, nil, 0, 0, err
	}

	queryNoLimit := datastore.NewQuery("Email").Filter("CreatedBy =", user.Id).Filter("SendAt >=", time.Now()).Filter("Cancel =", false).Filter("IsSent =", true)
	amountOfKeys, err := queryNoLimit.KeysOnly().GetAll(nil)
	if err != nil {
		log.Errorf("%v", err)
		return []models.Email{}, nil, 0, 0, err
	}

	emails = make([]models.Email, len(ks))
	err = nds.GetMulti(ks, emails)
	if err != nil {
		log.Errorf("%v", err)
		return []models.Email{}, nil, 0, 0, err
	}

	for i := 0; i < len(emails); i++ {
		emails[i].Format(ks[i], "emails")
	}

	// Add includes
	mediaLists := emailsToLists(r, emails)
	contacts := emailsToContacts(r, emails)
	includes := make([]interface{}, len(mediaLists)+len(contacts))
	for i := 0; i < len(mediaLists); i++ {
		includes[i] = mediaLists[i]
	}

	for i := 0; i < len(contacts); i++ {
		includes[i+len(mediaLists)] = contacts[i]
	}

	return emails, includes, len(emails), len(amountOfKeys), nil
}

func GetArchivedEmails(r *http.Request) ([]models.Email, interface{}, int, int, error) {
	emails := []models.Email{}

	user, err := controllers.GetCurrentUser(r)
	if err != nil {
		log.Errorf("%v", err)
		return []models.Email{}, nil, 0, 0, err
	}

	// Filter all emails that are in the future (scheduled for later)
	query := datastore.NewQuery("Email").Filter("CreatedBy =", user.Id).Filter("Cancel =", false).Filter("IsSent =", true).Filter("Archived =", true)
	query = controllers.ConstructQuery(query, r)
	ks, err := query.KeysOnly().GetAll(nil)
	if err != nil {
		log.Errorf("%v", err)
		return []models.Email{}, nil, 0, 0, err
	}

	emails = make([]models.Email, len(ks))
	err = nds.GetMulti(ks, emails)
	if err != nil {
		log.Errorf("%v", err)
		return []models.Email{}, nil, 0, 0, err
	}

	for i := 0; i < len(emails); i++ {
		emails[i].Format(ks[i], "emails")
	}

	// Add includes
	mediaLists := emailsToLists(r, emails)
	contacts := emailsToContacts(r, emails)
	includes := make([]interface{}, len(mediaLists)+len(contacts))
	for i := 0; i < len(mediaLists); i++ {
		includes[i] = mediaLists[i]
	}

	for i := 0; i < len(contacts); i++ {
		includes[i+len(mediaLists)] = contacts[i]
	}

	return emails, includes, len(emails), 0, nil
}

func GetTeamEmails(r *http.Request) ([]models.Email, interface{}, int, int, error) {
	return []models.Email{}, nil, 0, 0, nil
}

func GetEmailById(r *http.Request, id int64) (models.Email, error) {
	email, err := getEmail(r, id)
	if err != nil {
		log.Errorf("%v", err)
		return models.Email{}, err
	}
	return email, nil
}

func GetEmailUnauthorizedBulk(r *http.Request, ids []int64) ([]models.Email, interface{}, error) {
	email, err := getEmailUnauthorizedBulk(r, ids)
	if err != nil {
		log.Errorf("%v", err)
		return []models.Email{}, nil, err
	}

	return email, nil, nil
}

func GetEmailUnauthorized(r *http.Request, id string) (models.Email, interface{}, error) {
	// Get the details of the current user
	currentId, err := utilities.StringIdToInt(id)
	if err != nil {
		log.Errorf("%v", err)
		return models.Email{}, nil, err
	}

	email, err := getEmailUnauthorized(r, currentId)
	if err != nil {
		log.Errorf("%v", err)
		return models.Email{}, nil, err
	}

	return email, nil, nil
}

func GetEmailByIdUnauthorized(r *http.Request, id int64) (models.Email, interface{}, error) {
	email, err := getEmailUnauthorized(r, id)
	if err != nil {
		log.Errorf("%v", err)
		return models.Email{}, nil, err
	}

	return email, nil, nil
}

func GetEmail(r *http.Request, id string) (models.Email, interface{}, error) {
	// Get the details of the current user
	currentId, err := utilities.StringIdToInt(id)
	if err != nil {
		log.Errorf("%v", err)
		return models.Email{}, nil, err
	}

	email, err := getEmail(r, currentId)
	if err != nil {
		log.Errorf("%v", err)
		return models.Email{}, nil, err
	}

	includedFiles := []models.File{}
	includedContact := []models.Contact{}
	if len(email.Attachments) > 0 {
		for i := 0; i < len(email.Attachments); i++ {
			file, err := getFile(r, email.Attachments[i])
			if err == nil {
				includedFiles = append(includedFiles, file)
			} else {
				log.Errorf("%v", err)
			}
		}
	}

	if email.ContactId != 0 {
		contact, err := getContact(r, email.ContactId)
		if err != nil {
			log.Errorf("%v", err)
			return models.Email{}, nil, err
		}
		includedContact = append(includedContact, contact)
	}

	includes := make([]interface{}, len(includedFiles)+len(includedContact))

	for i := 0; i < len(includedFiles); i++ {
		includes[i] = includedFiles[i]
	}

	for i := 0; i < len(includedContact); i++ {
		includes[i+len(includedFiles)] = includedContact[i]
	}

	return email, includes, nil
}

/*
* Create methods
 */

func CreateEmailTransition(r *http.Request) ([]models.Email, interface{}, error) {
	buf, _ := ioutil.ReadAll(r.Body)

	currentUser, err := controllers.GetCurrentUser(r)
	if err != nil {
		log.Errorf("%v", err)
		return []models.Email{}, nil, err
	}

	// Figure out what the emailMethod we should use
	emailMethod := "sendgrid"
	if currentUser.SMTPValid && currentUser.ExternalEmail && currentUser.EmailSetting != 0 {
		emailMethod = "smtp"
	} else if currentUser.AccessToken != "" && currentUser.Gmail {
		emailMethod = "gmail"
	} else if currentUser.OutlookAccessToken != "" && currentUser.Outlook {
		emailMethod = "outlook"
	} else if currentUser.UseSparkPost {
		emailMethod = "sparkpost"
	}

	decoder := ffjson.NewDecoder()
	var email models.Email
	err = decoder.Decode(buf, &email)

	// If it is an array and you need to do BATCH processing
	if err != nil {
		var emails []models.Email

		arrayDecoder := ffjson.NewDecoder()
		err = arrayDecoder.Decode(buf, &emails)

		if err != nil {
			log.Errorf("%v", err)
			return []models.Email{}, nil, err
		}

		var keys []*datastore.Key
		emailIds := []int64{}

		for i := 0; i < len(emails); i++ {
			// Test if the email we are sending with is in the user's SendGridFrom or is their Email
			// Only valid if user is not using gmail, outlook, or smtp
			if emails[i].FromEmail != "" && !currentUser.Gmail && !currentUser.Outlook && !currentUser.ExternalEmail {
				userEmailValid := false
				if currentUser.Email == emails[i].FromEmail {
					userEmailValid = true
				}

				for x := 0; x < len(currentUser.Emails); x++ {
					if currentUser.Emails[x] == emails[i].FromEmail {
						userEmailValid = true
					}
				}

				// If this is if the email added is not valid in SendGridFrom
				if !userEmailValid {
					return []models.Email{}, nil, errors.New("The email requested is not confirmed by the user yet")
				}
			}

			emails[i].Id = 0
			emails[i].CreatedBy = currentUser.Id
			emails[i].Created = time.Now()
			emails[i].Updated = time.Now()
			emails[i].TeamId = currentUser.TeamId
			emails[i].IsSent = false
			emails[i].Method = emailMethod

			keys = append(keys, emails[i].Key(c))
		}

		if len(keys) < 300 {
			ks := []*datastore.Key{}
			err = nds.RunInTransaction(func(ctx context.Context) error {
				contextWithTimeout, _ := context.WithTimeout(time.Second * 150)
				ks, err = nds.PutMulti(contextWithTimeout, keys, emails)
				if err != nil {
					log.Errorf("%v", err)
					return err
				}
				return nil
			}, nil)

			for i := 0; i < len(ks); i++ {
				emails[i].Format(ks[i], "emails")
				emailIds = append(emailIds, emails[i].Id)
			}

			sync.EmailResourceBulkSync(r, emailIds)
			return emails, nil, err
		} else {
			firstHalfKeys := []*datastore.Key{}
			secondHalfKeys := []*datastore.Key{}
			thirdHalfKeys := []*datastore.Key{}
			fourHalfKeys := []*datastore.Key{}

			startOne := 0
			endOne := 100

			startTwo := 100
			endTwo := 200

			startThree := 200
			endThree := 300

			startFour := 300
			endFour := len(keys)

			err1 := nds.RunInTransaction(func(ctx context.Context) error {
				contextWithTimeout, _ := context.WithTimeout(time.Second * 150)
				firstHalfKeys, err = nds.PutMulti(contextWithTimeout, keys[startOne:endOne], emails[startOne:endOne])
				if err != nil {
					log.Errorf("%v", err)
					return err
				}
				return nil
			}, nil)

			err2 := nds.RunInTransaction(func(ctx context.Context) error {
				contextWithTimeout, _ := context.WithTimeout(time.Second * 150)
				secondHalfKeys, err = nds.PutMulti(contextWithTimeout, keys[startTwo:endTwo], emails[startTwo:endTwo])
				if err != nil {
					log.Errorf("%v", err)
					return err
				}
				return nil
			}, nil)

			err3 := nds.RunInTransaction(func(ctx context.Context) error {
				contextWithTimeout, _ := context.WithTimeout(time.Second * 150)
				thirdHalfKeys, err = nds.PutMulti(contextWithTimeout, keys[startThree:endThree], emails[startThree:endThree])
				if err != nil {
					log.Errorf("%v", err)
					return err
				}
				return nil
			}, nil)

			err4 := nds.RunInTransaction(func(ctx context.Context) error {
				contextWithTimeout, _ := context.WithTimeout(time.Second * 150)
				fourHalfKeys, err = nds.PutMulti(contextWithTimeout, keys[startFour:endFour], emails[startFour:endFour])
				if err != nil {
					log.Errorf("%v", err)
					return err
				}
				return nil
			}, nil)

			firstHalfKeys = append(firstHalfKeys, secondHalfKeys...)
			firstHalfKeys = append(firstHalfKeys, thirdHalfKeys...)
			firstHalfKeys = append(firstHalfKeys, fourHalfKeys...)

			for i := 0; i < len(firstHalfKeys); i++ {
				emails[i].Format(firstHalfKeys[i], "emails")
				emailIds = append(emailIds, emails[i].Id)
			}

			if err1 != nil {
				err = err1
			}

			if err2 != nil {
				err = err2
			}

			if err3 != nil {
				err = err3
			}

			if err4 != nil {
				err = err4
			}

			sync.EmailResourceBulkSync(r, emailIds)
			return emails, nil, err
		}
	}

	// Test if the email we are sending with is in the user's SendGridFrom or is their Email
	if email.FromEmail != "" {
		userEmailValid := false
		if currentUser.Email == email.FromEmail {
			userEmailValid = true
		}

		for i := 0; i < len(currentUser.Emails); i++ {
			if currentUser.Emails[i] == email.FromEmail {
				userEmailValid = true
			}
		}

		// If this is if the email added is not valid in SendGridFrom
		if !userEmailValid {
			return []models.Email{}, nil, errors.New("The email requested is not confirmed by you yet")
		}
	}

	email.CreatedBy = currentUser.Id
	email.Updated = time.Now()
	email.Created = time.Now()
	email.Method = emailMethod
	email.TeamId = currentUser.TeamId
	email.IsSent = false

	// Create email
	_, err = email.Create(r, currentUser)
	sync.ResourceSync(r, email.Id, "Email", "create")
	if err != nil {
		log.Errorf("%v", err)
		return []models.Email{}, nil, err
	}
	return []models.Email{email}, nil, nil
}

/*
* Filter methods
 */

func FilterEmailBySendGridID(sendGridId string) (models.Email, error) {
	// Get the id of the current email
	email, err := filterEmail("SendGridId", sendGridId)
	if err != nil {
		log.Errorf("%v", err)
		return models.Email{}, err
	}
	return email, nil
}

/*
* Update methods
 */

func UpdateEmail(r *http.Request, currentUser apiModels.User, email *models.Email, updatedEmail models.Email) (models.Email, interface{}, error) {
	if email.CreatedBy != currentUser.Id {
		return *email, nil, errors.New("You don't have permissions to edit this object")
	}

	utilities.UpdateIfNotBlank(&email.Subject, updatedEmail.Subject)
	utilities.UpdateIfNotBlank(&email.Body, updatedEmail.Body)
	utilities.UpdateIfNotBlank(&email.To, updatedEmail.To)

	email.CC = updatedEmail.CC
	email.BCC = updatedEmail.BCC

	if updatedEmail.ListId != 0 {
		email.ListId = updatedEmail.ListId
	}

	if updatedEmail.TemplateId != 0 {
		email.TemplateId = updatedEmail.TemplateId
	}

	email.Save(c)
	sync.ResourceSync(r, email.Id, "Email", "create")
	return *email, nil, nil
}

func UpdateSingleEmail(r *http.Request, id string) (models.Email, interface{}, error) {
	// Get the details of the current email
	email, _, err := GetEmail(r, id)
	if err != nil {
		log.Errorf("%v", err)
		return models.Email{}, nil, err
	}

	user, err := controllers.GetCurrentUser(r)
	if err != nil {
		log.Errorf("%v", err)
		return models.Email{}, nil, errors.New("Could not get user")
	}

	if !permissions.AccessToObject(email.CreatedBy, user.Id) {
		return models.Email{}, nil, errors.New("Forbidden")
	}

	decoder := ffjson.NewDecoder()
	var updatedEmail models.Email
	buf, _ := ioutil.ReadAll(r.Body)
	err = decoder.Decode(buf, &updatedEmail)
	if err != nil {
		log.Errorf("%v", err)
		return models.Email{}, nil, err
	}

	return UpdateEmail(r, user, &email, updatedEmail)
}

func UpdateBatchEmail(r *http.Request) ([]models.Email, interface{}, error) {
	decoder := ffjson.NewDecoder()
	var updatedEmails []models.Email
	buf, _ := ioutil.ReadAll(r.Body)
	err := decoder.Decode(buf, &updatedEmails)
	if err != nil {
		log.Errorf("%v", err)
		return []models.Email{}, nil, err
	}

	// Get logged in user
	user, err := controllers.GetCurrentUser(r)
	if err != nil {
		log.Errorf("%v", err)
		return []models.Email{}, nil, errors.New("Could not get user")
	}

	currentEmails := []models.Email{}
	for i := 0; i < len(updatedEmails); i++ {
		email, err := getEmail(r, updatedEmails[i].Id)
		if err != nil {
			log.Errorf("%v", err)
			return []models.Email{}, nil, err
		}

		if !permissions.AccessToObject(email.CreatedBy, user.Id) {
			return []models.Email{}, nil, errors.New("Forbidden")
		}

		currentEmails = append(currentEmails, email)
	}

	newEmails := []models.Email{}
	for i := 0; i < len(updatedEmails); i++ {
		updatedEmail, _, err := UpdateEmail(r, user, &currentEmails[i], updatedEmails[i])
		if err != nil {
			log.Errorf("%v", err)
			return []models.Email{}, nil, err
		}
		newEmails = append(newEmails, updatedEmail)
	}

	return newEmails, nil, nil
}

/*
* Action methods
 */

func CancelAllScheduled(r *http.Request) ([]models.Email, interface{}, int, int, error) {
	emails := []models.Email{}

	user, err := controllers.GetCurrentUser(r)
	if err != nil {
		log.Errorf("%v", err)
		return []models.Email{}, nil, 0, 0, err
	}

	// Filter all emails that are in the future (scheduled for later)
	query := datastore.NewQuery("Email").Filter("CreatedBy =", user.Id).Filter("SendAt >=", time.Now()).Filter("Cancel =", false).Filter("IsSent =", true)
	ks, err := query.KeysOnly().GetAll(nil)
	if err != nil {
		log.Errorf("%v", err)
		return []models.Email{}, nil, 0, 0, err
	}

	emails = make([]models.Email, len(ks))
	err = nds.GetMulti(ks, emails)
	if err != nil {
		log.Errorf("%v", err)
		return []models.Email{}, nil, 0, 0, err
	}

	for i := 0; i < len(emails); i++ {
		emails[i].Format(ks[i], "emails")
	}

	emailIds := []int64{} // Validated email ids
	for i := 0; i < len(emails); i++ {
		// If it has not been delivered and has a sentat date then we can cancel it
		// and that sendAt date is in the future.
		if !emails[i].Delievered && !emails[i].SendAt.IsZero() && emails[i].SendAt.After(time.Now()) {
			emails[i].Cancel = true
			emails[i].Save(c)
			emailIds = append(emailIds, emails[i].Id)
		}
	}

	sync.EmailResourceBulkSync(r, emailIds)
	return emails, nil, len(emails), 0, nil
}

func BulkCancelEmail(r *http.Request) ([]models.Email, interface{}, int, int, error) {
	buf, _ := ioutil.ReadAll(r.Body)
	decoder := ffjson.NewDecoder()
	var cancelEmails cancelEmailsBulk
	err := decoder.Decode(buf, &cancelEmails)
	if err != nil {
		log.Errorf("%v", err)
		return []models.Email{}, nil, 0, 0, err
	}

	emails := []models.Email{}
	emailIds := []int64{} // Validated email ids
	for i := 0; i < len(cancelEmails.Emails); i++ {
		email, err := getEmail(r, cancelEmails.Emails[i])
		if err != nil {
			log.Errorf("%v", err)
			continue
		}

		// If it has not have a sentat date then we can cancel it
		// and that sendAt date is in the future.
		if !email.SendAt.IsZero() && email.SendAt.After(time.Now()) {
			email.Cancel = true
			email.Save(c)
			emails = append(emails, email)
			emailIds = append(emailIds, email.Id)
		}
	}

	sync.EmailResourceBulkSync(r, emailIds)
	return emails, nil, len(emails), 0, nil
}

func CancelEmail(r *http.Request, id string) (models.Email, interface{}, error) {
	email, _, err := GetEmail(r, id)
	if err != nil {
		log.Errorf("%v", err)
		return models.Email{}, nil, err
	}

	// If it has a sentat date then we can cancel it
	// and that sendAt date is in the future.
	if !email.SendAt.IsZero() && email.SendAt.After(time.Now()) {
		email.Cancel = true
		email.Save(c)
		sync.ResourceSync(r, email.Id, "Email", "create")
		return email, nil, nil
	}

	return email, nil, errors.New("Email has already been delivered")
}

func ArchiveEmail(r *http.Request, id string) (models.Email, interface{}, error) {
	email, _, err := GetEmail(r, id)
	if err != nil {
		log.Errorf("%v", err)
		return models.Email{}, nil, err
	}

	email.Archived = true
	email.Save(c)

	// Remove memcache object for the campaign
	memcacheKey := GetEmailCampaignKey(email)
	memcache.Delete(memcacheKey)

	sync.ResourceSync(r, email.Id, "Email", "create")
	return email, nil, nil
}

func BulkSendEmail(r *http.Request) ([]models.Email, interface{}, int, int, error) {
	buf, _ := ioutil.ReadAll(r.Body)
	decoder := ffjson.NewDecoder()
	var bulkEmailIds models.BulkSendEmailIds
	err := decoder.Decode(buf, &bulkEmailIds)
	if err != nil {
		log.Errorf("%v", err)
		return []models.Email{}, nil, 0, 0, err
	}

	user, err := controllers.GetCurrentUser(r)
	if err != nil {
		log.Errorf("%v", err)
		return []models.Email{}, nil, 0, 0, err
	}

	// If user is not active then they can't send
	// emails
	if !user.IsActive {
		return []models.Email{}, nil, 0, 0, err
	}

	if user.IsBanned {
		return []models.Email{}, nil, 0, 0, err
	}

	var keys []*datastore.Key
	updatedEmails := []models.Email{}
	emailIds := []int64{}
	memcacheKey := ""

	// Since the emails should be the same, get the attachments here
	if len(bulkEmailIds.EmailIds) > 0 {
		emails, err := getEmailUnauthorizedBulk(r, bulkEmailIds.EmailIds)
		if err != nil {
			return []models.Email{}, nil, 0, 0, err
		}

		for i := 0; i < len(emails); i++ {
			singleEmail, err := sendEmail(r, emails[i])
			if err != nil {
				log.Errorf("%v", err)
				continue
			}

			keys = append(keys, singleEmail.Key(c))
			updatedEmails = append(updatedEmails, singleEmail)

			// sentTime := ""

			// Check if email has been scheduled or not
			if singleEmail.SendAt.IsZero() || singleEmail.SendAt.Before(time.Now()) {
				memcacheKey = GetEmailCampaignKey(singleEmail)
				emailIds = append(emailIds, singleEmail.Id)
				// sentTime = singleEmail.Created.Format(time.RFC3339)
			}
			// else {
			// 	sentTime = singleEmail.SendAt.Format(time.RFC3339)
			// }

			// lastCreatedMemcacheKey := "lastcontacted" + strconv.FormatInt(user.Id, 10) + emails[i].To
			// item1 := &memcache.Item{
			// 	Key:   lastCreatedMemcacheKey,
			// 	Value: []byte(sentTime),
			// }
			// memcache.Set( item1)
		}

		ks := []*datastore.Key{}
		err = nds.RunInTransaction(func(ctx context.Context) error {
			contextWithTimeout, _ := context.WithTimeout(time.Second * 150)
			ks, err = nds.PutMulti(contextWithTimeout, keys, updatedEmails)
			if err != nil {
				log.Errorf("%v", err)
				return err
			}
			return nil
		}, nil)

		// Delete a single memcache key since the emails should all have
		// the same subject (or baseSubject)
		if memcacheKey != "" {
			memcache.Delete(memcacheKey)
		}

		if len(emailIds) > 0 {
			sync.SendEmailsToEmailService(r, emailIds)
		}
	}

	return updatedEmails, nil, len(updatedEmails), 0, nil
}

func SendEmail(r *http.Request, id string) (models.Email, interface{}, error) {
	email, _, err := GetEmail(r, id)
	if err != nil {
		log.Errorf("%v", err)
		return models.Email{}, nil, err
	}

	singleEmail, err := sendEmail(r, email)
	if err != nil {
		log.Errorf("%v", err)
		return models.Email{}, nil, err
	}
	singleEmail.Save(c)

	// Check if email has been scheduled or not
	if email.SendAt.IsZero() || email.SendAt.Before(time.Now()) {
		// Remove memcache key for this particular email campaign
		memcacheKey := GetEmailCampaignKey(email)
		if memcacheKey != "" {
			memcache.Delete(memcacheKey)
		}

		// Sync email with email service if this is not a bulk email
		emailIds := []int64{email.Id}
		sync.SendEmailsToEmailService(r, emailIds)
	}

	return singleEmail, nil, nil
}

func MarkBounced(r *http.Request, e *models.Email, reason string) (*models.Email, error) {
	controllers.SetUser(r, e.CreatedBy)

	contacts, err := filterContactByEmail(e.To)
	if err != nil {
		log.Infof("%v", err)
	}

	for i := 0; i < len(contacts); i++ {
		contacts[i].EmailBounced = true
		contacts[i].Save(r)
	}

	_, err = e.MarkBounced(reason)
	return e, err
}

func MarkSpam(r *http.Request, e *models.Email) (*models.Email, error) {
	controllers.SetUser(r, e.CreatedBy)
	_, err := e.MarkSpam(c)
	return e, err
}

func MarkClicked(r *http.Request, e *models.Email) (*models.Email, error) {
	controllers.SetUser(r, e.CreatedBy)
	_, err := e.MarkClicked(c)
	return e, err
}

func MarkDelivered(r *http.Request, e *models.Email) (*models.Email, error) {
	_, err := e.MarkDelivered(c)
	return e, err
}

func MarkOpened(r *http.Request, e *models.Email) (*models.Email, error) {
	controllers.SetUser(r, e.CreatedBy)
	_, err := e.MarkOpened(c)
	return e, err
}

func MarkSendgridOpen(r *http.Request, e *models.Email) (*models.Email, error) {
	controllers.SetUser(r, e.CreatedBy)
	_, err := e.MarkSendgridOpened(c)
	return e, err
}

func MarkSendgridDrop(r *http.Request, e *models.Email) (*models.Email, error) {
	controllers.SetUser(r, e.CreatedBy)
	_, err := e.MarkSendgridDropped(c)
	return e, err
}

func GetEmailLogs(r *http.Request, id string) (interface{}, interface{}, error) {
	email, _, err := GetEmail(r, id)
	if err != nil {
		log.Errorf("%v", err)
		return models.Email{}, nil, err
	}

	user, err := controllers.GetCurrentUser(r)
	if err != nil {
		log.Errorf("%v", err)
		return email, nil, err
	}

	logs, _, _, err := search.SearchEmailLogByEmailId(r, user, email.Id)
	return logs, nil, err
}

func GetEmailSearch(r *http.Request) (interface{}, interface{}, int, int, error) {
	queryField := gcontext.Get(r, "q").(string)

	if queryField == "" {
		return nil, nil, 0, 0, nil
	}

	user, err := controllers.GetCurrentUser(r)
	if err != nil {
		log.Errorf("%v", err)
		return nil, nil, 0, 0, err
	}

	if strings.Contains(queryField, "date:") || strings.Contains(queryField, "subject:") || strings.Contains(queryField, "filter:") || strings.Contains(queryField, "baseSubject:") {
		emailFilters := strings.Split(queryField, ",")
		emailDate := ""
		emailSubject := ""
		emailBaseSubject := ""
		emailFilter := ""

		for i := 0; i < len(emailFilters); i++ {
			if strings.Contains(emailFilters[i], "date:") {
				emailDateArray := strings.Split(emailFilters[i], ":")
				if len(emailDateArray) > 1 {
					emailDate = strings.Join(emailDateArray[1:], ":")
					emailDate = strings.Replace(emailDate, "\\", "", -1)

					if last := len(emailDate) - 1; last >= 0 && emailDate[last] == '"' {
						emailDate = emailDate[:last]
					}

					if emailDate[0] == '"' {
						emailDate = emailDate[1:]
					}
				}
			} else if strings.Contains(emailFilters[i], "filter:") {
				emailFilterArray := strings.Split(emailFilters[i], ":")
				if len(emailFilterArray) > 1 {
					emailFilter = strings.Join(emailFilterArray[1:], ":")
					emailFilter = strings.Replace(emailFilter, "\\", "", -1)

					if last := len(emailFilter) - 1; last >= 0 && emailFilter[last] == '"' {
						emailFilter = emailFilter[:last]
					}

					if emailFilter[0] == '"' {
						emailFilter = emailFilter[1:]
					}
				}
			} else if strings.Contains(emailFilters[i], "subject:") {
				if len(emailFilters) > 2 {
					emailSubjectSplit := strings.Split(queryField, "subject:")
					emailFilters[i] = "subject:" + emailSubjectSplit[len(emailSubjectSplit)-1]
				}
				emailSubjectArray := strings.Split(emailFilters[i], ":")
				if len(emailSubjectArray) > 1 {
					log.Infof("%v", emailSubjectArray)
					// Recover the pieces when split by colon
					emailSubject = strings.Join(emailSubjectArray[1:], ":")
					emailSubject = strings.Replace(emailSubject, "\\", "", -1)

					if last := len(emailSubject) - 1; last >= 0 && emailSubject[last] == '"' {
						emailSubject = emailSubject[:last]
					}

					if emailSubject[0] == '"' {
						emailSubject = emailSubject[1:]
					}

					log.Infof("%v", emailSubject)
				}
			} else if strings.Contains(emailFilters[i], "baseSubject:") {
				emailBaseSubjectArray := strings.Split(emailFilters[i], ":")
				if len(emailBaseSubjectArray) > 1 {
					// Recover the pieces when split by colon
					emailBaseSubject = strings.Join(emailBaseSubjectArray[1:], ":")
					emailBaseSubject = strings.Replace(emailBaseSubject, "\\", "", -1)

					if last := len(emailBaseSubject) - 1; last >= 0 && emailBaseSubject[last] == '"' {
						emailBaseSubject = emailBaseSubject[:last]
					}

					if emailBaseSubject[0] == '"' {
						emailBaseSubject = emailBaseSubject[1:]
					}
				}
			}
		}

		if emailDate != "" || emailSubject != "" || emailFilter != "" || emailBaseSubject != "" {
			emails, count, total, err := search.SearchEmailsByQueryFields(r, user, emailDate, emailSubject, emailBaseSubject, emailFilter)

			// Add includes
			mediaLists := emailsToLists(r, emails)
			contacts := emailsToContacts(r, emails)
			includes := make([]interface{}, len(mediaLists)+len(contacts))
			for i := 0; i < len(mediaLists); i++ {
				includes[i] = mediaLists[i]
			}

			for i := 0; i < len(contacts); i++ {
				includes[i+len(mediaLists)] = contacts[i]
			}

			return emails, includes, count, total, err
		} else {
			return nil, nil, 0, 0, errors.New("Please enter a valid date or subject")
		}
	}

	emails, count, total, err := search.SearchEmailsByQuery(r, user, queryField)

	// Add includes
	mediaLists := emailsToLists(r, emails)
	contacts := emailsToContacts(r, emails)
	includes := make([]interface{}, len(mediaLists)+len(contacts))
	for i := 0; i < len(mediaLists); i++ {
		includes[i] = mediaLists[i]
	}

	for i := 0; i < len(contacts); i++ {
		includes[i+len(mediaLists)] = contacts[i]
	}

	return emails, includes, count, total, err
}

func GetEmailCampaigns(r *http.Request) (interface{}, interface{}, int, int, error) {
	user, err := controllers.GetCurrentUser(r)
	if err != nil {
		log.Errorf("%v", err)
		return nil, nil, 0, 0, err
	}

	emails, count, total, err := search.SearchEmailCampaignsByDate(r, user)
	return emails, nil, count, total, err
}

func GetEmailCampaignsForUser(r *http.Request, id string) (interface{}, interface{}, int, int, error) {
	user := apiModels.User{}
	err := errors.New("")

	switch id {
	case "me":
		user, err = controllers.GetCurrentUser(r)
		if err != nil {
			log.Errorf("%v", err)
			return []models.Email{}, nil, 0, 0, err
		}
	default:
		userId, err := utilities.StringIdToInt(id)
		if err != nil {
			log.Errorf("%v", err)
			return []models.Email{}, nil, 0, 0, err
		}
		user, _, err = controllers.GetUserById(r, userId)
		if err != nil {
			log.Errorf("%v", err)
			return []models.Email{}, nil, 0, 0, err
		}
	}

	currentUser, err := controllers.GetCurrentUser(r)
	if err != nil {
		log.Errorf("%v", err)
		return []models.Email{}, nil, 0, 0, err
	}

	if !permissions.AccessToObject(user.Id, currentUser.Id) && !currentUser.IsAdmin {
		err = errors.New("Forbidden")
		log.Errorf("%v", err)
		return []models.Email{}, nil, 0, 0, err
	}

	emails, count, total, err := search.SearchEmailCampaignsByDate(r, user)
	return emails, nil, count, total, err
}

func GetEmailProviderLimits(r *http.Request) (interface{}, interface{}, error) {
	user, err := controllers.GetCurrentUser(r)
	if err != nil {
		log.Errorf("%v", err)
		return nil, nil, err
	}

	emailProviderLimits := models.EmailProviderLimits{}
	emailProviderLimits.SendGridLimits = 2000
	emailProviderLimits.OutlookLimits = 500
	emailProviderLimits.GmailLimits = 500
	emailProviderLimits.SMTPLimits = 2000

	t := time.Now()
	todayDateMorning := time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.Local)
	todayDateNight := time.Date(t.Year(), t.Month(), t.Day(), 23, 59, 59, 59, time.Local)

	// SendGrid
	sendGrid, err := datastore.NewQuery("Email").Filter("CreatedBy =", user.Id).Filter("Method =", "sendgrid").Filter("IsSent =", true).Filter("Delievered =", true).Filter("Created <=", todayDateNight).Filter("Created >=", todayDateMorning).KeysOnly().GetAll(nil)
	if err != nil {
		log.Errorf("%v", err)
		return nil, nil, err
	}
	emailProviderLimits.SendGrid = len(sendGrid)

	// Outlook

	return emailProviderLimits, nil, nil
}

func GetEmailCampaignKey(email models.Email) string {
	emailSubject := email.Subject
	if email.BaseSubject != "" {
		emailSubject = email.BaseSubject
	}

	userIdString := strconv.FormatInt(email.CreatedBy, 10)
	dayFormat := email.Created.Format("2006-01-02")

	// Generate campaign name in the way that the memcache wants it
	campaignName := utilities.RemoveSpecialCharacters(emailSubject)
	campaignName = strings.ToLower(campaignName)
	campaignName = strings.Trim(campaignName, " ")
	campaignName = strings.Replace(campaignName, " ", "-", -1)

	memcacheKey := userIdString + "-" + dayFormat + "-" + campaignName
	return memcacheKey
}
