package models

import (
	"database/sql"
	"strconv"
	"strings"
	"time"

	"github.com/karolispx/golang-crypto-portfolio/helpers"
	_ "github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"
)

// UserProfile -
type UserProfile struct {
	EmailAddress           string `json:"email_address"`
	DateRegister           string `json:"date_register"`
	ProfitsEndpointEnabled string `json:"profits_endpoint_enabled"`
	ProfitsEndpoint        string `json:"profits_endpoint"`
}

// ProfitsEndpointAction -
func ProfitsEndpointAction(DB *sql.DB, userID int, action string) bool {
	// User settings from usersettings table
	rows, err := DB.Query("SELECT * FROM usersettings where userid = $1 AND ( name = $2 OR name = $3 )", userID, "profits_endpoint_enabled", "profits_endpoint")

	if err != nil {
		panic(err)
	}

	var ProfitsEndpointEnabled string

	// Foreach usersetting
	for rows.Next() {
		var idFromDB int
		var userIDFromDB int

		var settingNameFromDB string
		var settingValueFromDB string

		err = rows.Scan(&idFromDB, &userIDFromDB, &settingNameFromDB, &settingValueFromDB)

		if err != nil {
			panic(err)
		}

		if settingNameFromDB == "profits_endpoint_enabled" {
			ProfitsEndpointEnabled = settingValueFromDB
		}
	}

	okToEnable := false
	okToDisable := false
	okToRefresh := false

	if ProfitsEndpointEnabled == "yes" {
		okToDisable = true
		okToRefresh = true
	} else {
		okToEnable = true
	}

	if (action == "enable" && okToEnable) || (action == "disable" && okToDisable) {
		return ProfitsEndpointUpdate(DB, userID, action)
	} else if action == "refresh" && okToRefresh {
		return ProfitsEndpointGenerate(DB, userID)
	}

	return false
}

// ProfitsEndpointUpdate -
func ProfitsEndpointUpdate(DB *sql.DB, userID int, action string) bool {
	enableProfitsEndpointSettingExists := 0

	actionWord := "yes"

	if action == "disable" {
		actionWord = "no"
	}

	// See if user with this email address already exists
	row := DB.QueryRow("SELECT COUNT(*) FROM usersettings WHERE name = $1 AND userid = $2", "profits_endpoint_enabled", userID)

	err := row.Scan(&enableProfitsEndpointSettingExists)

	if err != nil {
		panic(err)
	}

	if enableProfitsEndpointSettingExists > 0 {
		// Update setting
		lastUpdatedID := 0

		err := DB.QueryRow("UPDATE usersettings SET value = $1 WHERE name = $2 AND userid = $3 returning usersettingid;",
			actionWord, "profits_endpoint_enabled", userID).Scan(&lastUpdatedID)

		if err != nil {
			panic(err)
		}

		if actionWord == "yes" {
			profitsEndpointSettingExists := 0

			row := DB.QueryRow("SELECT COUNT(*) FROM usersettings WHERE name = $1 AND userid = $2", "profits_endpoint", userID)

			err := row.Scan(&profitsEndpointSettingExists)

			if err != nil {
				panic(err)
			}

			if profitsEndpointSettingExists < 1 {
				// Generate new prodits endpoint
				lastInsertID := 0

				err = DB.QueryRow("INSERT INTO usersettings(userid, name, value) VALUES($1, $2, $3) returning usersettingid;", userID, "profits_endpoint", GenerateEndpoint()).Scan(&lastInsertID)

				if err != nil {
					panic(err)
				}
			}
		}
	} else {
		// Create new setting
		lastInsertID := 0
		err := DB.QueryRow("INSERT INTO usersettings(userid, name, value) VALUES($1, $2, $3) returning usersettingid;", userID, "profits_endpoint_enabled", actionWord).Scan(&lastInsertID)

		if err != nil {
			panic(err)
		}

		if actionWord == "yes" {
			// Generate new proFits endpoint
			lastInsertID = 0

			err = DB.QueryRow("INSERT INTO usersettings(userid, name, value) VALUES($1, $2, $3) returning usersettingid;", userID, "profits_endpoint", GenerateEndpoint()).Scan(&lastInsertID)

			if err != nil {
				panic(err)
			}
		}
	}

	return true
}

// ProfitsEndpointGenerate -
func ProfitsEndpointGenerate(DB *sql.DB, userID int) bool {
	profitsEndpointSettingExists := 0

	row := DB.QueryRow("SELECT COUNT(*) FROM usersettings WHERE name = $1 AND userid = $2", "profits_endpoint", userID)

	err := row.Scan(&profitsEndpointSettingExists)

	if err != nil {
		panic(err)
	}

	if profitsEndpointSettingExists < 1 {
		// Generate new prodits endpoint
		lastInsertID := 0

		err = DB.QueryRow("INSERT INTO usersettings(userid, name, value) VALUES($1, $2, $3) returning usersettingid;", userID, "profits_endpoint", GenerateEndpoint()).Scan(&lastInsertID)

		if err != nil {
			panic(err)
		}
	} else {
		lastUpdatedID := 0

		err = DB.QueryRow("UPDATE usersettings SET value = $1 where userid = $2 AND name = $3 returning usersettingid;",
			GenerateEndpoint(), userID, "profits_endpoint").Scan(&lastUpdatedID)

		if err != nil {
			panic(err)
		}
	}

	return true
}

// UserProfitsEndpoint -
func UserProfitsEndpoint(DB *sql.DB, endpoint string) int {
	// User settings from usersettings table
	rows, err := DB.Query("SELECT userid FROM usersettings where value = $1 AND name = $2", endpoint, "profits_endpoint")

	if err != nil {
		panic(err)
	}

	var ProfitsEndpointUserID int

	// Foreach usersetting
	for rows.Next() {
		var userIDFromDB int

		err = rows.Scan(&userIDFromDB)

		if err != nil {
			panic(err)
		}

		ProfitsEndpointUserID = userIDFromDB
	}

	if ProfitsEndpointUserID < 1 {
		return 0
	}

	// User settings from usersettings table
	rows, err = DB.Query("SELECT value FROM usersettings where userid = $1 AND name = $2", ProfitsEndpointUserID, "profits_endpoint_enabled")

	if err != nil {
		panic(err)
	}

	var ProfitsEndpointEnabledValue string

	// Foreach usersetting
	for rows.Next() {
		var settingValueFromDB string

		err = rows.Scan(&settingValueFromDB)

		if err != nil {
			panic(err)
		}

		ProfitsEndpointEnabledValue = settingValueFromDB
	}

	if ProfitsEndpointEnabledValue != "yes" {
		return 0
	}

	return ProfitsEndpointUserID
}

// GenerateEndpoint - generate
func GenerateEndpoint() string {
	currentTimeString := strconv.FormatInt(time.Now().UTC().UnixNano(), 10)

	currentTimeString = strings.Replace(currentTimeString, "0", "a", -1)
	currentTimeString = strings.Replace(currentTimeString, "1", "jh", -1)
	currentTimeString = strings.Replace(currentTimeString, "2", "zw", -1)
	currentTimeString = strings.Replace(currentTimeString, "3", "mp", -1)
	currentTimeString = strings.Replace(currentTimeString, "4", "1z", -1)

	return currentTimeString
}

// UserValidLogin by email address and password
func UserValidLogin(DB *sql.DB, emailAddress string, password string) int {
	// See if user with this email address and password exists
	rows, err := DB.Query("SELECT * FROM users where email_address = $1", emailAddress)

	if err != nil {
		panic(err)
	}

	userExists := false

	var passwordToCheck string
	var emailToCheck string
	var userID int

	// Foreach user in db
	for rows.Next() {
		var idFromDB int
		var emailFromDB string
		var passwordFromDB string
		var DateRegister string

		err = rows.Scan(&idFromDB, &emailFromDB, &passwordFromDB, &DateRegister)

		if err != nil {
			panic(err)
		}

		userID = idFromDB
		emailToCheck = emailFromDB
		passwordToCheck = passwordFromDB
		userExists = true
	}

	userValidLogin := false

	if userExists && emailToCheck != "" && passwordToCheck != "" {
		if err = bcrypt.CompareHashAndPassword([]byte(passwordToCheck), []byte(password)); err != nil {
			userValidLogin = false
		} else {
			userValidLogin = true
		}
	}

	if userValidLogin == true {
		return userID
	}

	return 0
}

// CountUsersWithEmailAddress to see if user with this email address exists already
func CountUsersWithEmailAddress(DB *sql.DB, emailAddress string) int {
	countUsers := 0

	// See if user with this email address already exists
	row := DB.QueryRow("SELECT COUNT(*) FROM users where email_address = $1", emailAddress)

	err := row.Scan(&countUsers)

	if err != nil {
		panic(err)
	}

	return countUsers
}

// CreateUser in the DB
func CreateUser(DB *sql.DB, emailAddress string, password string) int {
	var lastInsertID int

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), 8)
	userPassword := string(hashedPassword)

	// Insert account into db
	err = DB.QueryRow("INSERT INTO users(email_address, password, date_register ) VALUES($1, $2, $3) returning userid;", emailAddress, userPassword, helpers.GetCurrentDateTime()).Scan(&lastInsertID)

	if err != nil {
		panic(err)
	}

	return lastInsertID
}

// GetUserProfile by ID
func GetUserProfile(DB *sql.DB, userID int) UserProfile {
	var userProfile UserProfile

	// User info from users table
	rows, err := DB.Query("SELECT * FROM users where userid = $1", userID)

	if err != nil {
		panic(err)
	}

	// Foreach user in db
	for rows.Next() {
		var idFromDB int
		var emailFromDB string
		var passwordFromDB string
		var userDateRegister string

		err = rows.Scan(&idFromDB, &emailFromDB, &passwordFromDB, &userDateRegister)

		if err != nil {
			panic(err)
		}

		userProfile.DateRegister = userDateRegister
		userProfile.EmailAddress = emailFromDB
	}

	// User settings from usersettings table
	rows, err = DB.Query("SELECT * FROM usersettings where userid = $1 AND ( name = $2 OR name = $3 )", userID, "profits_endpoint_enabled", "profits_endpoint")

	if err != nil {
		panic(err)
	}

	// Foreach usersetting
	for rows.Next() {
		var idFromDB int
		var userIDFromDB int

		var settingNameFromDB string
		var settingValueFromDB string

		err = rows.Scan(&idFromDB, &userIDFromDB, &settingNameFromDB, &settingValueFromDB)

		if err != nil {
			panic(err)
		}

		if settingNameFromDB == "profits_endpoint_enabled" {
			userProfile.ProfitsEndpointEnabled = settingValueFromDB
		} else if settingNameFromDB == "profits_endpoint" {
			userProfile.ProfitsEndpoint = settingValueFromDB
		}
	}

	return userProfile
}
