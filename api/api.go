package api

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/asaskevich/govalidator"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/google/uuid"
	"github.com/pumpkin-a/wallet/models"
	"github.com/pumpkin-a/wallet/storage"
)

type jsonResponse struct {
	Text string `json:"server response: "`
}

type WalletStorage interface {
	InsertNewWallet(w models.Wallet) error
	DbGetWallet(walletId string) (*models.Wallet, error)
	WalletBalance(walletId string) (float64, error)
	DbDoTransfer(senderWallet, recipientWallet *models.Wallet, amount float64) error
	DbCreateHistoryRecord(h *models.HistoryRecord) error
	DbGetHistory(walletId string) (*[]models.HistoryRecord, error)
}

type Server struct {
	db WalletStorage
}

func NewServer(db WalletStorage) *Server {
	return &Server{db: db}
}

func serverResponse(w http.ResponseWriter, statusCode int, stringResponse string) {
	sr := &jsonResponse{stringResponse}
	jr, _ := json.Marshal(sr)
	w.WriteHeader(statusCode)
	w.Write(jr)
}

func RunServer(serverConfig models.Config) error {
	sqlDataBase, err := storage.DbConnection()
	if err != nil {
		log.Println("db connection error", err)
	}
	defer sqlDataBase.DB.Close()
	server := NewServer(sqlDataBase)
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Timeout(60 * time.Second))

	r.Route("/api/v1/wallet", func(r chi.Router) {
		r.Post("/", server.createWallet)
		// Subrouters
		r.Route("/{walletId}", func(r chi.Router) {
			r.Use(server.WalletCtx)
			r.Get("/", server.GetWallet)
			r.Get("/history", server.GetHistory)
			r.Route("/send", func(r chi.Router) {
				r.Post("/", server.DoTransfer)
			})
		})
	})

	log.Printf("The server was successfully started on port %s", serverConfig.ServerPort)
	err = http.ListenAndServe(serverConfig.ServerPort, r)
	if err != nil {
		log.Fatalln("error while starting server:", err)
	}
	return err
}

func (s *Server) WalletCtx(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		walletId := chi.URLParam(r, "walletId")
		ctx := context.WithValue(r.Context(), "WalletId", walletId)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (s *Server) createWallet(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	newWallet := models.Wallet{
		ID:      uuid.NewString(),
		Balance: 100.0,
	}

	err := s.db.InsertNewWallet(newWallet)
	if err != nil {
		log.Println("db insert error", err)
		serverResponse(w, http.StatusBadRequest, "Bad Request")
		return

	}
	b, err := json.Marshal(newWallet)
	if err != nil {
		log.Println("error with marshalling json:", err)
		return
	}
	w.Write(b)
}

func (s *Server) GetWallet(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	walletId := ctx.Value("WalletId").(string)
	if !govalidator.IsUUID(walletId) {
		log.Println("invalid walletId type")
		serverResponse(w, http.StatusNotFound, "the wallet was not found")
		return
	}
	wallet, err := s.db.DbGetWallet(walletId)
	if err != nil {
		serverResponse(w, http.StatusNotFound, "the wallet was not found")
		return
	}

	ok, _ := govalidator.ValidateStruct(wallet)
	if !ok || wallet.Balance < 0 {
		log.Println("invalid wallet type")
		serverResponse(w, http.StatusNotFound, "the wallet was not found")
		return
	}
	b, err := json.Marshal(wallet)
	if err != nil {
		log.Println("error with marshalling json in GetWallet:", err)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write(b)
	// w.Write([]byte(fmt.Sprintf("wallet id: %s, wallet balance: %f", wallet.ID, wallet.Balance)))
}

func (s *Server) DoTransfer(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	senderId := ctx.Value("WalletId").(string)

	w.Header().Set("content-type", "application/json")
	var transferParam models.Transfer
	err := json.NewDecoder(r.Body).Decode(&transferParam)
	if err != nil {
		log.Println("error decoding the JSON object in the request body in DoTransfer", err)
		serverResponse(w, http.StatusBadRequest, "Bad Request")
		return
	}

	ok, _ := govalidator.ValidateStruct(transferParam)
	if !ok || transferParam.Amount < 0 {
		log.Println("invalid transfer parametrs type")
		serverResponse(w, http.StatusBadRequest, "Bad Request")
		return
	}

	senderBalance, err := s.db.WalletBalance(senderId)
	if err != nil {
		serverResponse(w, http.StatusNotFound, "The sender's wallet was not found")
		return

	}

	recipientBalance, err := s.db.WalletBalance(transferParam.To)
	if err != nil {
		serverResponse(w, http.StatusNotFound, "The recipient's wallet was not found")
		return

	}

	sender := &models.Wallet{
		ID:      senderId,
		Balance: senderBalance,
	}
	recipient := &models.Wallet{
		ID:      transferParam.To,
		Balance: recipientBalance,
	}

	if sender.Balance < transferParam.Amount {
		serverResponse(w, http.StatusBadRequest, "Insufficient funds")
		return
	}

	err = s.db.DbDoTransfer(sender, recipient, transferParam.Amount)
	if err != nil {
		serverResponse(w, http.StatusBadRequest, "Transfer error")
		return
	}
	h := HistoryRecord(senderId, &transferParam)
	s.db.DbCreateHistoryRecord(h)
	serverResponse(w, http.StatusOK, "The transfer was successfully completed")

}

func HistoryRecord(senderId string, trPar *models.Transfer) *models.HistoryRecord {
	formattedTime := time.Now().Format(time.RFC3339)
	historyRecord := &models.HistoryRecord{
		Id:     uuid.NewString(),
		Time:   formattedTime,
		From:   senderId,
		To:     trPar.To,
		Amount: trPar.Amount,
	}
	return historyRecord
}

func (s *Server) GetHistory(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	walletId := ctx.Value("WalletId").(string)
	history, err := s.db.DbGetHistory(walletId)
	if err != nil {
		serverResponse(w, http.StatusNotFound, "the wallet was not found")
		return
	}
	// log.Println(history)
	b, err := json.Marshal(history)
	if err != nil {
		log.Println("error with marshalling json in GetWallet:", err)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write(b)

}
