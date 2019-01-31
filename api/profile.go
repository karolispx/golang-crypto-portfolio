package api

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/karolispx/golang-crypto-portfolio/helpers"
	"github.com/karolispx/golang-crypto-portfolio/models"
)

// GetProfile - get user profile
func GetProfile(w http.ResponseWriter, r *http.Request) {
	userID := helpers.ValidateJWT(w, r)

	if userID > 0 {
		DB := helpers.InitDB()

		getUserProfile := models.GetUserProfile(DB, userID)

		defer DB.Close()

		helpers.Respond(w, r, getUserProfile, "success", 200)

		return
	}
}

// UpdateProfitsEndpoint -
func UpdateProfitsEndpoint(w http.ResponseWriter, r *http.Request) {
	userID := helpers.ValidateJWT(w, r)

	if userID > 0 {
		vars := mux.Vars(r)
		action := vars["action"]

		if action == "" {
			response := "Such action does not exists.."

			helpers.Respond(w, r, response, "error", 422)

			return
		}

		if action != "enable" && action != "disable" && action != "refresh" {
			response := "Such action does not exists.."

			helpers.Respond(w, r, response, "error", 422)

			return
		}

		DB := helpers.InitDB()

		profitsEndpointAction := models.ProfitsEndpointAction(DB, userID, action)

		if profitsEndpointAction == false {
			helpers.DefaultErrorRespond(w, r)

			return
		}

		actionWord := "enabled"

		if action == "disable" {
			actionWord = "disabled"
		} else if action == "refresh" {
			actionWord = "generated"
		}

		response := "Profits endpoint has been " + actionWord + " successfully!"

		helpers.Respond(w, r, response, "success", 200)

		return
	}
}
