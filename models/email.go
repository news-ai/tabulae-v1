package models

import (
	"log"
	"net/http"
	"time"

	"github.com/news-ai/api-v1/db"
	apiModels "github.com/news-ai/api-v1/models"
)

type EmailProviderLimits struct {
	SendGrid       int `json:"sendgrid"`
	SendGridLimits int `json:"sendgridLimits"`
	Outlook        int `json:"outlook"`
	OutlookLimits  int `json:"outlookLimits"`
	Gmail          int `json:"gmail"`
	GmailLimits    int `json:"gmailLimits"`
	SMTP           int `json:"smtp"`
	SMTPLimits     int `json:"smtpLimits"`
}

type BulkSendEmailIds struct {
	EmailIds []int64 `json:"emailids"`
}

type SMTPSettings struct {
	Servername string `json:"servername"`

	EmailUser     string `json:"emailuser"`
	EmailPassword string `json:"emailpassword"`
}

type SMTPEmailSettings struct {
	Servername string `json:"servername"`

	EmailUser     string `json:"emailuser"`
	EmailPassword string `json:"emailpassword"`

	To      string `json:"to"`
	Subject string `json:"subject"`
	Body    string `json:"body"`
}

type UserEmailSetting struct {
	SMTPUsername string `json:"smtpusername"`
	SMTPPassword string `json:"smtppassword"`
}

type EmailSetting struct {
	apiModels.Base

	SMTPServer  string `json:"SMTPServer"`
	SMTPPortTLS int    `json:"SMTPPortTLS"`
	SMTPPortSSL int    `json:"SMTPPortSSL"`
	SMTPSSLTLS  bool   `json:"SMTPSSLTLS"`

	IMAPServer  string `json:"IMAPServer"`
	IMAPPortTLS int    `json:"IMAPPortTLS"`
	IMAPPortSSL int    `json:"IMAPPortSSL"`
	IMAPSSLTLS  bool   `json:"IMAPSSLTLS"`
}

type Email struct {
	apiModels.Base

	Method string `json:"method"`

	// Which list it belongs to
	ListId     int64 `json:"listid" apiModel:"List"`
	TemplateId int64 `json:"templateid" apiModel:"Template"`
	ContactId  int64 `json:"contactId" apiModel:"Contact"`
	ClientId   int64 `json:"clientid"`

	FromEmail string `json:"fromemail"`

	Sender      string `json:"sender"`
	To          string `json:"to"`
	Subject     string `json:"subject" datastore:",noindex"`
	BaseSubject string `json:"baseSubject" datastore:",noindex"`
	Body        string `json:"body" datastore:",noindex"`

	CC  []string `json:"cc"`  // Carbon copy email addresses
	BCC []string `json:"bcc"` // Blind carbon copy email addresses

	// User details
	FirstName string `json:"firstname"`
	LastName  string `json:"lastname"`

	SendAt time.Time `json:"sendat"`

	SendGridId  string `json:"-"`
	SparkPostId string `json:"-"`
	BatchId     string `json:"batchid"`

	GmailId       string `json:"gmailid"`
	GmailThreadId string `json:"gmailthreadid"`

	TeamId int64 `json:"teamid"`

	Attachments []int64 `json:"attachments" datastore:",noindex" apiModel:"File"`

	Delievered    bool   `json:"delivered"` // The email has been officially sent by our platform
	BouncedReason string `json:"bouncedreason"`
	Bounced       bool   `json:"bounced"`
	Clicked       int    `json:"clicked"`
	Opened        int    `json:"opened"`
	Spam          bool   `json:"spam"`
	Cancel        bool   `json:"cancel"`
	Dropped       bool   `json:"dropped"`

	SendGridOpened  int `json:"sendgridopened"`
	SendGridClicked int `json:"sendgridclicked"`

	Archived bool `json:"archived"`

	IsSent bool `json:"issent"` // Basically if the user has clicked on "/send"
}

/*
* Public methods
 */

/*
* Create methods
 */

func (e *Email) Create(r *http.Request, currentUser apiModels.UserPostgres) (*Email, error) {
	e.IsSent = false
	e.CreatedBy = currentUser.Id
	e.Created = time.Now()
	_, err := db.DB.Model(e).Returning("*").Insert()
	return e, err
}

func (es *EmailSetting) Create(r *http.Request, currentUser apiModels.UserPostgres) (*EmailSetting, error) {
	es.CreatedBy = currentUser.Id
	es.Created = time.Now()
	_, err := db.DB.Model(es).Returning("*").Insert()
	return es, err
}

/*
* Update methods
 */

// Function to save a new email into App Engine
func (e *Email) Save() (*Email, error) {
	// Update the Updated time
	e.Updated = time.Now()
	_, err := db.DB.Model(e).Update()
	return e, err
}

// Function to save a new email into App Engine
func (es *EmailSetting) Save() (*EmailSetting, error) {
	// Update the Updated time
	es.Updated = time.Now()
	_, err := db.DB.Model(es).Update()
	return es, err
}

func (e *Email) MarkSent(emailId string) (*Email, error) {
	e.IsSent = true
	e.SendGridId = emailId
	_, err := e.Save()
	if err != nil {
		log.Printf("%v", err)
		return e, err
	}
	return e, nil
}

func (e *Email) MarkBounced(reason string) (*Email, error) {
	e.Bounced = true
	e.Delievered = true
	e.BouncedReason = reason
	_, err := e.Save()
	if err != nil {
		log.Printf("%v", err)
		return e, err
	}
	return e, nil
}

func (e *Email) MarkClicked() (*Email, error) {
	if e.SendAt.IsZero() || e.SendAt.Before(time.Now()) {
		e.Clicked += 1
		e.Delievered = true
		_, err := e.Save()
		if err != nil {
			log.Printf("%v", err)
			return e, err
		}
	}
	return e, nil
}

func (e *Email) MarkDelivered() (*Email, error) {
	e.Delievered = true
	_, err := e.Save()
	if err != nil {
		log.Printf("%v", err)
		return e, err
	}
	return e, nil
}

func (e *Email) MarkSpam() (*Email, error) {
	e.Spam = true
	e.Delievered = true
	_, err := e.Save()
	if err != nil {
		log.Printf("%v", err)
		return e, err
	}
	return e, nil
}

func (e *Email) MarkOpened() (*Email, error) {
	// If already sent (sendAt is 0 or before current time)
	if e.SendAt.IsZero() || e.SendAt.Before(time.Now()) {
		e.Opened += 1
		e.Delievered = true
		_, err := e.Save()
		if err != nil {
			log.Printf("%v", err)
			return e, err
		}
	}
	return e, nil
}

func (e *Email) MarkSendgridOpened() (*Email, error) {
	e.SendGridOpened += 1
	e.Delievered = true
	_, err := e.Save()
	if err != nil {
		log.Printf("%v", err)
		return e, err
	}
	return e, nil
}

func (e *Email) MarkSendgridDropped() (*Email, error) {
	e.Dropped = true
	e.Delievered = true
	_, err := e.Save()
	if err != nil {
		log.Printf("%v", err)
		return e, err
	}
	return e, nil
}

func (e *Email) FillStruct(m map[string]interface{}) error {
	for k, v := range m {
		err := apiModels.SetField(e, k, v)
		if err != nil {
			return err
		}
	}
	return nil
}
