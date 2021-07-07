package main

import (
	"Spec-Center/authorization"
	"Spec-Center/controller"

	// "casbin/casbin-http-role-example/authorization"

	"log"
	// "github.com/dgrijalva/jwt-go"
)

var SECRET_KEY = []byte("gosecretkey")

func main() {
	authorization.SECRET_KEY = SECRET_KEY
	log.Print("Starting the application...")

	// client := utils.InitializeDatabase()
	// defer client.Disconnect(context.Background())

	controller.Start()

}
