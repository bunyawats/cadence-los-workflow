package main

import (
	"github.com/gin-gonic/gin"
	"log"
)

type ()

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
