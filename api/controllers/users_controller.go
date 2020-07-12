package controllers

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"golang-task/api/models"
	"golang-task/api/utils/formaterror"

	"github.com/gin-gonic/gin"
)

func (server *Server) CreateUser(c *gin.Context) {

	errList = map[string]string{}

	body, err := ioutil.ReadAll(c.Request.Body)
	fmt.Println(body)
	if err != nil {
		errList["Invalid_body"] = "Unable to get request"
		c.JSON(http.StatusUnprocessableEntity, gin.H{
			"error": errList,
		})
		return
	}

	user := models.User{}

	err = json.Unmarshal(body, &user)
	if err != nil {
		errList["Unmarshal_error"] = "Cannot unmarshal body"
		c.JSON(http.StatusUnprocessableEntity, gin.H{
			"error": errList,
		})
		return
	}
	user.Prepare()
	errorMessages := user.Validate("")
	if len(errorMessages) > 0 {
		errList = errorMessages
		c.JSON(http.StatusUnprocessableEntity, gin.H{
			"error": errList,
		})
		return
	}
	userCreated, err := user.SaveUser(server.DB)
	if err != nil {
		formattedError := formaterror.FormatError(err.Error())
		errList = formattedError
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": errList,
		})
		return
	}
	c.JSON(http.StatusCreated, gin.H{
		"response": userCreated,
	})
}
