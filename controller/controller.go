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
	router.Handle("/user", authorization.IsAuthorized(authEnforcer, AddUser)).Methods("POST")

	// To add user for particular company by updating user with its role in user_role collection
	router.Handle("/company/{company_id}/user/{user_id}/role", authorization.IsAuthorized(authEnforcer, UpdateCompanyRoleHandler)).Methods("POST")

	// ARTICLE
	router.Handle("/company/{company_id}/article", authorization.IsAuthorized(authEnforcer, GetArticlesHandler)).Methods("GET")
	router.Handle("/company/{company_id}/article/{article_id}", authorization.IsAuthorized(authEnforcer, GetSingleArticleHandler)).Methods("GET")
	router.Handle("/company/{company_id}/article/{article_id}", authorization.IsAuthorized(authEnforcer, DeleteArticleHandler)).Methods("DELETE")
	router.Handle("/company/{company_id}/article/{article_id}", authorization.IsAuthorized(authEnforcer, UpdateArticleHandler)).Methods("PUT")
	router.Handle("/company/{company_id}/article", authorization.IsAuthorized(authEnforcer, CreateArticleHandler)).Methods("POST")

	// //ROLE CHANGE :- only superadmin can change role of other user.
	router.Handle("/company/{company_id}/user/{user_id}/article/{article_id}/role", authorization.IsAuthorized(authEnforcer, UpdateArticleRoleHandler)).Methods("PUT")
	router.Handle("/company/{company_id}/user/{user_id}/role", authorization.IsAuthorized(authEnforcer, UpdateCompanyRoleHandler)).Methods("PUT")
	// router.Use(authorization.Authorizer(authEnforcer))
	log.Print("Server started on localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", router))
}
