package api

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/karolispx/golang-crypto-portfolio/helpers"
	"github.com/karolispx/golang-crypto-portfolio/models"
)

// GetProfits - get user profits
func GetProfits(w http.ResponseWriter, r *http.Request) {
	userID := helpers.ValidateJWT(w, r)

	if userID > 0 {
		DB := helpers.InitDB()

		defer DB.Close()

		type ResponseSuccessData struct {
			CoinData []models.Coin   `json:"coindata"`
			SyncData models.SyncInfo `json:"syncdata"`
		}

		// Get coins
		processCoins, syncInfo := models.GetUserCoins(userID, DB)

		helpers.Respond(w, r, ResponseSuccessData{CoinData: processCoins, SyncData: syncInfo}, "success", 200)

		return
	}
}

// GetProfitsEndpoint - get user profits for the endpoint
func GetProfitsEndpoint(w http.ResponseWriter, r *http.Request) {

	vars := mux.Vars(r)
	endpoint := vars["endpoint"]

	if endpoint == "" {
		response := "This profit endpoint does not exist or has not been enabled."

		helpers.Respond(w, r, response, "error", 422)

		return
	}

	DB := helpers.InitDB()

	defer DB.Close()

	type ResponseSuccessData struct {
		CoinData []models.Coin   `json:"coindata"`
		SyncData models.SyncInfo `json:"syncdata"`
	}

	userID := models.UserProfitsEndpoint(DB, endpoint)

	if userID < 1 {
		response := "This profit endpoint does not exist or has not been enabled."

		helpers.Respond(w, r, response, "error", 422)

		return
	}

	// Get coins
	processCoins, syncInfo := models.GetUserCoins(userID, DB)

	helpers.Respond(w, r, ResponseSuccessData{CoinData: processCoins, SyncData: syncInfo}, "success", 200)

	return

}

// GetSymbols - get coins with symbols
func GetSymbols(w http.ResponseWriter, r *http.Request) {
	DB := helpers.InitDB()

	defer DB.Close()

	type ResponseSuccessData struct {
		Coins []models.CoinSymbolInfo `json:"coins"`
	}

	// Get coins
	coins := models.GetCoins(DB)

	helpers.Respond(w, r, ResponseSuccessData{Coins: coins}, "success", 200)
}

// AddCoin - add new coin
func AddCoin(w http.ResponseWriter, r *http.Request) {
	userID := helpers.ValidateJWT(w, r)

	if userID > 0 {

		// CoinInformation - coin information
		type CoinInformation struct {
			Symbol   string `json:"symbol"`
			Invested string `json:"invested"`
			Amount   string `json:"amount"`
			Lives    string `json:"lives"`
		}

		coin := &CoinInformation{}

		err := json.NewDecoder(r.Body).Decode(coin)

		if err != nil {
			response := "Please provide all information. 1"

			helpers.Respond(w, r, response, "error", 422)

			return
		}

		if coin.Symbol == "" || coin.Invested == "" || coin.Amount == "" || coin.Lives == "" {
			response := "Please provide all information.2"

			helpers.Respond(w, r, response, "error", 422)

			return
		}

		DB := helpers.InitDB()

		AddCoinName, AddCoinPriceEUR := models.GetCoinBySymbol(DB, coin.Symbol)

		// No coin name came from from DB - coin doesn't exist
		if AddCoinName == "" || AddCoinPriceEUR < 0 {
			response := "This coin does not exist!"

			helpers.Respond(w, r, response, "error", 422)

			return
		}

		convertCoinInvested, err := strconv.ParseFloat(coin.Invested, 64)

		if err != nil {
			panic(err)
		}

		convertCoinAmount, err := strconv.ParseFloat(coin.Amount, 64)

		if err != nil {
			panic(err)
		}

		checkUserHasCoin := models.CheckUserHasCoin(DB, userID, AddCoinName, coin.Symbol)

		if checkUserHasCoin > 0 {
			response := "This coin exists in your portfolio already!"

			helpers.Respond(w, r, response, "error", 422)

			return
		}

		createCoin := models.CreateCoin(DB, userID, AddCoinName, coin.Symbol, convertCoinInvested, convertCoinAmount, AddCoinPriceEUR, coin.Lives)

		defer DB.Close()

		if createCoin < 1 {
			helpers.DefaultErrorRespond(w, r)

			return
		}

		response := "Coin has been added successfully!"

		helpers.Respond(w, r, response, "success", 200)

		return
	}
}

// EditCoin - edit coin
func EditCoin(w http.ResponseWriter, r *http.Request) {
	userID := helpers.ValidateJWT(w, r)

	if userID > 0 {
		vars := mux.Vars(r)
		coinid := vars["coinid"]

		if coinid == "" {
			response := "Please provide all information."

			helpers.Respond(w, r, response, "error", 422)

			return
		}

		// Coin - adding coin
		type Coin struct {
			Invested string `json:"invested"`
			Amount   string `json:"amount"`
			Lives    string `json:"lives"`
		}

		coin := &Coin{}

		err := json.NewDecoder(r.Body).Decode(coin)

		if err != nil {
			response := "Please provide all information."

			helpers.Respond(w, r, response, "error", 422)

			return
		}

		if coin.Invested == "" || coin.Amount == "" || coin.Lives == "" {
			response := "Please provide all information."

			helpers.Respond(w, r, response, "error", 422)

			return
		}

		DB := helpers.InitDB()

		// Check if this coin belongs to the user
		checkCoinBelongsToUser := models.CheckCoinBelongsToUser(DB, coinid, userID)

		if checkCoinBelongsToUser < 1 {
			response := "This coin does not belong to you!"

			helpers.Respond(w, r, response, "error", 422)

			return
		}

		convertCoinInvested, err := strconv.ParseFloat(coin.Invested, 64)

		if err != nil {
			panic(err)
		}

		convertCoinAmount, err := strconv.ParseFloat(coin.Amount, 64)

		if err != nil {
			panic(err)
		}

		updateCoin := models.UpdateCoin(DB, convertCoinInvested, convertCoinAmount, coin.Lives, coinid, userID)

		defer DB.Close()

		if updateCoin < 1 {
			helpers.DefaultErrorRespond(w, r)

			return
		}

		response := "Coin has been updated successfully!"

		helpers.Respond(w, r, response, "success", 200)

		return
	}
}

// DeleteCoin - delete coin
func DeleteCoin(w http.ResponseWriter, r *http.Request) {
	userID := helpers.ValidateJWT(w, r)

	if userID > 0 {
		vars := mux.Vars(r)
		coinid := vars["coinid"]

		if coinid == "" {
			response := "Please provide all information."

			helpers.Respond(w, r, response, "error", 422)

			return
		}

		DB := helpers.InitDB()

		// Check if this coin belongs to the user
		checkCoinBelongsToUser := models.CheckCoinBelongsToUser(DB, coinid, userID)

		if checkCoinBelongsToUser < 1 {
			response := "This coin does not belong to you!"

			helpers.Respond(w, r, response, "error", 422)

			return
		}

		removeCoin := models.RemoveCoin(DB, coinid)

		defer DB.Close()

		if removeCoin == false {
			helpers.DefaultErrorRespond(w, r)

			return
		}

		response := "Coin has been deleted successfully!"

		helpers.Respond(w, r, response, "success", 200)

		return
	}
}
