package main

import (
	"github.com/atul-wankhade/Spec-Center/authorization"
	"github.com/atul-wankhade/Spec-Center/controller"
	"github.com/atul-wankhade/Spec-Center/db"
	"log"
	// "github.com/dgrijalva/jwt-go"
)

var SECRET_KEY = []byte("gosecretkey")

func main() {
	authorization.SECRET_KEY = SECRET_KEY
	log.Print("Starting the application...")
	db.Indexing()
	db.SuperadminEntry()
	controller.Start()
}
