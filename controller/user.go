package controller

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"github.com/atul-wankhade/Spec-Center/authorization"
	"github.com/atul-wankhade/Spec-Center/db"
	"github.com/atul-wankhade/Spec-Center/model"
	"github.com/atul-wankhade/Spec-Center/utils"
	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/dgrijalva/jwt-go"
)

// "/company/{company_id}/user"
func AddUser(response http.ResponseWriter, request *http.Request, claims jwt.MapClaims) {
	response.Header().Set("Content-Type", "application/json")
	var user model.User
	var userRole model.UserRole
	keyVal := make(map[string]interface{})
	params := mux.Vars(request)
	companyID, ok := params["company_id"]
	if !ok {
		authorization.WriteError(http.StatusInternalServerError, "Decode Error", response, errors.New("unable to get companyid from claims"))
		return
	}

	body, err := ioutil.ReadAll(request.Body)
	if err != nil {
		panic(err.Error())
	}
	err = json.Unmarshal(body, &keyVal)
	if err != nil {
		authorization.WriteError(http.StatusBadRequest, "BAD REQUEST", response, err)
		return
	}

	userRole2 := fmt.Sprintf("%v", keyVal["role"])

	log.Println("&&&&&&&&&& company ID :", companyID)
	log.Println("&&&&&&&&&& company ID :", userRole2)
	user.ID = primitive.NewObjectID()
	user.FirstName = fmt.Sprintf("%v", keyVal["firstname"])
	user.LastName = fmt.Sprintf("%v", keyVal["lastname"])
	user.Password = fmt.Sprintf("%v", keyVal["password"])
	user.Email = fmt.Sprintf("%v", keyVal["email"])

	//setting default value for  role
	if userRole2 == "superadmin" || (userRole2 != "admin" && userRole2 != "member" && userRole2 != "anonymous") {
		authorization.WriteError(http.StatusBadRequest, "Invalid user role provided", response, errors.New("wrong role"))
		return
	}

	user.Password = utils.GetHash([]byte(user.Password))
	client := db.InitializeDatabase()
	defer client.Disconnect(context.Background())
	collection := client.Database(utils.Database).Collection(utils.UserCollection)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_, err = collection.InsertOne(ctx, user)
	if err != nil {
		response.WriteHeader(http.StatusConflict)
		response.Write([]byte(`{"message":"` + fmt.Sprintf("User with userid: %d OR email : %s is already present in database, please provide different userid or email.", user.ID, user.Email) + `"}`))
		return
	}
	// user role insertion in roles collection
	userRole.UserEmail = user.Email
	userRole.CompanyId = companyID
	userRole.Role = userRole2

	collection = client.Database(utils.Database).Collection(utils.CompanyRolesCollection)
	_, err = collection.InsertOne(ctx, userRole)
	if err != nil {
		response.WriteHeader(http.StatusInternalServerError)
		response.Write([]byte(`{"message":"` + err.Error() + `"}`))
		return
	}

	response.WriteHeader(http.StatusAccepted)
	response.Write([]byte(`{"message":"` + fmt.Sprintf("User with email: %s is added to company having id: %s with role: %s", user.Email, companyID, userRole2) + `"}`))
}

// // Inserting articlerole for newly created user for all article present in company
// func insertAllArticleRoleForNewUser(userID int, companyID int, role string) {
// 	client := db.InitializeDatabase()
// 	defer client.Disconnect(context.Background())
// 	database := client.Database("SPEC-CENTER")
// 	companyRoleCollection := database.Collection("article")
// 	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
// 	defer cancel()
// 	filter := primitive.M{"companyid": companyID}

// 	cursor, err := companyRoleCollection.Find(ctx, filter)
// 	if err != nil {
// 		log.Println(err)
// 	}
// 	defer cursor.Close(ctx)

// 	articleRoleCollection := database.Collection("articlerole")
// 	var articleRole model.ArticleRole
// 	var article model.Article

// 	for cursor.Next(ctx) {
// 		err := cursor.Decode(&article)
// 		if err != nil {
// 			log.Println(err)
// 		}
// 		articleRole.ArticleId = article.ArticleID
// 		articleRole.CompanyId = companyID
// 		articleRole.Role = role
// 		articleRole.UserId = userID

// 		_, err = articleRoleCollection.InsertOne(ctx, articleRole)
// 		if err != nil {
// 			log.Printf("Failed to add article role for article id : %d, user id : %d, error : %w", article.ArticleID, userID, err)
// 		}
// 		log.Printf("ArticleRole for  article id : %d, for user id : %d , for company id : %d added successfully", article.ArticleID, userID, companyID)
// 	}
// }
