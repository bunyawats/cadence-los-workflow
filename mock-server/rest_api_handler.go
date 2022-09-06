package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
)

func NLOS_NotificationHandler(c *gin.Context) {

	fmt.Println("Call NLOS_NotificationHandler API")

	var request DEResult
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	Publish2RabbitMQ(&request)

	c.JSON(http.StatusOK, request)
}

func CreateNewLoanApplicationHandler(c *gin.Context) {

	fmt.Println("Call SubmitFormOneHandler API")

	appID := c.Param("appId")

	if err := CreateNewLoanApplication(appID); err != nil {
		c.JSON(http.StatusConflict, gin.H{
			"error": err.Error(),
		})
		return
	}

	StartWorkflow(appID)
	c.JSON(http.StatusOK, gin.H{
		"appID": appID,
	})
}

func SubmitFormOneHandler(c *gin.Context) {

	fmt.Println("Call SubmitFormOneHandler API")

	var request LoanApplication
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}
	request.AppID = c.Param("appId")

	loanApp, err := SaveFormOne(&request)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": err.Error(),
		})
		return
	}

	CompleteActivity(request.AppID, "SUCCESS")

	c.JSON(http.StatusOK, loanApp)
}

func SubmitFormTwoHandler(c *gin.Context) {

	fmt.Println("Call SubmitFormTwoHandler API")

	var request LoanApplication
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}
	request.AppID = c.Param("appId")

	loanApp, err := SaveFormTwo(&request)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": err.Error(),
		})
		return
	}

	CompleteActivity(request.AppID, "SUCCESS")

	c.JSON(http.StatusOK, loanApp)
}

func QueryStateHandler(c *gin.Context) {

	fmt.Println("Call QueryStateHandler API")

	appID := c.Param("appId")
	taskToken := QueryApplicationState(appID)
	if taskToken != nil {
		c.JSON(http.StatusOK, gin.H{
			"appID":      appID,
			"WorkflowID": taskToken.WorkflowID,
			"RunID":      taskToken.RunID,
		})
		return
	}
	c.JSON(http.StatusNotFound, gin.H{
		"error": "Token not found",
	})

}
