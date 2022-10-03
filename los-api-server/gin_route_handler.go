package main

import (
	"cadence-los-workflow/common"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	cadence_client "go.uber.org/cadence/client"
	"net/http"
	"time"
)

type (
	GinHandlerHelper struct {
		MongodbHelper  *common.MongodbHelper
		RabbitMqHelper *common.RabbitMqHelper
		LosHelper      *common.LosHelper
		CadenceClient  cadence_client.Client
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

func (g GinHandlerHelper) AutoRunLosWorkflowHandler(c *gin.Context) {

	fmt.Println("Call SubmitFormOneHandler API")

	appID := c.Param("appId")

	ex := common.StartWorkflow(g.LosHelper)

	cb, _ := json.Marshal(appID)

	g.LosHelper.SignalWorkflow(
		ex.ID,
		common.SignalName,
		&common.SignalPayload{
			Action:  common.Create,
			Content: cb,
		},
	)
	time.Sleep(time.Second)
	s := common.QueryApplicationState(g.MongodbHelper, g.LosHelper, appID)
	common.AssertState(common.Created, s.State)

	cb, _ = json.Marshal(&common.LoanApplication{
		AppID: appID,
		Fname: "bunyawat",
		Lname: "singchai",
	})

	g.LosHelper.SignalWorkflow(
		ex.ID,
		common.SignalName,
		&common.SignalPayload{
			Action:  common.SubmitFormOne,
			Content: cb,
		},
	)
	time.Sleep(time.Second)
	s = common.QueryApplicationState(g.MongodbHelper, g.LosHelper, appID)
	common.AssertState(common.FormOneSubmitted, s.State)

	cb, _ = json.Marshal(&common.LoanApplication{
		AppID:   appID,
		Email:   "bunyawat.s@gmail.com",
		PhoneNo: "0868372995",
	})

	g.LosHelper.SignalWorkflow(
		ex.ID,
		common.SignalName,
		&common.SignalPayload{
			Action:  common.SubmitFormTwo,
			Content: cb,
		},
	)
	time.Sleep(time.Second)
	s = common.QueryApplicationState(g.MongodbHelper, g.LosHelper, appID)
	common.AssertState(common.FormTwoSubmitted, s.State)

	cb, _ = json.Marshal(appID)

	g.LosHelper.SignalWorkflow(
		ex.ID,
		common.SignalName,
		&common.SignalPayload{
			Action:  common.SubmitDEOne,
			Content: cb,
		},
	)
	time.Sleep(time.Second)
	s = common.QueryApplicationState(g.MongodbHelper, g.LosHelper, appID)
	common.AssertState(common.DEOneSubmitted, s.State)

	cb, _ = json.Marshal(&common.DEResult{
		AppID:  appID,
		Status: common.Approve,
	})

	g.LosHelper.SignalWorkflow(
		ex.ID,
		common.SignalName,
		&common.SignalPayload{
			Action:  common.DEOneResultNotification,
			Content: cb,
		},
	)
	time.Sleep(time.Second)
	s = common.QueryApplicationState(g.MongodbHelper, g.LosHelper, appID)
	fmt.Printf("current state: %v\n", s.State)

	c.JSON(http.StatusOK, gin.H{
		"appID": appID,
		"s":     s,
	})
}

func (g GinHandlerHelper) CreateNewLoanApplicationHandler(c *gin.Context) {

	fmt.Println("Call SubmitFormOneHandler API")

	appID := c.Param("appId")

	cb, _ := json.Marshal(appID)

	ex := common.StartWorkflow(g.LosHelper)
	g.LosHelper.SignalWorkflow(
		ex.ID,
		common.SignalName,
		&common.SignalPayload{
			Action:  common.Create,
			Content: cb,
		},
	)
	time.Sleep(time.Second)
	s := common.QueryApplicationState(g.MongodbHelper, g.LosHelper, appID)
	//common.AssertState(common.Created, s.State)

	c.JSON(http.StatusOK, gin.H{
		"appID":   appID,
		"s":       s.State,
		"content": s.Content,
	})
}

func (g GinHandlerHelper) SubmitFormOneHandler(c *gin.Context) {

	fmt.Println("Call SubmitFormOneHandler API")

	var r common.LoanApplication
	if err := c.ShouldBindJSON(&r); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}
	r.AppID = c.Param("appId")

	id, err := g.MongodbHelper.GetWorkflowIdByAppID(r.AppID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": err.Error(),
		})
		return
	}

	cb, _ := json.Marshal(&r)

	g.LosHelper.SignalWorkflow(
		id,
		common.SignalName,
		&common.SignalPayload{
			Action:  common.SubmitFormOne,
			Content: cb,
		},
	)
	time.Sleep(time.Second)
	_ = common.QueryApplicationState(g.MongodbHelper, g.LosHelper, r.AppID)
	//common.AssertState(common.FormOneSubmitted, state.State)

	c.JSON(http.StatusOK, &r)
}

func (g GinHandlerHelper) SubmitFormTwoHandler(c *gin.Context) {

	fmt.Println("Call SubmitFormTwoHandler API")

	var r common.LoanApplication
	if err := c.ShouldBindJSON(&r); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}
	r.AppID = c.Param("appId")

	id, err := g.MongodbHelper.GetWorkflowIdByAppID(r.AppID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": err.Error(),
		})
		return
	}

	cb, _ := json.Marshal(&r)

	g.LosHelper.SignalWorkflow(
		id,
		common.SignalName,
		&common.SignalPayload{
			Action:  common.SubmitFormTwo,
			Content: cb,
		},
	)
	time.Sleep(time.Second)
	_ = common.QueryApplicationState(g.MongodbHelper, g.LosHelper, r.AppID)
	//common.AssertState(common.FormTwoSubmitted, state.State)

	c.JSON(http.StatusOK, &r)
}

func (g GinHandlerHelper) SubmitDeOneHandler(c *gin.Context) {

	fmt.Println("Call SubmitDeOneHandler API")

	appID := c.Param("appId")

	id, err := g.MongodbHelper.GetWorkflowIdByAppID(appID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": err.Error(),
		})
		return
	}

	cb, _ := json.Marshal(appID)

	g.LosHelper.SignalWorkflow(
		id,
		common.SignalName,
		&common.SignalPayload{
			Action:  common.SubmitDEOne,
			Content: cb,
		},
	)

	time.Sleep(time.Second)
	s := common.QueryApplicationState(g.MongodbHelper, g.LosHelper, appID)
	//common.AssertState(common.DEOneSubmitted, s.State)

	c.JSON(http.StatusOK, gin.H{
		"appID":   appID,
		"s":       s.State,
		"content": s.Content,
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

	g.RabbitMqHelper.PublishDEResult(&request)

	time.Sleep(time.Second)
	s := common.QueryApplicationState(g.MongodbHelper, g.LosHelper, request.AppID)

	c.JSON(http.StatusOK, gin.H{
		"appID":   request.AppID,
		"s":       s.State,
		"content": s.Content,
	})
}

func (g GinHandlerHelper) QueryStateHandler(c *gin.Context) {

	fmt.Println("Call QueryStateHandler API")

	appID := c.Param("appId")
	s := common.QueryApplicationState(g.MongodbHelper, g.LosHelper, appID)
	if s != nil {
		c.JSON(http.StatusOK, gin.H{
			"app_id":  appID,
			"content": s.Content,
			"state":   s.State,
		})
		return
	}
	c.JSON(http.StatusNotFound, gin.H{
		"error": "Token not found",
	})

}
