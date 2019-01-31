package helpers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
)

// Response message
type Response struct {
	Type     string      `json:"type"`
	Response interface{} `json:"response"`
}

// Token information
type Token struct {
	UserID int
	jwt.StandardClaims
}

// GetCurrentDateTime in string format
func GetCurrentDateTime() string {
	currentTime := time.Now()

	return currentTime.Format("2006.01.02 15:04:05")
}

// Respond - process rest api response
func Respond(w http.ResponseWriter, r *http.Request, response interface{}, responseType string, statusCode int) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")

	w.WriteHeader(statusCode)

	returnResponse := Response{
		Type:     responseType,
		Response: response,
	}

	json.NewEncoder(w).Encode(returnResponse)
}

// DefaultErrorRespond - respond with default error
func DefaultErrorRespond(w http.ResponseWriter, r *http.Request) {
	response := "An error occured. Please try again later."

	Respond(w, r, response, "error", 500)
}

// GenerateJWT - generating JWT
func GenerateJWT(userID int) string {
	Config := GetConfig()

	token := jwt.New(jwt.SigningMethodHS256)

	claims := token.Claims.(jwt.MapClaims)

	claims["UserID"] = userID
	claims["exp"] = time.Now().Add(time.Minute * 10080).Unix() // 7 days - 10080

	tokenString, err := token.SignedString([]byte(Config.JWTSigningKey))

	if err != nil {
		fmt.Println(err)
	}

	return tokenString
}

// ValidateJWT - validate JWT
func ValidateJWT(w http.ResponseWriter, r *http.Request) int {
	tokenHeader := r.Header.Get("Authorization")

	if tokenHeader == "" {
		response := "You need to be logged in to access this page."

		Respond(w, r, response, "error", 403)

		return 0
	}

	splitted := strings.Split(tokenHeader, " ")

	if len(splitted) != 2 {
		response := "You need to be logged in to access this page."

		Respond(w, r, response, "error", 403)

		return 0
	}

	Config := GetConfig()

	tokenPart := splitted[1]

	headerToken := &Token{}

	token, err := jwt.ParseWithClaims(tokenPart, headerToken, func(token *jwt.Token) (interface{}, error) {
		return []byte(Config.JWTSigningKey), nil
	})

	// Malformed token, returns with http code 403 as usual
	if err != nil {
		response := "You need to be logged in to access this page."

		Respond(w, r, response, "error", 403)

		return 0
	}

	if !token.Valid {
		response := "You need to be logged in to access this page."

		Respond(w, r, response, "error", 403)

		return 0
	}

	return headerToken.UserID
}

// ValidateEmailAddress - validate email address
func ValidateEmailAddress(email string) bool {
	Re := regexp.MustCompile(`^[a-z0-9._%+\-]+@[a-z0-9.\-]+\.[a-z]{2,4}$`)

	return Re.MatchString(email)
}
