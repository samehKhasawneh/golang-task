package controllers

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"golang-task/api/auth"
	"golang-task/api/models"
	"golang-task/api/security"
	"golang-task/api/utils/formaterror"

	"github.com/gin-gonic/gin"

	"golang.org/x/crypto/bcrypt"
)

func (server *Server) Login(c *gin.Context) {

	errList = map[string]string{}

	body, err := ioutil.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{
			"error": "Unable to get request",
		})
		return
	}
	user := models.User{}
	err = json.Unmarshal(body, &user)
	if err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{
			"error": "Cannot unmarshal body",
		})
		return
	}
	user.Prepare()
	errorMessages := user.Validate("login")
	if len(errorMessages) > 0 {
		c.JSON(http.StatusUnprocessableEntity, gin.H{
			"error": errorMessages,
		})
		return
	}
	tokens, err := server.SignIn(user.Email, user.Password)
	if err != nil {
		formattedError := formaterror.FormatError(err.Error())
		c.JSON(http.StatusUnprocessableEntity, gin.H{
			"error": formattedError,
		})
		return
	}
	c.JSON(http.StatusOK, tokens)
}

func (server *Server) SignIn(email, password string) (map[string]string, error) {

	var err error

	user := models.User{}

	err = server.DB.Debug().Model(models.User{}).Where("email = ?", email).Take(&user).Error
	if err != nil {
		fmt.Println("this is the error getting the user: ", err)
		return nil, err
	}
	err = security.VerifyPassword(user.Password, password)
	if err != nil && err == bcrypt.ErrMismatchedHashAndPassword {
		fmt.Println("this is the error hashing the password: ", err)
		return nil, err
	}
	token, err := auth.CreateToken(user.ID)
	if err != nil {
		return nil, err
	}
	err = auth.CreateAuth(server.RedisClient, user.ID, token)
	if err != nil {
		return nil, err
	}
	tokens := map[string]string{
		"access_token":  token.AccessToken,
		"refresh_token": token.RefreshToken,
	}
	return tokens, nil
}

func (server *Server) Logout(c *gin.Context) {
	au, err := auth.ExtractTokenMetadata(c.Request)
	if err != nil {
		c.JSON(http.StatusUnauthorized, "unauthorized")
		return
	}
	deleted, delErr := auth.DeleteAuth(server.RedisClient, au.AccessUuid)
	if delErr != nil || deleted == 0 {
		c.JSON(http.StatusUnauthorized, "unauthorized")
		return
	}
	c.JSON(http.StatusOK, "Successfully logged out")
}
