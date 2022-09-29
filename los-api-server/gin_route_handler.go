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

	r.POST("/nlos/autorun/application/:appId", g.AutoRunLosWorkflowHandler)

	r.POST("/nlos/create/application/:appId", g.CreateNewLoanApplicationHandler)
	r.POST("/nlos/submit/form_one/:appId", g.SubmitFormOneHandler)
	r.POST("/nlos/submit/form_two/:appId", g.SubmitFormTwoHandler)
	r.POST("/nlos/submit/de_one/:appId", g.SubmitDeOneHandler)
	r.POST("/nlos/notification/de_one", g.SendDEResultHandler)
	r.GET("/nlos/query/state/:appId", g.QueryStateHandler)
}

func assertState(expected, actual common.State) {
	if expected != actual {
		message := fmt.Sprintf("Workflow in wrong state. Expected %v Actual %v", expected, actual)
		panic(message)
	}
}

func (g GinHandlerHelper) AutoRunLosWorkflowHandler(c *gin.Context) {

	fmt.Println("Call SubmitFormOneHandler API")

	appID := c.Param("appId")

	ex := common.StartWorkflow(g.H)

	g.H.SignalWorkflow(
		ex.ID,
		common.SignalName,
		&common.SignalPayload{
			Action:  common.Create,
			Content: common.Content{"appID": appID},
		},
	)
	time.Sleep(time.Second)
	state := common.QueryApplicationState(g.M, g.H, appID)
	assertState(common.Created, state.State)

	g.H.SignalWorkflow(
		ex.ID,
		common.SignalName,
		&common.SignalPayload{
			Action: common.SubmitFormOne,
			Content: common.Content{
				"appID": appID,
				"fname": "bunyawat",
				"lname": "singchai",
			},
		},
	)
	time.Sleep(time.Second)
	state = common.QueryApplicationState(g.M, g.H, appID)
	assertState(common.FormOneSubmitted, state.State)

	g.H.SignalWorkflow(
		ex.ID,
		common.SignalName,
		&common.SignalPayload{
			Action: common.SubmitFormTwo,
			Content: common.Content{
				"appID":   appID,
				"email":   "bunyawat.s@gmail.com",
				"phoneNo": "0868372995",
			},
		},
	)
	time.Sleep(time.Second)
	state = common.QueryApplicationState(g.M, g.H, appID)
	assertState(common.FormTwoSubmitted, state.State)

	g.H.SignalWorkflow(
		ex.ID,
		common.SignalName,
		&common.SignalPayload{
			Action:  common.SubmitDEOne,
			Content: common.Content{"appID": appID},
		},
	)
	time.Sleep(time.Second)
	state = common.QueryApplicationState(g.M, g.H, appID)
	assertState(common.DEOneSubmitted, state.State)

	g.H.SignalWorkflow(
		ex.ID,
		common.SignalName,
		&common.SignalPayload{
			Action: common.DEOneResultNotification,
			Content: common.Content{
				"appID":  appID,
				"status": common.Approve,
			},
		},
	)
	time.Sleep(time.Second)
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

	ex := common.StartWorkflow(g.H)
	g.H.SignalWorkflow(
		ex.ID,
		common.SignalName,
		&common.SignalPayload{
			Action:  common.Create,
			Content: common.Content{"appID": appID},
		},
	)
	time.Sleep(time.Second)
	state := common.QueryApplicationState(g.M, g.H, appID)
	//assertState(common.Created, state.State)

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

	taskTokenStr, err := g.M.GetTokenByAppID(request.AppID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": err.Error(),
		})
		return
	}

	taskToken := common.DeserializeTaskToken([]byte(taskTokenStr))
	g.H.SignalWorkflow(
		taskToken.WorkflowID,
		common.SignalName,
		&common.SignalPayload{
			Action: common.SubmitFormOne,
			Content: common.Content{
				"appID": request.AppID,
				"fname": request.Fname,
				"lname": request.Lname,
			},
		},
	)
	time.Sleep(time.Second)
	_ = common.QueryApplicationState(g.M, g.H, request.AppID)
	//assertState(common.FormOneSubmitted, state.State)

	c.JSON(http.StatusOK, &request)
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

	taskTokenStr, err := g.M.GetTokenByAppID(request.AppID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": err.Error(),
		})
		return
	}

	taskToken := common.DeserializeTaskToken([]byte(taskTokenStr))
	g.H.SignalWorkflow(
		taskToken.WorkflowID,
		common.SignalName,
		&common.SignalPayload{
			Action: common.SubmitFormTwo,
			Content: common.Content{
				"appID":   request.AppID,
				"email":   request.Email,
				"phoneNo": request.PhoneNo,
			},
		},
	)
	time.Sleep(time.Second)
	_ = common.QueryApplicationState(g.M, g.H, request.AppID)
	//assertState(common.FormTwoSubmitted, state.State)

	c.JSON(http.StatusOK, &request)
}

func (g GinHandlerHelper) SubmitDeOneHandler(c *gin.Context) {

	fmt.Println("Call SubmitDeOneHandler API")

	appID := c.Param("appId")

	taskTokenStr, err := g.M.GetTokenByAppID(appID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": err.Error(),
		})
		return
	}

	taskToken := common.DeserializeTaskToken([]byte(taskTokenStr))

	g.H.SignalWorkflow(
		taskToken.WorkflowID,
		common.SignalName,
		&common.SignalPayload{
			Action:  common.SubmitDEOne,
			Content: common.Content{"appID": appID},
		},
	)

	time.Sleep(time.Second)
	state := common.QueryApplicationState(g.M, g.H, appID)
	//assertState(common.DEOneSubmitted, state.State)

	c.JSON(http.StatusOK, gin.H{
		"appID":   appID,
		"state":   state.State,
		"content": state.Content,
	})
}

func (g GinHandlerHelper) SendDEResultHandler(c *gin.Context) {

	fmt.Println("Call SendDEResultHandler API")

	var request common.DEResult
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	g.R.PublishDEResult(&request)

	time.Sleep(time.Second)
	state := common.QueryApplicationState(g.M, g.H, request.AppID)

	c.JSON(http.StatusOK, gin.H{
		"appID":   request.AppID,
		"state":   state.State,
		"content": state.Content,
	})
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
