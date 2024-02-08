package models

type Wallet struct {
	ID      string  `json:"id" valid:"uuid"`
	Balance float64 `json:"balance" valid:"float"`
}

type Transfer struct {
	To     string  `json:"to" valid:"uuid"`
	Amount float64 `json:"amount" valid:"float"`
}

type Config struct {
	ServerPort string
}

var GlobalConfig = Config{
	ServerPort: ":8080",
}

type HistoryRecord struct {
	Id     string  `json:"id" valid:"uuid"`
	Time   string  `json:"time"`
	From   string  `json:"from" valid:"uuid"`
	To     string  `json:"to" valid:"uuid"`
	Amount float64 `json:"amount" valid:"float"`
}

//
