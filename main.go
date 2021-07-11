package main

import (
	"log"
	"github.com/atul-wankhade/Spec-Center/controller"
	"github.com/atul-wankhade/Spec-Center/db"
	// "github.com/dgrijalva/jwt-go"
)

func main() {
	log.Print("Starting the application...")
	db.Indexing()
	db.AddRoles()
	db.SuperadminEntry()
	db.CompanyEntry()
	// go worker.Worker()

	controller.Start()
}

// p, admin, /company/*/article, GET
// p, admin, /article, GET
// p, admin, /article, DELETE
// p, admin, /company/*/article/*/articles, PUT

// p, member, /company/*/article, GET
// p, member, /company/*/article/*/articles, GET
// p, member, /company/*/article/*/articles, DELETE
// p, member, /company/*/article/*/articles, PUT
