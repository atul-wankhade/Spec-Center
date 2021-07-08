package controller

import (
	"github.com/atul-wankhade/Spec-Center/authorization"
	"log"
	"net/http"

	"github.com/casbin/casbin"
	"github.com/gorilla/mux"
)

func Start() {
	// setup casbin auth rules
	authEnforcer, err := casbin.NewEnforcerSafe("./auth_model.conf", "./policy.csv")
	if err != nil {
		log.Fatal(err)
	}

	router := mux.NewRouter()

	//LOGIN & USER ADD
	router.HandleFunc("/login/{companyid}", LoginHandler).Methods("POST")
	router.Handle("/adduser", authorization.IsAuthorized(authEnforcer, AddUser)).Methods("POST")

	// ARTICLE
	router.Handle("/all_articles", authorization.IsAuthorized(authEnforcer, GetArticlesHandler)).Methods("GET")
	router.Handle("/article", authorization.IsAuthorized(authEnforcer, DeleteArticleHandler)).Methods("DELETE")
	router.Handle("/article", authorization.IsAuthorized(authEnforcer, UpdateArticleHandler)).Methods("PUT")
	router.Handle("/article", authorization.IsAuthorized(authEnforcer, CreateArticleHandler)).Methods("POST")

	//ROLE CHANGE :- only superadmin can change role of other user.
	router.Handle("/articlerole", authorization.IsAuthorized(authEnforcer, UpdateArticleRoleHandler)).Methods("PUT")
	router.Handle("/role", authorization.IsAuthorized(authEnforcer, UpdateCompanyRoleHandler)).Methods("PUT")

	// router.Use(authorization.Authorizer(authEnforcer))
	log.Print("Server started on localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", router))
}
