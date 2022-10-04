package main

import (
	"cadence-los-workflow/common"
	"cadence-los-workflow/model"
	"cadence-los-workflow/service"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"go.uber.org/cadence/client"
	"net/http"
	"time"
)

type (
	GinHandlerHelper struct {
		MongodbService  *service.MongodbService
		RabbitMqService *service.RabbitMqService
		WorkflowHelper  *common.WorkflowHelper
		CadenceClient   client.Client
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

	ex := service.StartWorkflow(g.WorkflowHelper)

	cb, _ := json.Marshal(appID)

	g.WorkflowHelper.SignalWorkflow(
		ex.ID,
		model.SignalName,
		&model.SignalPayload{
			Action:  model.Create,
			Content: cb,
		},
	)
	time.Sleep(time.Second)
	s := service.QueryApplicationState(g.MongodbService, g.WorkflowHelper, appID)
	service.AssertState(model.Created, s.State)

	cb, _ = json.Marshal(&model.LoanApplication{
		AppID: appID,
		Fname: "bunyawat",
		Lname: "singchai",
	})

	g.WorkflowHelper.SignalWorkflow(
		ex.ID,
		model.SignalName,
		&model.SignalPayload{
			Action:  model.SubmitFormOne,
			Content: cb,
		},
	)
	time.Sleep(time.Second)
	s = service.QueryApplicationState(g.MongodbService, g.WorkflowHelper, appID)
	service.AssertState(model.FormOneSubmitted, s.State)

	cb, _ = json.Marshal(&model.LoanApplication{
		AppID:   appID,
		Email:   "bunyawat.s@gmail.com",
		PhoneNo: "0868372995",
	})

	g.WorkflowHelper.SignalWorkflow(
		ex.ID,
		model.SignalName,
		&model.SignalPayload{
			Action:  model.SubmitFormTwo,
			Content: cb,
		},
	)
	time.Sleep(time.Second)
	s = service.QueryApplicationState(g.MongodbService, g.WorkflowHelper, appID)
	service.AssertState(model.FormTwoSubmitted, s.State)

	cb, _ = json.Marshal(appID)

	g.WorkflowHelper.SignalWorkflow(
		ex.ID,
		model.SignalName,
		&model.SignalPayload{
			Action:  model.SubmitDEOne,
			Content: cb,
		},
	)
	time.Sleep(time.Second)
	s = service.QueryApplicationState(g.MongodbService, g.WorkflowHelper, appID)
	service.AssertState(model.DEOneSubmitted, s.State)

	cb, _ = json.Marshal(&model.DEResult{
		AppID:  appID,
		Status: model.Approve,
	})

	g.WorkflowHelper.SignalWorkflow(
		ex.ID,
		model.SignalName,
		&model.SignalPayload{
			Action:  model.DEOneResultNotification,
			Content: cb,
		},
	)
	time.Sleep(time.Second)
	s = service.QueryApplicationState(g.MongodbService, g.WorkflowHelper, appID)
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

	ex := service.StartWorkflow(g.WorkflowHelper)
	g.WorkflowHelper.SignalWorkflow(
		ex.ID,
		model.SignalName,
		&model.SignalPayload{
			Action:  model.Create,
			Content: cb,
		},
	)
	time.Sleep(time.Second)
	s := service.QueryApplicationState(g.MongodbService, g.WorkflowHelper, appID)
	//common.AssertState(common.Created, s.State)

	c.JSON(http.StatusOK, gin.H{
		"appID":   appID,
		"s":       s.State,
		"content": s.Content,
	})
}

func (g GinHandlerHelper) SubmitFormOneHandler(c *gin.Context) {

	fmt.Println("Call SubmitFormOneHandler API")

	var r model.LoanApplication
	if err := c.ShouldBindJSON(&r); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}
	r.AppID = c.Param("appId")

	id, err := g.MongodbService.GetWorkflowIdByAppID(r.AppID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": err.Error(),
		})
		return
	}

	cb, _ := json.Marshal(&r)

	g.WorkflowHelper.SignalWorkflow(
		id,
		model.SignalName,
		&model.SignalPayload{
			Action:  model.SubmitFormOne,
			Content: cb,
		},
	)
	time.Sleep(time.Second)
	_ = service.QueryApplicationState(g.MongodbService, g.WorkflowHelper, r.AppID)
	//common.AssertState(common.FormOneSubmitted, state.State)

	c.JSON(http.StatusOK, &r)
}

func (g GinHandlerHelper) SubmitFormTwoHandler(c *gin.Context) {

	fmt.Println("Call SubmitFormTwoHandler API")

	var r model.LoanApplication
	if err := c.ShouldBindJSON(&r); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}
	r.AppID = c.Param("appId")

	id, err := g.MongodbService.GetWorkflowIdByAppID(r.AppID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": err.Error(),
		})
		return
	}

	cb, _ := json.Marshal(&r)

	g.WorkflowHelper.SignalWorkflow(
		id,
		model.SignalName,
		&model.SignalPayload{
			Action:  model.SubmitFormTwo,
			Content: cb,
		},
	)
	time.Sleep(time.Second)
	_ = service.QueryApplicationState(g.MongodbService, g.WorkflowHelper, r.AppID)
	//common.AssertState(common.FormTwoSubmitted, state.State)

	c.JSON(http.StatusOK, &r)
}

func (g GinHandlerHelper) SubmitDeOneHandler(c *gin.Context) {

	fmt.Println("Call SubmitDeOneHandler API")

	appID := c.Param("appId")

	id, err := g.MongodbService.GetWorkflowIdByAppID(appID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": err.Error(),
		})
		return
	}

	cb, _ := json.Marshal(appID)

	g.WorkflowHelper.SignalWorkflow(
		id,
		model.SignalName,
		&model.SignalPayload{
			Action:  model.SubmitDEOne,
			Content: cb,
		},
	)

	time.Sleep(time.Second)
	s := service.QueryApplicationState(g.MongodbService, g.WorkflowHelper, appID)
	//common.AssertState(common.DEOneSubmitted, s.State)

	c.JSON(http.StatusOK, gin.H{
		"appID":   appID,
		"s":       s.State,
		"content": s.Content,
	})
}

func (g GinHandlerHelper) SendDEResultHandler(c *gin.Context) {

	fmt.Println("Call SendDEResultHandler API")

	var request model.DEResult
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	g.RabbitMqService.PublishDEResult(&request)

	time.Sleep(time.Second)
	s := service.QueryApplicationState(g.MongodbService, g.WorkflowHelper, request.AppID)

	c.JSON(http.StatusOK, gin.H{
		"appID":   request.AppID,
		"s":       s.State,
		"content": s.Content,
	})
}

func (g GinHandlerHelper) QueryStateHandler(c *gin.Context) {

	fmt.Println("Call QueryStateHandler API")

	appID := c.Param("appId")
	s := service.QueryApplicationState(g.MongodbService, g.WorkflowHelper, appID)
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
