package main

import (
	"github.com/gin-gonic/gin"
	"log"
)

type (
	LoanApplication struct {
		AppID     string `bson:"appID" json:"appID"`
		Fname     string `bson:"fname" json:"fname"`
		Lname     string `bson:"lname" json:"lname"`
		Email     string `bson:"email" json:"email"`
		PhoneNo   string `bson:"phoneNo" json:"phoneNo"`
		TaskToken string `bson:"taskToken" json:"taskToken"`
		LastState string `bson:"lastState" json:"lastState"`
	}
)

func main() {

	router := gin.Default()

	router.POST("/nlos/notification/de_one", NLOS_NotificationHandler)
	router.POST("/nlos/create/application/:appId", CreateNewLoanApplicationHandler)
	router.POST("/nlos/submit/form_one/:appId", SubmitFormOneHandler)
	router.POST("/nlos/submit/form_two/:appId", SubmitFormTwoHandler)
	router.GET("/nlos/query/state/:appId", QueryStateHandler)

	log.Printf(" [*] Waiting for message. To exit press CYRL+C")
	err := router.Run(":5500")
	if err != nil {
		log.Fatalln(err)
	}
}
