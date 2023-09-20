package main

import (
	"log"

	"github.com/ayo-ajayi/bloggy/app"
	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(".env"); err != nil {
		log.Println(err)
		return
	}
	app := app.NewApp(":8080", app.BlogRouter())
	app.Start()
}
