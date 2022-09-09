package main

import (
	"cadence-los-workflow/common"
	"fmt"
	"github.com/gin-gonic/gin"
	cadence_client "go.uber.org/cadence/client"
	"net/http"
	"os"
)

const (
	mongoUri      = "MONGO_URI"
	mongoDatabase = "MONGO_DATABASE"
)

var (
	m              *common.MongodbHelper
	h              common.LosHelper
	workflowClient cadence_client.Client
)

func init() {

	m = common.NewMongodbHelper(common.MongodbConfig{
		MongoUri:      os.Getenv(mongoUri),
		MongoDatabase: os.Getenv(mongoDatabase),
	})

	h.SetupServiceConfig()
	var err error
	workflowClient, err = h.Builder.BuildCadenceClient()
	if err != nil {
		panic(err)
	}
}

func NLOS_NotificationHandler(c *gin.Context) {

	fmt.Println("Call NLOS_NotificationHandler API")

	var request common.DEResult
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

	if err := m.CreateNewLoanApplication(appID); err != nil {
		c.JSON(http.StatusConflict, gin.H{
			"error": err.Error(),
		})
		return
	}

	common.StartWorkflow(&h, appID)
	c.JSON(http.StatusOK, gin.H{
		"appID": appID,
	})
}

func SubmitFormOneHandler(c *gin.Context) {

	fmt.Println("Call SubmitFormOneHandler API")

	var request common.LoanApplication
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}
	request.AppID = c.Param("appId")

	loanApp, err := m.SaveFormOne(&request)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": err.Error(),
		})
		return
	}

	common.CompleteActivity(m, workflowClient, request.AppID, "SUCCESS")

	c.JSON(http.StatusOK, loanApp)
}

func SubmitFormTwoHandler(c *gin.Context) {

	fmt.Println("Call SubmitFormTwoHandler API")

	var request common.LoanApplication
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}
	request.AppID = c.Param("appId")

	loanApp, err := m.SaveFormTwo(&request)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": err.Error(),
		})
		return
	}

	common.CompleteActivity(m, workflowClient, request.AppID, "SUCCESS")

	c.JSON(http.StatusOK, loanApp)
}

func QueryStateHandler(c *gin.Context) {

	fmt.Println("Call QueryStateHandler API")

	appID := c.Param("appId")
	taskToken := common.QueryApplicationState(m, &h, appID)
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
