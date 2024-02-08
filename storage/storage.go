package storage

import (
	"database/sql"
	"errors"
	"log"

	"github.com/asaskevich/govalidator"
	"github.com/pumpkin-a/wallet/models"
)

// func NewConnection() (*SqlDataBase, error) {
// 	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
// 		host, port, user, password, dbname)
// 	DB, err := sql.Open("postgres", psqlInfo)
// 	// defer sqlDataBase.DB.Close()
// 	if err != nil {
// 		log.Println("DB open error", err)
// 		return nil, err
// 	}
// 	sqlDataBase := SqlDataBase{DB: DB}
// 	return &sqlDataBase, nil
// }

func (DB *SqlDataBase) InsertNewWallet(w models.Wallet) error {
	_, err := DB.DB.Exec("insert into wallet_main  (wallet_id, balance) values ($1, $2)",
		w.ID, w.Balance)
	if err != nil {
		log.Println("error adding a new wallet to the database", err)
		return err
	}
	log.Println("the entry of new wallet has been successfully added to the database")
	return nil
}

func (DB *SqlDataBase) WalletBalance(walletId string) (float64, error) {
	r := DB.DB.QueryRow("select balance from wallet_main where wallet_id = $1",
		walletId)
	var b float64
	err := r.Scan(&b)
	log.Println("balance is", b)
	if err != nil {
		if err == sql.ErrNoRows {
			log.Printf("the entry with id %s does not exist", walletId)
			return 0, errors.New("the wallet was not found")
		}
		log.Println("error in SendersBalance when performing scan:", err)
		return 0, err
	}
	return b, nil
}

func (DB *SqlDataBase) DbDoTransfer(senderWallet, recipientWallet *models.Wallet, amount float64) error {
	_, err := DB.DB.Exec(`UPDATE public.wallet_main 
	SET balance = CASE 
		WHEN wallet_id = $1 THEN $2
		WHEN wallet_id = $3 THEN $4
		ELSE balance
	END
	WHERE wallet_id IN ($1, $3);`,
		senderWallet.ID, senderWallet.Balance-amount, recipientWallet.ID, recipientWallet.Balance+amount)
	if err != nil {
		log.Println("error in updating the sender's and recipient's balance", err)
		return err
	}
	log.Println("the balance has been successfully updated")
	return nil
}

func (DB *SqlDataBase) DbCreateHistoryRecord(h *models.HistoryRecord) error {
	_, err := DB.DB.Exec(`insert into history (history_id, transfer_time , sender, receiver, amount) 
						values ($1, $2, $3, $4, $5)`,
		h.Id, h.Time, h.From, h.To, h.Amount)
	if err != nil {
		log.Println("error adding a new history record to the database", err)
		return err
	}
	log.Println("the entry of history record has been successfully added to the database")

	_, err = DB.DB.Exec(`insert into main_and_history (wallet_id_in_connection, history_id_in_connection)
						values ($2, $1), ($3, $1)`,
		h.Id, h.From, h.To)
	if err != nil {
		log.Println("error when adding a new record to the history_and_main table in the database", err)
		return err
	}
	log.Println("adding a new record to the history_and_main table in the database was successful")
	return nil

}

func (DB *SqlDataBase) DbGetHistory(walletId string) (*[]models.HistoryRecord, error) {
	history := []models.HistoryRecord{}
	rows, err := DB.DB.Query(`select h.history_id, h.transfer_time, h.sender, h.receiver, h.amount from history h 
		join main_and_history mah on mah.history_id_in_connection = h.history_id 
		where mah.wallet_id_in_connection = $1`,
		walletId)
	if err != nil {
		log.Println("error in DbGetHistory when when executing a request", err)
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		h := models.HistoryRecord{}
		err := rows.Scan(&h.Id, &h.Time, &h.From, &h.To, &h.Amount)
		if err != nil {
			log.Println("error in DbGetHistory when performing scan", err)
			return nil, err
		}
		ok, err := govalidator.ValidateStruct(h)
		if !ok || h.Amount < 0 {
			log.Println("invalid history type")
			return nil, err
		}
		history = append(history, h)
	}
	if len(history) == 0 {
		log.Println("error in DbGetHistory: no history records found")
		return nil, errors.New("no history records found")
	}
	return &history, nil
}

func (DB *SqlDataBase) DbGetWallet(walletId string) (*models.Wallet, error) {
	wallet := DB.DB.QueryRow(`select * from wallet_main where wallet_id = $1`,
		walletId)
	w := models.Wallet{}
	err := wallet.Scan(&w.ID, &w.Balance)
	if err != nil {
		log.Println("error in DbGetWallet when performing scan", err)
		return nil, err
	}

	log.Println("The data was successfully received", w)
	return &w, nil
}
