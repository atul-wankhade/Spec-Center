package authorization

import (
	"Spec-Center/model"
	"context"
	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/casbin/casbin"
	"github.com/dgrijalva/jwt-go"
	"go.mongodb.org/mongo-driver/mongo"
)
var SECRET_KEY = []byte("gosecretkey")
var Client *mongo.Client

func IsAuthorized(e *casbin.Enforcer, endpoint func(http.ResponseWriter, *http.Request, jwt.MapClaims)) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		var claims jwt.MapClaims
		if r.Header["Token"] != nil {
			token, err := jwt.Parse(r.Header["Token"][0], func(token *jwt.Token) (interface{}, error) {
				if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, fmt.Errorf("there was an error")
				}
				return SECRET_KEY, nil
			})

			if err != nil {
				WriteError(http.StatusInternalServerError, "PARSE ERROR, Invalid token!", w, err)
				return
			}
			if token.Valid {
				fmt.Println(token.Claims)
				claims = token.Claims.(jwt.MapClaims)
			}
		} else {
			fmt.Fprintf(w, "Not Authorized")
		}
		role := claims["userrole"]
		companyID, ok := claims["companyid"].(float64)
k := claims["company_id"].(float64)

		if !ok {
			WriteError(http.StatusInternalServerError, "ERROR", w, errors.New("interface conversion error"))
			return
		}

		//casbin enforce
		res, err := e.EnforceSafe(role, r.URL.Path, r.Method)
		if err != nil {
			WriteError(http.StatusInternalServerError, "ERROR", w, err)
			return
		}
		if res {
			fmt.Println("enforcer result is true")
		} else {
			//@
			fmt.Println("@@@@@@@@",role)
			WriteError(http.StatusForbidden, "FORBIDDEN, unauthorized", w, errors.New("unauthorized"))
			return
		}

		if role != "superadmin" && strings.Contains(r.URL.Path, "/article") {
			var article model.ArticleRole
			articleID := vars["articleid"]
			filter := []primitive.M{{"articleid": articleID}, {"companyid": companyID}}
			collection := Client.Database("SPEC-CENTER").Collection("articlerole")
			err := collection.FindOne(context.Background(), primitive.M{"$and": filter}).Decode(&article)
			if err != nil {
				WriteError(http.StatusInternalServerError, "DECODE ERROR", w, errors.New("unable to decode article"))
				return
			}
			articleRole := article.Role

			if (articleRole == "member" && r.Method == "GET") || articleRole == "admin" {
				endpoint(w, r, claims)
			} else {
				WriteError(http.StatusUnauthorized, "UNAUTHORIZED", w, errors.New("user unauthorized"))
				return
			}
		}
		endpoint(w, r, claims)
	})

}

func WriteError(status int, message string, w http.ResponseWriter, err error) {
	log.Print("ERROR: ", err.Error())
	w.WriteHeader(status)
	w.Write([]byte(message))
}
