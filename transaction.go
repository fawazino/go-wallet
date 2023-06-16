package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"gorm.io/gorm"
)

type Transaction struct {
	gorm.Model
	Transaction_Type   string `json:"transaction_type"`
	Transaction_status string `json:"transaction_status"`
	Amount             uint   `json:"amount"`
	Description        string `json:"description"`
	Reference          string `json:"reference"`
	WalletID           uint   `json:"walletId"`
	UserID             int    `json:"userId"`
}

func GetTransactions(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	ctx := r.Context()
	cl := ctx.Value(&Cl{})
	var transactions []Transaction
	DB.Where(&Transaction{UserID: cl.(*Claims).UserId}).Find(&transactions)

	if err != nil {
		w.WriteHeader(http.StatusNotFound)
	}
	if len(transactions) != 0 {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(&transactions)
	} else {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("No transaction found"))
	}
}

func GetTransactionByID(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	ctx := r.Context()
	cl := ctx.Value(&Cl{})

	id, _ := strconv.Atoi(mux.Vars(r)["id"])
	var transactions Transaction
	DB.Where(&Transaction{UserID: cl.(*Claims).UserId}).Where("id = ?", id).Find(&transactions)

	if err != nil {
		w.WriteHeader(http.StatusNotFound)
	}

	if transactions.ID == uint(id) {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(&transactions)
	} else {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("No transaction found"))
	}
}

func DebitTransaction(w http.ResponseWriter, r *http.Request, amount uint, walletId uint) {
	reference := uuid.New().String()
	var wallet Wallet
	DB.Where("id = ?", walletId).First(&wallet)
	DB.Create(&Transaction{
		Transaction_Type:   "Debit",
		Transaction_status: "success",
		Amount:             amount,
		Description:        fmt.Sprintf(" transfer to %d", wallet.AccNumber),
		Reference:          reference,
		WalletID:           walletId,
		UserID:             wallet.UserID,
	})
	json.NewEncoder(w).Encode(&wallet)
}

func CreditTransaction(w http.ResponseWriter, r *http.Request, amount uint, walletId uint) {
	reference := uuid.New()
	var wallet Wallet
	DB.Where("id = ?", walletId).First(&wallet)
	DB.Where("id = ?", walletId).First(&wallet)
	DB.Create(&Transaction{
		Transaction_Type:   "Debit",
		Transaction_status: "success",
		Amount:             amount,
		Description:        fmt.Sprintf(" transfer from %d", wallet.AccNumber),
		Reference:          reference.String(),
		WalletID:           walletId,
		UserID:             wallet.UserID,
	})
	json.NewEncoder(w).Encode(&wallet)
}
