package main

import (
	"log"

	"github.com/atul-wankhade/Spec-Center/controller"
	"github.com/atul-wankhade/Spec-Center/db"
	//"github.com/atul-wankhade/Spec-Center/worker"
	// "github.com/dgrijalva/jwt-go"
)

func main() {
	log.Print("Starting the application...")
	db.Indexing()

	db.AddRoles()

	db.SuperadminEntry()
	db.CompanyEntry()
	//	go worker.Worker()s
	controller.Start()
}
