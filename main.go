package main

import (
	"log"

	"github.com/atul-wankhade/Spec-Center/controller"
	"github.com/atul-wankhade/Spec-Center/db"
)

func main() {
	log.Print("Starting the application...")
	db.Indexing()

	db.AddRoles()

	db.SuperadminEntry()
	//	go worker.Worker()s
	controller.Start()
}
