package api

import (
	"encoding/json"
	"net/http"

	_ "github.com/lib/pq"

	"github.com/karolispx/golang-crypto-portfolio/helpers"
	"github.com/karolispx/golang-crypto-portfolio/models"
)

// User information
type User struct {
	Email     string `json:"email"`
	Password  string `json:"password"`
	Password2 string `json:"password2"`
	Token     string `json:"token"`
}

// Login - process user login
func Login(w http.ResponseWriter, r *http.Request) {
	user := &User{}

	err := json.NewDecoder(r.Body).Decode(user)

	if err != nil {
		response := "Please provide all information."

		helpers.Respond(w, r, response, "error", 422)

		return
	}

	if user.Email == "" || user.Password == "" {
		response := "Please provide all information."

		helpers.Respond(w, r, response, "error", 422)

		return
	}

	DB := helpers.InitDB()

	userID := models.UserValidLogin(DB, user.Email, user.Password)

	defer DB.Close()

	if userID > 0 {
		generateJWT := helpers.GenerateJWT(userID)

		if generateJWT != "" {
			helpers.Respond(w, r, generateJWT, "success", 200)

			return
		}

		helpers.DefaultErrorRespond(w, r)

		return
	}

	response := "We could not log you in."

	helpers.Respond(w, r, response, "error", 403)

	return
}

// Register - process user registration
func Register(w http.ResponseWriter, r *http.Request) {
	user := &User{}

	err := json.NewDecoder(r.Body).Decode(user)

	if err != nil {
		response := "Please provide all information."

		helpers.Respond(w, r, response, "error", 422)

		return
	}

	if user.Email == "" || user.Password == "" || user.Password2 == "" {
		response := "Please provide all information."

		helpers.Respond(w, r, response, "error", 422)

		return
	} else if helpers.ValidateEmailAddress(user.Email) != true {
		response := "Email address is not valid!"

		helpers.Respond(w, r, response, "error", 422)

		return
	} else if user.Password != user.Password2 {
		response := "Passwords do not match!"

		helpers.Respond(w, r, response, "error", 422)

		return
	}

	DB := helpers.InitDB()

	countUsers := models.CountUsersWithEmailAddress(DB, user.Email)

	if countUsers > 0 {
		response := "This email address is taken already!"

		helpers.Respond(w, r, response, "error", 422)
		return
	}

	lastInsertID := models.CreateUser(DB, user.Email, user.Password)

	defer DB.Close()

	if lastInsertID < 1 {
		helpers.DefaultErrorRespond(w, r)

		return
	}

	generateJWT := helpers.GenerateJWT(lastInsertID)

	if generateJWT != "" {
		helpers.Respond(w, r, generateJWT, "success", 200)

		return
	}

	helpers.DefaultErrorRespond(w, r)

	return
}
