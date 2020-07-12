package controllers

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"golang-task/api/models"

	"github.com/dghubble/go-twitter/twitter"
	"github.com/dghubble/oauth1"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v7"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres" //postgres database driver
)

type Server struct {
	DB            *gorm.DB
	Router        *gin.Engine
	TwitterClient *twitter.Client
	RedisClient   *redis.Client
}

// Credentials for Twitter
type Credentials struct {
	ConsumerKey       string
	ConsumerSecret    string
	AccessToken       string
	AccessTokenSecret string
}

var errList = make(map[string]string)

func (server *Server) Initialize(Dbdriver, DbUser, DbPassword, DbPort, DbHost, DbName string) {

	var err error

	DBURL := fmt.Sprintf("host=%s port=%s user=%s dbname=%s sslmode=disable password=%s", DbHost, DbPort, DbUser, DbName, DbPassword)
	server.DB, err = gorm.Open(Dbdriver, DBURL)
	if err != nil {
		log.Fatal("Error connecting to postgres:", err)
	}

	//database migration
	server.DB.Debug().AutoMigrate(
		&models.User{},
		&models.Tweet{},
	)
	dsn := os.Getenv("REDIS_DSN")

	server.RedisClient = redis.NewClient(&redis.Options{
		Addr:     dsn,
		Password: "",
		DB:       0,
	})
	_, err = server.RedisClient.Ping().Result()
	if err != nil {
		panic(err)
	}
	server.Router = gin.Default()

	server.initializeRoutes()
}

func (server *Server) InitTwitterClient() {
	creds := Credentials{
		AccessToken:       os.Getenv("AccessToken"),
		AccessTokenSecret: os.Getenv("AccessTokenSecret"),
		ConsumerKey:       os.Getenv("ConsumerKey"),
		ConsumerSecret:    os.Getenv("ConsumerSecret"),
	}

	config := oauth1.NewConfig(creds.ConsumerKey, creds.ConsumerSecret)
	token := oauth1.NewToken(creds.AccessToken, creds.AccessTokenSecret)

	httpClient := config.Client(oauth1.NoContext, token)
	server.TwitterClient = twitter.NewClient(httpClient)

	verifyParams := &twitter.AccountVerifyParams{
		SkipStatus:   twitter.Bool(true),
		IncludeEmail: twitter.Bool(true),
	}

	_, _, err := server.TwitterClient.Accounts.VerifyCredentials(verifyParams)
	if err != nil {
		log.Fatal("Error getting Twitter client:", err)
	}
}

func (server *Server) Run(addr string) {
	log.Fatal(http.ListenAndServe(addr, server.Router))
}
