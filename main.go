package main

import (
	"github.com/atul-wankhade/Spec-Center/controller"
	"github.com/atul-wankhade/Spec-Center/db"
	"log"
	// "github.com/dgrijalva/jwt-go"
)

func main() {
	log.Print("Starting the application...")
	db.Indexing()
	db.SuperadminEntry()
	controller.Start()
}
