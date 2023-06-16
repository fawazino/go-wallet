package main

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Wallet struct {
	gorm.Model
	AccNumber   int `json:"accNumber"`
	Balance     int `json:"balance"`
	UserID      int `json:"userID"`
	Transaction []Transaction
}

type TransferCredentials struct {
	Amount        int    `json:"amount"`
	Description   string `json:"description"`
	RecipientAcct int    `json:"recipientAcct"`
}

func GetMyWallet(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	ctx := r.Context()
	cl := ctx.Value(&Cl{})

	var wallet Wallet
	DB.Where(&Wallet{UserID: cl.(*Claims).UserId}).First(&wallet)

	if err != nil {
		w.WriteHeader(http.StatusNotFound)
	}

	if wallet.UserID == cl.(*Claims).UserId {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(&wallet)
	} else {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("No wallet found"))
	}
}

func CreateWallet(w http.ResponseWriter, r *http.Request, userId int) {
	w.Header().Set("Content-Type", "application/json")
	var wallet Wallet
	wal := DB.Where(&Wallet{UserID: userId}).First(&wallet)

	if wal.Error == nil {
		w.WriteHeader(http.StatusMethodNotAllowed)
		w.Write([]byte("Wallet already created for this user"))
		json.NewEncoder(w).Encode(wal.RowsAffected)
	}
	DB.Create(&Wallet{
		AccNumber: 10000000000 + rand.Intn(9999999999),
		UserID:    userId,
	})
	json.NewEncoder(w).Encode(wal.RowsAffected)

}

func Transfer(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	ctx := r.Context()
	cl := ctx.Value(&Cl{})

	reference := uuid.New().String()
	var wallet Wallet
	var recipientWallet Wallet
	transferCred := new(TransferCredentials)
	DB.Where(&Wallet{UserID: cl.(*Claims).UserId}).First(&wallet)

	if err != nil {
		w.WriteHeader(http.StatusNotFound)
	}

	if wallet.UserID == cl.(*Claims).UserId {

		json.NewDecoder(r.Body).Decode(transferCred)

		recipient := DB.Where(&Wallet{AccNumber: int(transferCred.RecipientAcct)}).First(&recipientWallet)
		fmt.Print(transferCred)
		if recipient.Error != nil {
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte("Recipient account not found"))
		}
		if wallet.Balance < transferCred.Amount {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("insufficient funds"))
		} else {
			DB.Create(&Transaction{
				Transaction_Type:   "Debit",
				Transaction_status: "success",
				Amount:             uint(transferCred.Amount),
				Description:        fmt.Sprintf(" transfer for '%s' to %d", transferCred.Description, recipientWallet.AccNumber),
				Reference:          reference,
				WalletID:           wallet.ID,
				UserID:             wallet.UserID,
			})

			DB.Create(&Transaction{
				Transaction_Type:   "Credit",
				Transaction_status: "success",
				Amount:             uint(transferCred.Amount),
				Description:        fmt.Sprintf(" transfer for '%s' from %d", transferCred.Description, wallet.AccNumber),
				Reference:          reference,
				WalletID:           recipientWallet.ID,
				UserID:             recipientWallet.UserID,
			})
			DB.Where(&Wallet{UserID: cl.(*Claims).UserId}).Updates(Wallet{Balance: (wallet.Balance - transferCred.Amount)})
			DB.Where(&Wallet{AccNumber: int(transferCred.RecipientAcct)}).Updates(Wallet{Balance: (recipientWallet.Balance + transferCred.Amount)})
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("Transfer Successful"))
		}

	}
}
