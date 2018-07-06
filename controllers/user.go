package controllers

import (
	"errors"
	"log"
	"net/http"

	gcontext "github.com/gorilla/context"

	"github.com/news-ai/api-v1/controllers"
	"github.com/news-ai/api-v1/models"
	// "github.com/news-ai/tabulae-v1/sync"
)

func RegisterUser(r *http.Request, user models.User) (models.UserPostgres, bool, error) {
	existingUser, err := controllers.GetUserByEmail(user.Email)

	if err != nil {
		// Validation if the email is null
		if user.Email == "" {
			noEmailErr := errors.New("User does have an email")
			log.Printf("%v", noEmailErr)
			log.Printf("%v", user)
			return models.UserPostgres{}, false, noEmailErr
		}

		// Add the user to datastore
		userPostgres := models.UserPostgres{}
		userPostgres.Data = user
		_, err = userPostgres.Create()
		if err != nil {
			log.Printf("%v", err)
			return models.UserPostgres{}, false, err
		}

		// sync.ResourceSync(r, user.Id, "User", "create")

		// Set the user
		gcontext.Set(r, "user", userPostgres)
		controllers.Update(r, &userPostgres)

		// Create a sample media list for the user
		// _, _, err = CreateSampleMediaList(c, r, user)
		// if err != nil {
		// 	log.Printf("%v", err)
		// }
		return userPostgres, true, nil
	}

	if user.RefreshToken != "" {
		existingUser.Data.RefreshToken = user.RefreshToken
	}

	if !existingUser.Data.Gmail {
		existingUser.Data.TokenType = user.TokenType
		existingUser.Data.GoogleExpiresIn = user.GoogleExpiresIn
		existingUser.Data.Gmail = user.Gmail
		existingUser.Data.GoogleId = user.GoogleId
		existingUser.Data.AccessToken = user.AccessToken
		existingUser.Data.GoogleCode = user.GoogleCode
		existingUser.Save()
	}

	return existingUser, false, errors.New("User with the email already exists")
}
