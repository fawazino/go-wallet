package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/golang-jwt/jwt"
	"github.com/joho/godotenv"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

type Cl struct {
}

type Response struct {
	Type    string
	Message string
}

func goDotEnvVariable(key string) string {

	err := godotenv.Load(".env")

	if err != nil {
		log.Fatalf("Error loading .env file")
	}

	return os.Getenv(key)
}

func initialMigration() {
	DB, err = gorm.Open(mysql.Open(DNS), &gorm.Config{})
	if err != nil {
		fmt.Println(err.Error())
		panic("cannot connecct to db")
	}
	DB.AutoMigrate(&User{}, &Wallet{}, &Transaction{})
}

func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(bytes), err
}

func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

func verifyJWT(endpointHandler func(w http.ResponseWriter, r *http.Request)) http.HandlerFunc {

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		c, err := r.Cookie("token")
		if err != nil {
			if err == http.ErrNoCookie {
				w.WriteHeader(http.StatusUnauthorized)
				w.Write([]byte("You're Unauthorized due to no token"))

				return
			}
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("You're Unauthorized due to error parsing the JWT"))
			return
		}

		// Get the JWT string from the cookie
		tknStr := c.Value

		// Initialize a new instance of `Claims`
		claims := &Claims{}

		tkn, err := jwt.ParseWithClaims(tknStr, claims, func(token *jwt.Token) (interface{}, error) {
			return jwtKey, nil
		})
		if err != nil {
			if err == jwt.ErrSignatureInvalid {
				w.WriteHeader(http.StatusUnauthorized)
				w.Write([]byte("You're Unauthorized due to invalid token"))
				return
			}
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("You're Unauthorized "))
			return
		}
		if !tkn.Valid {
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte("You're Unauthorized due to invalid token"))
			return
		}
		ctx = context.WithValue(ctx, &Cl{}, tkn.Claims)
		endpointHandler(w, r.WithContext(ctx))

	})
}
