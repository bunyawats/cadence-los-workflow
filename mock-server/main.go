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

	router.POST("/nlos/notification", NLOS_NotificationHandler)
	router.POST("/create/app/:appId", CreateNewLoanApplicationHandler)
	router.POST("/submit/formone/:appId", SubmitFormOneHandler)
	router.POST("/submit/formtwo/:appId", SubmitFormTwoHandler)

	log.Printf(" [*] Waiting for message. To exit press CYRL+C")
	err := router.Run(":5500")
	if err != nil {
		log.Fatalln(err)
	}
}
