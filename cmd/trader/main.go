package main

import (
	"log"

	_ "github.com/joho/godotenv/autoload"
	_ "kraken-trader/pkg/logger"
)

func main() {
	if err := Execute(); err != nil {
		log.Fatal(err)
	}
}
