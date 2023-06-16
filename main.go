package main

import (
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

func initializeRouter() {
	r := mux.NewRouter()

	r.HandleFunc("/signup", SignUp)
	r.HandleFunc("/login", Login).Methods("POST")
	r.HandleFunc("/logout", verifyJWT(Logout))

	r.HandleFunc("/user", verifyJWT(GetMyDetails))
	r.HandleFunc("/wallet", verifyJWT(GetMyWallet))
	r.HandleFunc("/transfer", verifyJWT(Transfer))

	r.HandleFunc("/transactions", verifyJWT(GetTransactions))
	r.HandleFunc("/transactions/{id}", verifyJWT(GetTransactionByID))

	log.Fatal(http.ListenAndServe(":9001", r))
}

func main() {
	initialMigration()
	initializeRouter()
}
