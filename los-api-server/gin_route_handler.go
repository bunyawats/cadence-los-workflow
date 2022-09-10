package main

import (
	"cadence-los-workflow/common"
	"fmt"
	"github.com/gin-gonic/gin"
	cadence_client "go.uber.org/cadence/client"
	"net/http"
)

type (
	GinHandlerHelper struct {
		M *common.MongodbHelper
		R *common.RabbitMqHelper
		H *common.LosHelper
		W cadence_client.Client
	}
)

func (g GinHandlerHelper) NlosNotificationHandler(c *gin.Context) {

	fmt.Println("Call NlosNotificationHandler API")

	var request common.DEResult
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	g.R.Publish2RabbitMQ(&request)

	c.JSON(http.StatusOK, request)
}

func (g GinHandlerHelper) CreateNewLoanApplicationHandler(c *gin.Context) {

	fmt.Println("Call SubmitFormOneHandler API")

	appID := c.Param("appId")

	if err := g.M.CreateNewLoanApplication(appID); err != nil {
		c.JSON(http.StatusConflict, gin.H{
			"error": err.Error(),
		})
		return
	}

	common.StartWorkflow(g.H, appID)
	c.JSON(http.StatusOK, gin.H{
		"appID": appID,
	})
}

func (g GinHandlerHelper) SubmitFormOneHandler(c *gin.Context) {

	fmt.Println("Call SubmitFormOneHandler API")

	var request common.LoanApplication
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}
	request.AppID = c.Param("appId")

	loanApp, err := g.M.SaveFormOne(&request)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": err.Error(),
		})
		return
	}

	common.CompleteActivity(g.M, g.W, request.AppID, "SUCCESS")

	c.JSON(http.StatusOK, loanApp)
}

func (g GinHandlerHelper) SubmitFormTwoHandler(c *gin.Context) {

	fmt.Println("Call SubmitFormTwoHandler API")

	var request common.LoanApplication
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}
	request.AppID = c.Param("appId")

	loanApp, err := g.M.SaveFormTwo(&request)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": err.Error(),
		})
		return
	}

	common.CompleteActivity(g.M, g.W, request.AppID, "SUCCESS")

	c.JSON(http.StatusOK, loanApp)
}

func (g GinHandlerHelper) QueryStateHandler(c *gin.Context) {

	fmt.Println("Call QueryStateHandler API")

	appID := c.Param("appId")
	taskToken := common.QueryApplicationState(g.M, g.H, appID)
	if taskToken != nil {
		c.JSON(http.StatusOK, gin.H{
			"app_id":      appID,
			"workflow_id": taskToken.WorkflowID,
			"run_id":      taskToken.RunID,
			"state":       taskToken.State,
		})
		return
	}
	c.JSON(http.StatusNotFound, gin.H{
		"error": "Token not found",
	})

}
