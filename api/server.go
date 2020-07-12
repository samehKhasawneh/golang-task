package api

import (
	"fmt"
	"log"
	"os"

	"golang-task/api/controllers"

	"github.com/joho/godotenv"
)

var server = controllers.Server{}

func init() {
	if err := godotenv.Load(); err != nil {
		log.Print("sad .env file found")
	}
}

func Run() {
	var err error
	err = godotenv.Load()
	if err != nil {
		log.Fatalf("Error getting env, %v", err)
	} else {
		fmt.Println("DB is working")
	}
	server.Initialize(os.Getenv("DB_DRIVER"), os.Getenv("DB_USER"), os.Getenv("DB_PASSWORD"), os.Getenv("DB_PORT"), os.Getenv("DB_HOST"), os.Getenv("DB_NAME"))
	server.InitTwitterClient()
	apiPort := fmt.Sprintf(":%s", os.Getenv("API_PORT"))
	fmt.Printf("Listening to port %s", apiPort)

	server.Run(apiPort)
}
