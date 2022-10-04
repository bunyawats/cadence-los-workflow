package main

import (
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
		Service       service.WorkflowService
		CadenceClient client.Client
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

	ex := g.Service.StartWorkflow()

	cb, _ := json.Marshal(appID)

	g.Service.WorkflowHelper.SignalWorkflow(
		ex.ID,
		model.SignalName,
		&model.SignalPayload{
			Action:  model.Create,
			Content: cb,
		},
	)
	time.Sleep(time.Second)
	s := g.Service.QueryApplicationState(appID)
	service.AssertState(model.Created, s.State)

	cb, _ = json.Marshal(&model.LoanApplication{
		AppID: appID,
		Fname: "bunyawat",
		Lname: "singchai",
	})

	g.Service.WorkflowHelper.SignalWorkflow(
		ex.ID,
		model.SignalName,
		&model.SignalPayload{
			Action:  model.SubmitFormOne,
			Content: cb,
		},
	)
	time.Sleep(time.Second)
	s = g.Service.QueryApplicationState(appID)
	service.AssertState(model.FormOneSubmitted, s.State)

	cb, _ = json.Marshal(&model.LoanApplication{
		AppID:   appID,
		Email:   "bunyawat.s@gmail.com",
		PhoneNo: "0868372995",
	})

	g.Service.WorkflowHelper.SignalWorkflow(
		ex.ID,
		model.SignalName,
		&model.SignalPayload{
			Action:  model.SubmitFormTwo,
			Content: cb,
		},
	)
	time.Sleep(time.Second)
	s = g.Service.QueryApplicationState(appID)
	service.AssertState(model.FormTwoSubmitted, s.State)

	cb, _ = json.Marshal(appID)

	g.Service.WorkflowHelper.SignalWorkflow(
		ex.ID,
		model.SignalName,
		&model.SignalPayload{
			Action:  model.SubmitDEOne,
			Content: cb,
		},
	)
	time.Sleep(time.Second)
	s = g.Service.QueryApplicationState(appID)
	service.AssertState(model.DEOneSubmitted, s.State)

	cb, _ = json.Marshal(&model.DEResult{
		AppID:  appID,
		Status: model.Approve,
	})

	g.Service.WorkflowHelper.SignalWorkflow(
		ex.ID,
		model.SignalName,
		&model.SignalPayload{
			Action:  model.DEOneResultNotification,
			Content: cb,
		},
	)
	time.Sleep(time.Second)
	s = g.Service.QueryApplicationState(appID)
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

	ex := g.Service.StartWorkflow()
	g.Service.WorkflowHelper.SignalWorkflow(
		ex.ID,
		model.SignalName,
		&model.SignalPayload{
			Action:  model.Create,
			Content: cb,
		},
	)
	time.Sleep(time.Second)
	s := g.Service.QueryApplicationState(appID)
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

	id, err := g.Service.MongodbService.GetWorkflowIdByAppID(r.AppID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": err.Error(),
		})
		return
	}

	cb, _ := json.Marshal(&r)

	g.Service.WorkflowHelper.SignalWorkflow(
		id,
		model.SignalName,
		&model.SignalPayload{
			Action:  model.SubmitFormOne,
			Content: cb,
		},
	)
	time.Sleep(time.Second)
	_ = g.Service.QueryApplicationState(r.AppID)
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

	id, err := g.Service.MongodbService.GetWorkflowIdByAppID(r.AppID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": err.Error(),
		})
		return
	}

	cb, _ := json.Marshal(&r)

	g.Service.WorkflowHelper.SignalWorkflow(
		id,
		model.SignalName,
		&model.SignalPayload{
			Action:  model.SubmitFormTwo,
			Content: cb,
		},
	)
	time.Sleep(time.Second)
	_ = g.Service.QueryApplicationState(r.AppID)
	//common.AssertState(common.FormTwoSubmitted, state.State)

	c.JSON(http.StatusOK, &r)
}

func (g GinHandlerHelper) SubmitDeOneHandler(c *gin.Context) {

	fmt.Println("Call SubmitDeOneHandler API")

	appID := c.Param("appId")

	id, err := g.Service.MongodbService.GetWorkflowIdByAppID(appID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": err.Error(),
		})
		return
	}

	cb, _ := json.Marshal(appID)

	g.Service.WorkflowHelper.SignalWorkflow(
		id,
		model.SignalName,
		&model.SignalPayload{
			Action:  model.SubmitDEOne,
			Content: cb,
		},
	)

	time.Sleep(time.Second)
	s := g.Service.QueryApplicationState(appID)
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

	g.Service.RabbitMqService.PublishDEResult(&request)

	time.Sleep(time.Second)
	s := g.Service.QueryApplicationState(request.AppID)

	c.JSON(http.StatusOK, gin.H{
		"appID":   request.AppID,
		"s":       s.State,
		"content": s.Content,
	})
}

func (g GinHandlerHelper) QueryStateHandler(c *gin.Context) {

	fmt.Println("Call QueryStateHandler API")

	appID := c.Param("appId")
	s := g.Service.QueryApplicationState(appID)
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
