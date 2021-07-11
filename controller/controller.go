package controller

import (
	"log"
	"net/http"

	"github.com/atul-wankhade/Spec-Center/authorization"
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
	router.HandleFunc("/login", LoginHandler).Methods("POST")

	router.Handle("/company/{company_id}/user", authorization.IsAuthorized(authEnforcer, AddUser)).Methods("POST")

	// ARTICLE
	router.Handle("/company/{company_id}/article", authorization.IsAuthorized(authEnforcer, GetArticlesHandler)).Methods("GET")
	router.Handle("/company/{company_id}/article/{article_id}/articles", authorization.IsAuthorized(authEnforcer, GetSingleArticleHandler)).Methods("GET")
	router.Handle("/company/{company_id}/article/{article_id}/articles", authorization.IsAuthorized(authEnforcer, DeleteArticleHandler)).Methods("DELETE")
	router.Handle("/company/{company_id}/article/{article_id}/articles", authorization.IsAuthorized(authEnforcer, UpdateArticleHandler)).Methods("PUT")
	router.Handle("/company/{company_id}/articles", authorization.IsAuthorized(authEnforcer, CreateArticleHandler)).Methods("POST")

	// //ROLE CHANGE :- only superadmin can change role of other user.
	// router.Handle("/company/{company_id}/user/{email}/article/{article_id}/role", authorization.IsAuthorized(authEnforcer, UpdateArticleRoleHandler)).Methods("PUT")
	router.Handle("/company/{company_id}/user/{email}/role", authorization.IsAuthorized(authEnforcer, UpdateCompanyRoleHandler)).Methods("PUT")

	// router.Use(authorization.Authorizer(authEnforcer))
	log.Print("Server started on localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", router))
}
