package controller

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/atul-wankhade/Spec-Center/authorization"
	"github.com/atul-wankhade/Spec-Center/db"
	"github.com/atul-wankhade/Spec-Center/model"
	"github.com/atul-wankhade/Spec-Center/utils"
	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
	"gopkg.in/mgo.v2/bson"

	"github.com/dgrijalva/jwt-go"
)

func CreateArticleHandler(w http.ResponseWriter, r *http.Request, claims jwt.MapClaims) {
	w.Header().Set("Content-Type", "application/json")

	var article model.Article
	articleID := primitive.NewObjectID()
	err := json.NewDecoder(r.Body).Decode(&article)
	if err != nil {
		authorization.WriteError(http.StatusBadRequest, "BAD REQUEST BODY", w, err)
		return
	}

	params := mux.Vars(r)
	companyID, ok := params["company_id"]
	if !ok {
		authorization.WriteError(http.StatusInternalServerError, "Decode Error", w, errors.New("unable to get companyid from claims"))
		return
	}

	article.CompanyID = companyID
	article.ID = articleID

	client := db.InitializeDatabase()
	defer client.Disconnect(context.Background())
	collection := client.Database(utils.Database).Collection(utils.ArticleCollection)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_, err = collection.InsertOne(ctx, article)
	if err != nil {
		w.WriteHeader(http.StatusConflict)
		w.Write([]byte(`{"message":"` + fmt.Sprintf("Article with articleid: %s is already present in database, please provide different articleid", article.ID.Hex()) + `"}`))
		return
	}

	w.WriteHeader(http.StatusAccepted)
	w.Write([]byte(`{"message":"` + fmt.Sprintf("Article with article id: %s is added to company having id: %s ", articleID.Hex(), companyID) + `"}`))
}

func DeleteArticleHandler(response http.ResponseWriter, request *http.Request, claims jwt.MapClaims) {
	response.Header().Set("Content-Type", "application/json")
	params := mux.Vars(request)
	companyID, ok := params["company_id"]
	if !ok {
		authorization.WriteError(http.StatusInternalServerError, "Decode Error", response, errors.New("unable to get companyid from request parameters"))
		return
	}

	articleStringID, ok := params["article_id"]
	if !ok {
		authorization.WriteError(http.StatusInternalServerError, "Decode Error", response, errors.New("unable to get article id from request parameters"))
		return
	}
	articleID, err := primitive.ObjectIDFromHex(articleStringID)
	if err != nil {
		authorization.WriteError(http.StatusInternalServerError, "Convert Error", response, errors.New("conversion error"))
		return
	}

	emailInterface, ok := claims["email"]
	if !ok {
		authorization.WriteError(http.StatusInternalServerError, "Decode Error", response, errors.New("unable to get email from claims"))
		return
	}
	email := fmt.Sprintf("%v", emailInterface)

	roleInterface, ok := claims["role"]
	if !ok {
		authorization.WriteError(http.StatusInternalServerError, "Decode Error", response, errors.New("unable to get email from claims"))
		return
	}
	role := fmt.Sprintf("%v", roleInterface)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	//first finding out role on particular article from articlerole collection, using getUserArticleRole function

	fmt.Println(email, companyID, articleID)

	authorized := authorization.IsAuthorizedForArticle(companyID, email, role, request.Method, articleID)
	if !authorized {
		authorization.WriteError(http.StatusUnauthorized, "UNAUTHORIZED", response, errors.New("unauthorized"))
		return
	}

	client := db.InitializeDatabase()
	defer client.Disconnect(context.Background())
	// deleting the article with given id
	collection := client.Database(utils.Database).Collection(utils.ArticleCollection)
	_, err = collection.DeleteOne(ctx, primitive.M{"_id": articleID})
	if err != nil {
		authorization.WriteError(http.StatusInternalServerError, "Delete Error", response, errors.New("error while deleting article"))
		return
	}

	// As article is deleted, deleting role on that articles from articlerole collection
	collection1 := client.Database(utils.Database).Collection(utils.ArticleRoleCollection)
	_, err = collection1.DeleteMany(ctx, primitive.M{"article_id": articleStringID})
	if err != nil {
		authorization.WriteError(http.StatusInternalServerError, "Delete Error", response, errors.New("error while deleting article roles"))
		return
	}
	response.WriteHeader(http.StatusOK)
	response.Write([]byte(`{"message":"` + fmt.Sprintf("Article with id: %d is successfully deleted!", articleID) + `"}`))
}

func UpdateArticleHandler(response http.ResponseWriter, request *http.Request, claims jwt.MapClaims) {
	response.Header().Set("Content-Type", "application/json")
	var article model.Article

	err := json.NewDecoder(request.Body).Decode(&article)
	if err != nil {
		authorization.WriteError(http.StatusBadRequest, "DECODE ERROR", response, err)
		return
	}

	// articleID := primitive.ObjectIDFromHex()

	params := mux.Vars(request)
	companyID, ok := params["company_id"]
	if !ok {
		authorization.WriteError(http.StatusInternalServerError, "Decode Error", response, errors.New("unable to get companyid from parameters"))
		return
	}

	articleID, err := primitive.ObjectIDFromHex(params["article_id"])
	if err != nil {
		authorization.WriteError(http.StatusInternalServerError, "Decode Error", response, errors.New("unable to get article id from parameters"))
		return
	}

	if article.CompanyID != companyID {
		authorization.WriteError(http.StatusBadRequest, "Please provide correct company id", response, errors.New("wrong companyid"))
		return
	}

	if article.ID != articleID {
		authorization.WriteError(http.StatusBadRequest, "Please provide correct article id", response, errors.New("wrong article id"))
		return
	}

	emailInterface, ok := claims["email"]
	if !ok {
		authorization.WriteError(http.StatusInternalServerError, "Decode Error", response, errors.New("unable to get email from claims"))
		return
	}
	email := fmt.Sprintf("%v", emailInterface)

	roleInterface, ok := claims["role"]
	if !ok {
		authorization.WriteError(http.StatusInternalServerError, "Decode Error", response, errors.New("unable to get email from claims"))
		return
	}
	role := fmt.Sprintf("%v", roleInterface)

	authorized := authorization.IsAuthorizedForArticle(companyID, email, role, request.Method, articleID)
	if !authorized {
		authorization.WriteError(http.StatusUnauthorized, "UNAUTHORIZED", response, errors.New("unauthorized"))
		return
	}
	// Updating the article with given id
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)

	defer cancel()

	client := db.InitializeDatabase()
	defer client.Disconnect(context.Background())

	collection := client.Database(utils.Database).Collection(utils.ArticleCollection)
	filter := primitive.M{"article_id": articleID, "company_id": companyID}
	opts := options.Update().SetUpsert(false)
	update := bson.D{{"$set", bson.D{{"body", article.Body}}}}
	_, err = collection.UpdateOne(ctx, filter, update, opts)
	if err != nil {
		authorization.WriteError(http.StatusBadRequest, "update error", response, err)
		return
	}
	response.WriteHeader(http.StatusOK)
	response.Write([]byte(`{"message":"` + fmt.Sprintf("Article with id: %d is successfully updated!", articleID) + `"}`))
}

func GetArticlesHandler(response http.ResponseWriter, request *http.Request, claims jwt.MapClaims) {
	response.Header().Set("Content-Type", "application/json")
	params := mux.Vars(request)
	companyID, ok := params["company_id"]
	if !ok {
		authorization.WriteError(http.StatusInternalServerError, "Decode Error", response, errors.New("unable to get companyid from parameters"))
		return
	}

	var articles []model.Article

	client := db.InitializeDatabase()
	defer client.Disconnect(context.Background())
	collection := client.Database(utils.Database).Collection(utils.ArticleCollection)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cursor, err := collection.Find(ctx, bson.M{"company_id": companyID})
	if err != nil {
		response.WriteHeader(http.StatusInternalServerError)
		response.Write([]byte(`{"message":"` + "Please provide correct Details, check company_id" + err.Error() + `"}`))
		return
	}
	defer cursor.Close(ctx)
	for cursor.Next(ctx) {
		var article model.Article
		cursor.Decode(&article)
		articles = append(articles, article)
	}
	if err := cursor.Err(); err != nil {
		response.WriteHeader(http.StatusInternalServerError)
		response.Write([]byte(`{ "message": "` + err.Error() + `" }`))
		return
	}
	json.NewEncoder(response).Encode(articles)
}

// to get role of user on particular article
func getUserArticleRole(userID int, companyID int, articleID int) (string, error) {
	var articleRole model.ArticleRole
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	client := db.InitializeDatabase()
	defer client.Disconnect(context.Background())
	collection := client.Database("SPEC-CENTER").Collection("articlerole")
	filter := primitive.M{"articleid": articleID, "userid": int(userID), "companyid": int(companyID)}
	err := collection.FindOne(ctx, filter).Decode(&articleRole)
	if err != nil {
		return "", err
	}
	return articleRole.Role, nil
}

// // Inserting articlerole for newly created article for all user present in company with there role
// func insertRolesForNewArticle(articleID, companyID int) {
// 	client := db.InitializeDatabase()
// 	defer client.Disconnect(context.Background())
// 	database := client.Database("SPEC-CENTER")

// 	companyRoleCollection := database.Collection("role")
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
// 	var userRole model.Roles

// 	for cursor.Next(ctx) {
// 		err := cursor.Decode(&userRole)
// 		if err != nil {
// 			log.Println(err)
// 		}
// 		articleRole.ArticleId = articleID
// 		articleRole.CompanyId = companyID
// 		articleRole.Role = userRole.Role
// 		articleRole.UserId = userRole.UserId

// 		_, err = articleRoleCollection.InsertOne(ctx, articleRole)
// 		if err != nil {
// 			log.Printf("Failed to add article role for article id : %d, user id : %d, error : %w", articleID, userRole.UserId, err)
// 		}
// 		log.Printf("Role on new article with article id : %d, for user id : %d , for company id : %d added successfully", articleID, userRole.UserId, companyID)
// 	}
// }

func GetSingleArticleHandler(w http.ResponseWriter, r *http.Request, claims jwt.MapClaims) {
	w.Header().Set("Content-Type", "application/json")
	params := mux.Vars(r)

	articleID, err := primitive.ObjectIDFromHex(params["article_id"])
	if err != nil {
		authorization.WriteError(http.StatusInternalServerError, "Decode Error", w, errors.New("unable to get article id from parameters"))
		return
	}
	client := db.InitializeDatabase()
	defer client.Disconnect(context.Background())
	var article model.Article
	filter := primitive.M{"_id": articleID}
	err = client.Database(utils.Database).Collection(utils.ArticleCollection).FindOne(context.Background(), filter).Decode(&article)
	if err != nil {
		authorization.WriteError(http.StatusInternalServerError, "unable to get article", w, err)
		return
	}
	json.NewEncoder(w).Encode(article)
	w.WriteHeader(http.StatusOK)

}
