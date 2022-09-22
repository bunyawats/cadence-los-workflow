package main

import (
	"cadence-los-workflow/common"
	"fmt"
	"github.com/gin-gonic/gin"
	cadence_client "go.uber.org/cadence/client"
	"net/http"
	"time"
)

type (
	GinHandlerHelper struct {
		M *common.MongodbHelper
		R *common.RabbitMqHelper
		H *common.LosHelper
		W cadence_client.Client
	}
)

func (g GinHandlerHelper) RegisterRouter(r *gin.Engine) {

	r.POST("/nlos/autorun/application/:appId", g.AutoRunLosWorkflownHandler)

	r.POST("/nlos/create/application/:appId", g.CreateNewLoanApplicationHandler)
	r.POST("/nlos/submit/form_one/:appId", g.SubmitFormOneHandler)
	r.POST("/nlos/submit/form_two/:appId", g.SubmitFormTwoHandler)
	r.POST("/nlos/notification/de_one", g.NlosNotificationHandler)
	r.GET("/nlos/query/state/:appId", g.QueryStateHandler)
}

func assertState(expected, actual common.State) {
	if expected != actual {
		message := fmt.Sprintf("Workflow in wrong state. Expected %v Actual %v", expected, actual)
		panic(message)
	}
}

func (g GinHandlerHelper) AutoRunLosWorkflownHandler(c *gin.Context) {

	fmt.Println("Call SubmitFormOneHandler API")

	appID := c.Param("appId")

	if err := g.M.CreateNewLoanApplication(appID); err != nil {
		c.JSON(http.StatusConflict, gin.H{
			"error": err.Error(),
		})
		return
	}

	ex := common.StartWorkflow(g.H, appID)

	g.H.SignalWorkflow(ex.ID, common.SignalName, &common.SignalPayload{Action: common.Create})
	time.Sleep(time.Second * 5)
	state := common.QueryApplicationState(g.M, g.H, appID)
	assertState(common.Created, state.State)

	g.H.SignalWorkflow(ex.ID, common.SignalName, &common.SignalPayload{Action: common.SubmitFormOne})
	time.Sleep(time.Second * 5)
	state = common.QueryApplicationState(g.M, g.H, appID)
	assertState(common.FormOneSubmitted, state.State)

	g.H.SignalWorkflow(ex.ID, common.SignalName, &common.SignalPayload{Action: common.SubmitFormTwo})
	time.Sleep(time.Second * 5)
	state = common.QueryApplicationState(g.M, g.H, appID)
	assertState(common.FormTwoSubmitted, state.State)

	g.H.SignalWorkflow(ex.ID, common.SignalName, &common.SignalPayload{Action: common.SubmitDEOne})
	time.Sleep(time.Second * 5)
	state = common.QueryApplicationState(g.M, g.H, appID)
	assertState(common.DEOneSubmitted, state.State)

	content := common.Approve
	g.H.SignalWorkflow(ex.ID, common.SignalName, &common.SignalPayload{Action: common.DEOneResultNotification, Content: content})
	time.Sleep(time.Second * 5)
	state = common.QueryApplicationState(g.M, g.H, appID)
	fmt.Printf("current state: %v", state)

	c.JSON(http.StatusOK, gin.H{
		"appID": appID,
		"state": state,
	})
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
	state := common.QueryApplicationState(g.M, g.H, appID)
	fmt.Println(state)

	c.JSON(http.StatusOK, gin.H{
		"appID":   appID,
		"state":   state.State,
		"content": state.Content,
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

func (g GinHandlerHelper) QueryStateHandler(c *gin.Context) {

	fmt.Println("Call QueryStateHandler API")

	appID := c.Param("appId")
	queryResult := common.QueryApplicationState(g.M, g.H, appID)
	if queryResult != nil {
		c.JSON(http.StatusOK, gin.H{
			"app_id":  appID,
			"content": queryResult.Content,
			"state":   queryResult.State,
		})
		return
	}
	c.JSON(http.StatusNotFound, gin.H{
		"error": "Token not found",
	})

}
