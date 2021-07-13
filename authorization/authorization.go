package authorization

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/atul-wankhade/Spec-Center/db"
	"github.com/atul-wankhade/Spec-Center/model"
	"github.com/atul-wankhade/Spec-Center/utils"
	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/casbin/casbin"
	"github.com/dgrijalva/jwt-go"
)

var SECRET = utils.GetEnvVariable("SECRET")
var SECRET_KEY = []byte(SECRET)

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
				WriteError(http.StatusInternalServerError, "PARSE ERROR, Invalid or expired token!", w, err)
				return
			}
			if token.Valid {
				fmt.Println(token.Claims)
				claims = token.Claims.(jwt.MapClaims)
			}
		} else {
			fmt.Fprintf(w, "Not Authorized")
		}

		userRole := model.UserRole{}
		userEmail := claims["user_email"]
		companyID := vars["company_id"]
		client := db.InitializeDatabase()
		err := client.Database(utils.Database).Collection(utils.CompanyRolesCollection).FindOne(context.Background(), primitive.M{"email": userEmail, "company_id": companyID}).Decode(&userRole)
		if err != nil {
			WriteError(http.StatusUnauthorized, "Invalid companyID", w, err)
			return
		}

		claims["role"] = userRole.Role

		var url string
		if strings.Contains(r.URL.Path, "/article") && !strings.Contains(r.URL.Path, "/role") {
			url = utils.ArticleURLMatcher
		} else if strings.Contains(r.URL.Path, "/role") {
			url = utils.RoleURLMatcher
		} else {
			url = r.URL.Path
		}

		//casbin enforce
		res, err := e.EnforceSafe(userRole.Role, url, r.Method)
		if err != nil {
			WriteError(http.StatusInternalServerError, "ERROR", w, err)
			return
		}
		if res {
			fmt.Println("enforcer result is true")
		} else {
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

func IsAuthorizedForArticle(companyID, email, role string, httpMethod string, articleID primitive.ObjectID) bool {
	client := db.InitializeDatabase()
	defer client.Disconnect(context.Background())

	var articleRole model.ArticleRole
	filterForArticleRole := primitive.M{"article_id": articleID.Hex(), "email": email, "company_id": companyID}
	log.Println("######", companyID, articleID, role, email)
	err := client.Database(utils.Database).Collection(utils.ArticleRoleCollection).FindOne(context.Background(), filterForArticleRole).Decode(&articleRole)
	log.Println("#############", err)
	if err != nil {
		if role == "member" && (httpMethod == "PUT" || httpMethod == "DELETE") {
			log.Println("Unauthorized")
			return false
		}
		return true
	}

	if (articleRole.Role == "member" && (httpMethod == "PUT" || httpMethod == "DELETE")) || articleRole.Role == "anonymous" {
		return false
	}
	return true
}
