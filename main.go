package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/karolispx/golang-crypto-portfolio/api"
	"github.com/karolispx/golang-crypto-portfolio/helpers"
	"github.com/karolispx/golang-crypto-portfolio/models"
)

// Main function
func main() {
	DB := helpers.InitDB()

	defer DB.Close()

	// Update all coin info from CMC
	go models.UpdateCoinsFromCMC(10, DB)

	// Update all user's coins
	go models.UpdateUsersCoins(5, DB)

	// Initialize routes
	InitRoutes()
}

// InitRoutes initializes routes
func InitRoutes() {
	Config := helpers.GetConfig()

	router := mux.NewRouter().StrictSlash(true)

	router.HandleFunc(Config.RestAPIPath+"/auth/register", api.Register).Methods("POST")
	router.HandleFunc(Config.RestAPIPath+"/auth/login", api.Login).Methods("POST")

	router.HandleFunc(Config.RestAPIPath+"/profile", api.GetProfile).Methods("GET")

	router.HandleFunc(Config.RestAPIPath+"/endpoint/profits/{endpoint}", api.GetProfitsEndpoint).Methods("GET")
	router.HandleFunc(Config.RestAPIPath+"/endpoint/profits/{action}", api.UpdateProfitsEndpoint).Methods("PUT")

	router.HandleFunc(Config.RestAPIPath+"/portfolio/profits", api.GetProfits).Methods("GET")
	router.HandleFunc(Config.RestAPIPath+"/portfolio/symbols", api.GetSymbols).Methods("GET")

	router.HandleFunc(Config.RestAPIPath+"/portfolio/coins", api.AddCoin).Methods("POST")
	router.HandleFunc(Config.RestAPIPath+"/portfolio/coins/{coinid}", api.EditCoin).Methods("PUT")
	router.HandleFunc(Config.RestAPIPath+"/portfolio/coins/{coinid}", api.DeleteCoin).Methods("DELETE")

	fmt.Println("Server is running on: " + Config.RestAPIURL + ":" + Config.Port)

	headersOk := handlers.AllowedHeaders([]string{"X-Requested-With", "Content-Type", "Authorization"})
	originsOk := handlers.AllowedOrigins([]string{"*"})
	methodsOk := handlers.AllowedMethods([]string{"GET", "HEAD", "POST", "PUT", "OPTIONS", "DELETE"})

	log.Fatal(http.ListenAndServe(":"+Config.Port, handlers.CORS(originsOk, headersOk, methodsOk)(router)))
}
