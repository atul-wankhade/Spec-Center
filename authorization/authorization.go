package authorization

import (
	//"Spec-Center/model"
	//"context"
	//"go.mongodb.org/mongo-driver/bson/primitive"
	//"time"

	"errors"
	"fmt"
	"log"
	"net/http"

	"github.com/atul-wankhade/Spec-Center/utils"

	//"strings"

	"github.com/casbin/casbin"
	"github.com/dgrijalva/jwt-go"
)

var SECRET = utils.GetEnvVariable("SECRET")
var SECRET_KEY = []byte(SECRET)

func IsAuthorized(e *casbin.Enforcer, endpoint func(http.ResponseWriter, *http.Request, jwt.MapClaims)) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		//vars := mux.Vars(r)
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
		//companyID, ok := claims["companyid"].(float64)
		//if !ok {
		//	WriteError(http.StatusInternalServerError, "ERROR", w, errors.New("interface conversion error"))
		//	return
		//}

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

		fmt.Println("FINISHED")
		endpoint(w, r, claims)

	})

}

func WriteError(status int, message string, w http.ResponseWriter, err error) {
	log.Print("ERROR: ", err.Error())
	w.WriteHeader(status)
	w.Write([]byte(message))
}
