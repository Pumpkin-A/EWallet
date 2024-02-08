package main

import (
	_ "github.com/lib/pq"
	"github.com/pumpkin-a/wallet/api"
	"github.com/pumpkin-a/wallet/models"
)

func main() {
	api.RunServer(models.GlobalConfig)
}
