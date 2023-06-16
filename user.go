package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"gorm.io/gorm"
)

var DB *gorm.DB
var err error

var jwtKey = []byte(goDotEnvVariable("JWTKEY"))

var DNS = goDotEnvVariable("DNS")

type User struct {
	gorm.Model
	FirstName string `json:"firstname"`
	LastName  string `json:"lastname"`
	Email     string `json:"email"`
	Password  string `json:"password"`
	Wallet    Wallet
}
type LoginCred struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}
type Claims struct {
	Email  string `json:"email"`
	UserId int    `json:"userId"`
	jwt.RegisteredClaims
}

func SignUp(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var user User
	json.NewDecoder(r.Body).Decode(&user)
	user.Password, err = HashPassword(user.Password)
	if err != nil {
		fmt.Println(http.StatusBadRequest)
		return
	}
	DB.Create(&user)
	json.NewEncoder(w).Encode(&user)
	CreateWallet(w, r, int(user.ID))
}

func Login(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var logCred LoginCred
	err := json.NewDecoder(r.Body).Decode(&logCred)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	var user User
	DB.Where("email = ?", logCred.Email).First(&user)

	if user.Email != logCred.Email {
		w.WriteHeader(http.StatusUnauthorized)
		fmt.Fprint(w, "Incorrect Credential")
		return
	}

	match := CheckPasswordHash(logCred.Password, user.Password)
	if !match {
		w.WriteHeader(http.StatusUnauthorized)
		fmt.Fprint(w, "Incorrect Password")
		return
	}

	expirationTime := time.Now().Add(10 * time.Minute)

	claims := &Claims{
		Email:  logCred.Email,
		UserId: int(user.ID),
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	tokenString, err := token.SignedString(jwtKey)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:    "token",
		Value:   tokenString,
		Expires: expirationTime,
	})

	json.NewEncoder(w).Encode(tokenString)
	fmt.Fprint(w, "Login Successful")

}

func Logout(w http.ResponseWriter, r *http.Request) {
	// immediately clear the token cookie
	http.SetCookie(w, &http.Cookie{
		Name:    "token",
		Expires: time.Now(),
	})
}

func GetMyDetails(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	ctx := r.Context()
	cl := ctx.Value(&Cl{})
	var user User
	DB.Where("id = ?", cl.(*Claims).UserId).First(&user)
	DB.Preload("Wallet", "user_id = ?", cl.(*Claims).UserId).Find(&user)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
	}
	if user.ID != 0 {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(&user)
	} else {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("No user found"))
	}
}
