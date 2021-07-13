package main

import (
	"github.com/atul-wankhade/Spec-Center/controller"
	"github.com/atul-wankhade/Spec-Center/db"
	"log"
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
