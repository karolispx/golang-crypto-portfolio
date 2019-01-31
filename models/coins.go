package models

import (
	"database/sql"
	"encoding/json"
	"io/ioutil"
	"math"
	"math/rand"
	"net/http"
	"time"

	"github.com/karolispx/golang-crypto-portfolio/helpers"
	_ "github.com/lib/pq"
)

// CoinInfoFromCMC - coin info from coinmarketcap
type CoinInfoFromCMC struct {
	Data []struct {
		ID     int    `json:"id"`
		Name   string `json:"name"`
		Symbol string `json:"symbol"`
		Quote  struct {
			EUR struct {
				Price float64 `json:"price"`
			} `json:"EUR"`
		} `json:"quote"`
	} `json:"data"`
}

// Coin struct - store coin's info
type Coin struct {
	UserCoinID  int     `json:"coinid"`
	Name        string  `json:"name"`
	Symbol      string  `json:"symbol"`
	Invested    float64 `json:"invested"`
	Amount      float64 `json:"amount"`
	MadeLost    float64 `json:"madelost"`
	Worth       float64 `json:"worth"`
	PriceEur    float64 `json:"priceeur"`
	Lives       string  `json:"lives"`
	DateAdded   string  `json:"date_added"`
	DateUpdated string  `json:"date_updated"`
}

// SyncInfo - store info about the sync
type SyncInfo struct {
	Profit   float64 `json:"profit"`
	Invested float64 `json:"invested"`
	Worth    float64 `json:"worth"`
	LastSync string  `json:"last_sync"`
}

// CoinSymbolInfo - coin symbol info
type CoinSymbolInfo struct {
	Name   string `json:"name"`
	Symbol string `json:"symbol"`
}

// UpdateCoinsFromCMC - get all coins from CMC and update them in DB
func UpdateCoinsFromCMC(minutes int, DB *sql.DB) {
	d := time.Duration(minutes) * time.Minute
	rand.Seed(time.Now().Unix())

	for range time.Tick(d) {
		request, err := http.NewRequest("GET", "https://pro-api.coinmarketcap.com/v1/cryptocurrency/listings/latest?limit=5000&convert=EUR", nil)

		if err != nil {
			panic(err)
		}

		Config := helpers.GetConfig()

		randomNumberForCMCAPIKey := rand.Intn(9-0) + 0

		request.Header.Set("X-Cmc_pro_api_key", Config.CMCAPIKeys[randomNumberForCMCAPIKey].APIKey)

		response, err := http.DefaultClient.Do(request)

		if err != nil {
			panic(err)
		}

		defer response.Body.Close()

		items := CoinInfoFromCMC{}

		body, err := ioutil.ReadAll(response.Body)

		if err != nil {
			panic(err)
		}

		err = json.Unmarshal(body, &items)

		if err != nil {
			panic(err)
		}

		for _, item := range items.Data {
			count := 0

			row := DB.QueryRow("SELECT COUNT(*) FROM coins where cmcid = $1", item.ID)

			err := row.Scan(&count)

			if err != nil {
				panic(err)
			}

			if count < 1 {
				// Insert new coin into database
				lastInsertID := 0
				err = DB.QueryRow("INSERT INTO coins(cmcid, name, symbol, priceeur) VALUES($1, $2, $3, $4) returning coinid;", item.ID, item.Name, item.Symbol, item.Quote.EUR.Price).Scan(&lastInsertID)

				if err != nil {
					panic(err)
				}
			} else {
				// Update existing coin in the database
				lastUpdatedID := 0

				err = DB.QueryRow("UPDATE coins SET name = $1, symbol = $2, priceeur = $3 where cmcid = $4 returning coinid;",
					item.Name, item.Symbol, item.Quote.EUR.Price, item.ID).Scan(&lastUpdatedID)

				if err != nil {
					panic(err)
				}
			}
		}
	}
}

// UpdateUsersCoins - update coins for all users
func UpdateUsersCoins(minutes int, DB *sql.DB) {
	d := time.Duration(minutes) * time.Minute
	rand.Seed(time.Now().Unix())

	for range time.Tick(d) {

		// Get all coins for this user
		rows, err := DB.Query("SELECT * FROM users")

		if err != nil {
			panic(err)
		}

		// Foreach coin
		for rows.Next() {
			var UserID int
			var Email string
			var Password string
			var DateRegister string

			err = rows.Scan(&UserID, &Email, &Password, &DateRegister)

			if err != nil {
				panic(err)
			}

			_ = UpdateUserCoins(UserID, DB)
		}
	}
}

// UpdateUserCoins - update user coins
func UpdateUserCoins(userID int, DB *sql.DB) []Coin {
	// Get all coins for this user
	rows, err := DB.Query("SELECT * FROM usercoins WHERE userid = $1", userID)

	if err != nil {
		panic(err)
	}

	var coins []Coin

	// Foreach coin
	for rows.Next() {
		var UserCoinID int
		var UserID int
		var Name string
		var Symbol string
		var Invested float64
		var Amount float64
		var MadeLost float64
		var Worth float64
		var PriceEur float64
		var Lives string
		var DateAdded string
		var DateUpdated string

		err = rows.Scan(&UserCoinID, &UserID, &Name, &Symbol, &Invested, &Amount, &MadeLost, &Worth, &PriceEur, &Lives, &DateAdded, &DateUpdated)

		if err != nil {
			panic(err)
		}

		Invested = math.Round(Invested*100) / 100
		Amount = math.Round(Amount*100) / 100

		// Get coin info for this coin name and symbol
		rows, err := DB.Query("SELECT * FROM coins WHERE name = $1 AND symbol = $2", Name, Symbol)

		if err != nil {
			panic(err)
		}

		// Foreach coin
		for rows.Next() {
			var CoinID int
			var CMCCoinID int
			var CoinName string
			var CoinSymbol string
			var CoinPriceEur float64

			err = rows.Scan(&CoinID, &CMCCoinID, &CoinName, &CoinSymbol, &CoinPriceEur)

			if err != nil {
				panic(err)
			}

			calculatedPrice := CoinPriceEur * Amount

			CoinWorth := calculatedPrice
			CoinMadeLost := calculatedPrice - Invested

			CoinWorth = math.Round(CoinWorth*100) / 100
			CoinMadeLost = math.Round(CoinMadeLost*100) / 100

			// Update user coin with coin info
			var lastUpdatedID int

			err = DB.QueryRow("UPDATE usercoins SET date_updated = $1, Worth = $2, MadeLost = $3, PriceEur = $4 where userid = $5 AND name = $6 AND symbol = $7 returning usercoinid;",
				helpers.GetCurrentDateTime(), CoinWorth, CoinMadeLost, CoinPriceEur, userID, CoinName, CoinSymbol).Scan(&lastUpdatedID)

			if err != nil {
				panic(err)
			}

			coins = append(coins, Coin{
				UserCoinID:  UserCoinID,
				Name:        Name,
				Symbol:      Symbol,
				Invested:    Invested,
				Amount:      Amount,
				MadeLost:    CoinMadeLost,
				Worth:       CoinWorth,
				PriceEur:    PriceEur,
				Lives:       Lives,
				DateAdded:   DateAdded,
				DateUpdated: DateUpdated,
			})
		}
	}

	return coins
}

// GetUserCoins - get all user coins
func GetUserCoins(userID int, DB *sql.DB) ([]Coin, SyncInfo) {
	// Update user coins
	coins := UpdateUserCoins(userID, DB)

	var syncInfo SyncInfo

	syncInfo.LastSync = helpers.GetCurrentDateTime()
	syncInfo.Invested = 0
	syncInfo.Worth = 0
	syncInfo.Profit = 0

	for _, coin := range coins {
		syncInfo.Invested = syncInfo.Invested + coin.Invested
		syncInfo.Worth = syncInfo.Worth + coin.Worth
	}

	syncInfo.Profit = syncInfo.Worth - syncInfo.Invested

	syncInfo.Invested = math.Round(syncInfo.Invested*100) / 100
	syncInfo.Worth = math.Round(syncInfo.Worth*100) / 100
	syncInfo.Profit = math.Round(syncInfo.Profit*100) / 100

	return coins, syncInfo
}

// GetCoins - get all coins
func GetCoins(DB *sql.DB) []CoinSymbolInfo {

	// Get all coins
	rows, err := DB.Query("SELECT * FROM coins")

	if err != nil {
		panic(err)
	}

	var coins []CoinSymbolInfo

	// Foreach coin
	for rows.Next() {
		var CoinID int
		var CMCCoinID int
		var CoinName string
		var CoinSymbol string
		var CoinPriceEur float64

		err = rows.Scan(&CoinID, &CMCCoinID, &CoinName, &CoinSymbol, &CoinPriceEur)

		if err != nil {
			panic(err)
		}

		coins = append(coins, CoinSymbolInfo{
			Name:   CoinName,
			Symbol: CoinSymbol,
		})
	}

	return coins
}

// GetCoinBySymbol -
func GetCoinBySymbol(DB *sql.DB, coinSymbol string) (coinNameReturn string, coinPriceEURReturn float64) {
	var AddCoinName string
	var AddCoinPriceEUR float64

	// Get all coins with this symbol
	rows, err := DB.Query("SELECT name, priceeur FROM coins WHERE symbol = $1", coinSymbol)

	if err != nil {
		panic(err)
	}

	// Foreach coin
	for rows.Next() {
		var CoinName string
		var CoinPriceEUR float64

		err = rows.Scan(&CoinName, &CoinPriceEUR)

		if err != nil {
			panic(err)
		}

		AddCoinName = CoinName
		AddCoinPriceEUR = CoinPriceEUR
	}

	return AddCoinName, AddCoinPriceEUR
}

// CheckCoinBelongsToUser -
func CheckCoinBelongsToUser(DB *sql.DB, coinid string, userID int) int {
	count := 0

	row := DB.QueryRow("SELECT COUNT(*) FROM usercoins where usercoinid = $1 AND userid = $2", coinid, userID)

	err := row.Scan(&count)

	if err != nil {
		panic(err)
	}

	return count
}

// CheckUserHasCoin -
func CheckUserHasCoin(DB *sql.DB, userID int, AddCoinName string, coinSymbol string) int {
	// Check to see if user has this coin in their portfolio already
	count := 0

	row := DB.QueryRow("SELECT COUNT(*) FROM usercoins where userid = $1 AND name = $2 AND symbol = $3", userID, AddCoinName, coinSymbol)

	err := row.Scan(&count)

	if err != nil {
		panic(err)
	}

	return count
}

// CreateCoin -
func CreateCoin(DB *sql.DB, userID int, AddCoinName string, coinSymbol string, convertCoinInvested float64, convertCoinAmount float64, AddCoinPriceEUR float64, coinLives string) int {
	lastInsertID := 0

	err := DB.QueryRow("INSERT INTO usercoins(userid, name, symbol, invested, amount, madelost, worth, priceeur, lives, date_added, date_updated) VALUES($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11) returning usercoinid;",
		userID, AddCoinName, coinSymbol, convertCoinInvested, convertCoinAmount, 0, 0, AddCoinPriceEUR, coinLives, helpers.GetCurrentDateTime(), helpers.GetCurrentDateTime()).Scan(&lastInsertID)

	if err != nil {
		panic(err)
	}

	return lastInsertID
}

// UpdateCoin -
func UpdateCoin(DB *sql.DB, convertCoinInvested float64, convertCoinAmount float64, coinLives string, coinid string, userID int) int {
	lastUpdatedID := 0

	err := DB.QueryRow("UPDATE usercoins SET invested = $1, amount = $2, lives = $3, date_updated = $4 WHERE usercoinid = $5 AND userid = $6 returning usercoinid;",
		convertCoinInvested, convertCoinAmount, coinLives, helpers.GetCurrentDateTime(), coinid, userID).Scan(&lastUpdatedID)

	if err != nil {
		panic(err)
	}

	return lastUpdatedID
}

// RemoveCoin -
func RemoveCoin(DB *sql.DB, coinid string) bool {
	_, err := DB.Exec("DELETE FROM usercoins where usercoinid = $1", coinid)

	if err != nil {
		return false
	}

	return true
}
