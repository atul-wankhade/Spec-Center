package controller

import (
	"Spec-Center/authorization"
	"Spec-Center/db"
	"Spec-Center/model"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"github.com/dgrijalva/jwt-go"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func AddUser(response http.ResponseWriter, request *http.Request, claims jwt.MapClaims) {
	response.Header().Set("Content-Type", "application/json")
	var user model.User
	var role model.Roles
	keyVal := make(map[string]interface{})
	companyID := int(claims["companyid"].(float64))

	body, err := ioutil.ReadAll(request.Body)
	if err != nil {
		panic(err.Error())
	}
	err = json.Unmarshal(body, &keyVal)
	if err != nil {
		authorization.WriteError(http.StatusBadRequest, "BAD REQUEST", response, err)
		return
	}

	userRole := fmt.Sprintf("%v", keyVal["role"])
	userId, ok := keyVal["id"].(float64)
	if !ok {
		authorization.WriteError(http.StatusBadRequest, "Decode Error, Wrong userid provided, please check its type.", response, errors.New("wrong userid"))
		return
	}

	user.ID = int(userId)
	user.FirstName = fmt.Sprintf("%v", keyVal["firstname"])
	user.LastName = fmt.Sprintf("%v", keyVal["lastname"])
	user.Password = fmt.Sprintf("%v", keyVal["password"])
	user.Email = fmt.Sprintf("%v", keyVal["email"])

	//setting default value for  role
	if userRole == "superadmin" && userRole != "admin" && userRole != "member" && userRole != "anonymous" {
		authorization.WriteError(http.StatusBadRequest, "Invalid user role provided", response, errors.New("wrong role"))
		return
	}

	user.Password = getHash([]byte(user.Password))
	client := db.InitializeDatabase()
	defer client.Disconnect(context.Background())
	collection := client.Database("SPEC-CENTER").Collection("user")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_, err = collection.InsertOne(ctx, user)
	if err != nil {
		response.WriteHeader(http.StatusConflict)
		response.Write([]byte(`{"message":"` + fmt.Sprintf("User with userid: %d is already present in database, please provide different userid", user.ID ) + `"}`))
		return
	}
	// user role insertion in roles collection
	role.UserId = user.ID
	role.CompanyId = companyID
	role.Role = userRole

	collection = client.Database("SPEC-CENTER").Collection("role")
	_, err = collection.InsertOne(ctx, role)
	if err != nil {
		response.WriteHeader(http.StatusInternalServerError)
		response.Write([]byte(`{"message":"` + err.Error() + `"}`))
		return
	}

	go insertAllArticleRoleForNewUser(user.ID, companyID, userRole)

	response.WriteHeader(http.StatusAccepted)
	response.Write([]byte(`{"message":"` + fmt.Sprintf("User with userid: %d is added to company having id: %d with role: %s", int(userId), companyID, userRole) + `"}`))
}

// Inserting articlerole for newly created user for all article present in company
func insertAllArticleRoleForNewUser(userID int, companyID int, role string) {
	client := db.InitializeDatabase()
	defer client.Disconnect(context.Background())
	database := client.Database("SPEC-CENTER")
	companyRoleCollection := database.Collection("article")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	filter := primitive.M{"companyid": companyID}

	cursor, err := companyRoleCollection.Find(ctx, filter)
	if err != nil {
		log.Println(err)
	}
	defer cursor.Close(ctx)

	articleRoleCollection := database.Collection("articlerole")
	var articleRole model.ArticleRole
	var article model.Article

	for cursor.Next(ctx) {
		err := cursor.Decode(&article)
		if err != nil {
			log.Println(err)
		}
		articleRole.ArticleId = article.ArticleID
		articleRole.CompanyId = companyID
		articleRole.Role = role
		articleRole.UserId = userID

		_, err = articleRoleCollection.InsertOne(ctx, articleRole)
		if err != nil {
			log.Printf("Failed to add article role for article id : %d, user id : %d, error : %w", article.ArticleID, userID, err)
		}
		log.Printf("ArticleRole for  article id : %d, for user id : %d , for company id : %d added successfully", article.ArticleID, userID, companyID)
	}
}
