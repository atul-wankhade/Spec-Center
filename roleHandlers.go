package main

import (
	"Spec-Center/authorization"
	"Spec-Center/model"
	"encoding/json"
	"net/http"

	"github.com/dgrijalva/jwt-go"
)

func UpdateArticleRoleHandler(w http.ResponseWriter, r *http.Request, claims jwt.MapClaims) {
	var articleRole model.ArticleRole
	err := json.NewDecoder(r.Body).Decode(&articleRole)
	if err != nil {
		authorization.WriteError(http.StatusBadRequest, "DECODE ERROR", w, err)
		return
	}

}
