package controllers

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"

	"golang-task/api/auth"
	"golang-task/api/models"
	"golang-task/api/utils/formaterror"

	"github.com/dghubble/go-twitter/twitter"
	"github.com/gin-gonic/gin"
)

type QueryString struct {
	SearchString string `json:"search_string"`
	Limit        int    `json:"limit"`
}

// Tweet struct
type TwitterObj struct {
	Text         string    `json:"text"`
	Lang         string    `json:"lang"`
	ReplyCount   int       `json:"reply_count"`
	QuoteCount   int       `json:"quote_count"`
	RetweetCount int       `json:"retweet_count"`
	CreatedAt    time.Time `json:"created"`
}

func (server *Server) SearchTweets(c *gin.Context) {

	errList = map[string]string{}

	var tweets []TwitterObj
	var params *QueryString

	if err := c.BindJSON(&params); err != nil {
		errList["params"] = "Params error"
		c.JSON(http.StatusNotFound, gin.H{
			"error": errList,
		})
		return
	}

	search, _, err := server.TwitterClient.Search.Tweets(&twitter.SearchTweetParams{
		Query: params.SearchString,
		Count: params.Limit,
	})

	if err != nil {
		errList["No_tweets"] = "Something went wrong"
		c.JSON(http.StatusNotFound, gin.H{
			"error": errList,
		})
		return
	}

	for _, c := range search.Statuses {
		d, _ := time.Parse(time.RubyDate, c.CreatedAt)
		tweets = append(tweets, TwitterObj{CreatedAt: d, Lang: c.Lang, Text: c.Text, ReplyCount: c.ReplyCount,
			QuoteCount: c.QuoteCount, RetweetCount: c.RetweetCount})
	}

	c.JSON(http.StatusOK, gin.H{
		"response": tweets,
	})
}

func (server *Server) SaveTweets(c *gin.Context) {
	errList = map[string]string{}

	var tweets []models.Tweet
	body, _ := ioutil.ReadAll(c.Request.Body)
	if err := json.Unmarshal(body, &tweets); err != nil {
		errList["params"] = "Params error"
		c.JSON(http.StatusNotFound, gin.H{
			"error": errList,
		})
		return
	}
	tokenAuth, err := auth.ExtractTokenMetadata(c.Request)
	if err != nil {
		c.JSON(http.StatusUnauthorized, "unauthorized")
		return
	}
	uid, err := auth.FetchAuth(server.RedisClient, tokenAuth)
	if err != nil {
		c.JSON(http.StatusUnauthorized, "unauthorized")
		return
	}
	user := models.User{}
	err = server.DB.Debug().Model(models.User{}).Where("id = ?", uid).Take(&user).Error
	if err != nil {
		errList["Unauthorized"] = "Unauthorized"
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": errList,
		})
		return
	}
	for _, c := range tweets {
		c.AuthorID = uid
		c.Prepare()
		c.Validate()
		c.SaveTweet(server.DB)
	}
	c.JSON(http.StatusCreated, gin.H{})
}

func (server *Server) CreateTweet(c *gin.Context) {

	errList = map[string]string{}

	body, err := ioutil.ReadAll(c.Request.Body)
	if err != nil {
		errList["Invalid_body"] = "Unable to get request"
		c.JSON(http.StatusUnprocessableEntity, gin.H{
			"error": errList,
		})
		return
	}
	tweet := models.Tweet{}

	err = json.Unmarshal(body, &tweet)
	if err != nil {
		errList["Unmarshal_error"] = "Cannot unmarshal body"
		c.JSON(http.StatusUnprocessableEntity, gin.H{
			"error": errList,
		})
		return
	}
	tokenAuth, err := auth.ExtractTokenMetadata(c.Request)
	if err != nil {
		c.JSON(http.StatusUnauthorized, "unauthorized")
		return
	}
	uid, err := auth.FetchAuth(server.RedisClient, tokenAuth)
	if err != nil {
		c.JSON(http.StatusUnauthorized, "unauthorized")
		return
	}

	user := models.User{}
	err = server.DB.Debug().Model(models.User{}).Where("id = ?", uid).Take(&user).Error
	if err != nil {
		errList["Unauthorized"] = "Unauthorized"
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": errList,
		})
		return
	}

	tweet.AuthorID = uid

	tweet.Prepare()
	errorMessages := tweet.Validate()
	if len(errorMessages) > 0 {
		errList = errorMessages
		c.JSON(http.StatusUnprocessableEntity, gin.H{
			"error": errList,
		})
		return
	}

	tweetCreated, err := tweet.SaveTweet(server.DB)
	if err != nil {
		errList := formaterror.FormatError(err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": errList,
		})
		return
	}
	c.JSON(http.StatusCreated, gin.H{
		"response": tweetCreated,
	})
}

type Page struct {
	PageNum int `json:"page"`
}

func (server *Server) GetTweets(c *gin.Context) {

	tweet := models.Tweet{}
	var page Page
	if err := c.BindJSON(&page); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var offset int = 10 * page.PageNum
	tweets, err := tweet.FindAllTweets(server.DB, offset)
	if err != nil {
		errList["No_tweet"] = "Tweet Not Found"
		c.JSON(http.StatusNotFound, gin.H{
			"error": errList,
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"response": tweets,
	})
}

func (server *Server) GetTweet(c *gin.Context) {

	tweetID := c.Param("id")
	tid, err := strconv.ParseUint(tweetID, 10, 64)
	if err != nil {
		errList["Invalid_request"] = "Invalid Request"
		c.JSON(http.StatusBadRequest, gin.H{
			"error": errList,
		})
		return
	}
	tweet := models.Tweet{}

	tweetReceived, err := tweet.FindTweetByID(server.DB, tid)
	if err != nil {
		errList["No_tweet"] = "Tweet Not Found"
		c.JSON(http.StatusNotFound, gin.H{
			"error": errList,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"response": tweetReceived,
	})
}

func (server *Server) UpdateTweet(c *gin.Context) {

	errList = map[string]string{}

	tweetID := c.Param("id")
	tid, err := strconv.ParseUint(tweetID, 10, 64)
	if err != nil {
		errList["Invalid_request"] = "Invalid Request"
		c.JSON(http.StatusBadRequest, gin.H{
			"error": errList,
		})
		return
	}
	tokenAuth, err := auth.ExtractTokenMetadata(c.Request)
	if err != nil {
		c.JSON(http.StatusUnauthorized, "unauthorized")
		return
	}
	uid, err := auth.FetchAuth(server.RedisClient, tokenAuth)
	if err != nil {
		c.JSON(http.StatusUnauthorized, "unauthorized")
		return
	}

	origTweet := models.Tweet{}
	err = server.DB.Debug().Model(models.Tweet{}).Where("id = ?", tid).Take(&origTweet).Error
	if err != nil {
		errList["No_tweet"] = "Tweet Not Found"
		c.JSON(http.StatusNotFound, gin.H{
			"error": errList,
		})
		return
	}
	if uid != origTweet.AuthorID {
		errList["Unauthorized"] = "Unauthorized"
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": errList,
		})
		return
	}

	body, err := ioutil.ReadAll(c.Request.Body)
	if err != nil {
		errList["Invalid_body"] = "Unable to get request"
		c.JSON(http.StatusUnprocessableEntity, gin.H{
			"error": errList,
		})
		return
	}

	tweet := models.Tweet{}
	err = json.Unmarshal(body, &tweet)
	if err != nil {
		errList["Unmarshal_error"] = "Cannot unmarshal body"
		c.JSON(http.StatusUnprocessableEntity, gin.H{
			"error": errList,
		})
		return
	}
	tweet.ID = origTweet.ID
	tweet.AuthorID = origTweet.AuthorID

	tweet.Prepare()
	errorMessages := tweet.Validate()
	if len(errorMessages) > 0 {
		errList = errorMessages
		c.JSON(http.StatusUnprocessableEntity, gin.H{
			"error": errList,
		})
		return
	}
	tweetUpdated, err := tweet.UpdateTweet(server.DB)
	if err != nil {
		errList := formaterror.FormatError(err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": errList,
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"response": tweetUpdated,
	})
}

func (server *Server) DeleteTweet(c *gin.Context) {

	tweetID := c.Param("id")

	tid, err := strconv.ParseUint(tweetID, 10, 64)
	if err != nil {
		errList["Invalid_request"] = "Invalid Request"
		c.JSON(http.StatusBadRequest, gin.H{
			"error": errList,
		})
		return
	}

	tokenAuth, err := auth.ExtractTokenMetadata(c.Request)
	if err != nil {
		c.JSON(http.StatusUnauthorized, "unauthorized")
		return
	}
	_, err = auth.FetchAuth(server.RedisClient, tokenAuth)
	if err != nil {
		c.JSON(http.StatusUnauthorized, "unauthorized")
		return
	}

	tweet := models.Tweet{}
	err = server.DB.Debug().Model(models.Tweet{}).Where("id = ?", tid).Take(&tweet).Error
	if err != nil {
		errList["No_tweet"] = "Tweet Not Found"
		c.JSON(http.StatusNotFound, gin.H{
			"error": errList,
		})
		return
	}

	_, err = tweet.DeleteTweet(server.DB)
	if err != nil {
		errList["Other_error"] = "Please try again later"
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": errList,
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"response": "Tweet deleted",
	})
}

func (server *Server) GetUserTweets(c *gin.Context) {

	userID := c.Param("id")

	uid, err := strconv.ParseUint(userID, 10, 64)
	if err != nil {
		errList["Invalid_request"] = "Invalid Request"
		c.JSON(http.StatusBadRequest, gin.H{
			"error": errList,
		})
		return
	}
	tweet := models.Tweet{}
	tweets, err := tweet.FindUserTweets(server.DB, uint32(uid))
	if err != nil {
		errList["No_tweets"] = "No Tweets Found"
		c.JSON(http.StatusNotFound, gin.H{
			"error": errList,
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"response": tweets,
	})
}
