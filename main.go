package main

import (
	"Spec-Center/authorization"
	"Spec-Center/controller"
	"Spec-Center/db"

	// "casbin/casbin-http-role-example/authorization"

	"log"
	// "github.com/dgrijalva/jwt-go"
)

var SECRET_KEY = []byte("gosecretkey")

func main() {
	authorization.SECRET_KEY = SECRET_KEY
	log.Print("Starting the application...")
	controller.Start()
	db.Indexing()
}
