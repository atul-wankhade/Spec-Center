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

	// validation check on body field
	if article.Body == "" {
		authorization.WriteError(http.StatusBadRequest, "Invalid payload or nil body parameter", w, errors.New("invalid request body"))
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

	emailInterface, ok := claims["user_email"]
	if !ok {
		authorization.WriteError(http.StatusInternalServerError, "Decode Error", response, errors.New("unable to get email from claims"))
		return
	}
	email := fmt.Sprintf("%v", emailInterface)

	roleInterface, ok := claims["role"]
	if !ok {
		authorization.WriteError(http.StatusInternalServerError, "Decode Error", response, errors.New("unable to get role from claims"))
		return
	}
	role := fmt.Sprintf("%v", roleInterface)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	authorized := authorization.IsAuthorizedForArticle(companyID, email, role, request.Method, articleID)
	if !authorized {
		authorization.WriteError(http.StatusUnauthorized, "UNAUTHORIZED", response, errors.New("unauthorized"))
		return
	}

	client := db.InitializeDatabase()
	defer client.Disconnect(context.Background())
	// deleting the article with given id
	collection := client.Database(utils.Database).Collection(utils.ArticleCollection)
	result, err := collection.DeleteOne(ctx, primitive.M{"_id": articleID})
	if err != nil {
		authorization.WriteError(http.StatusInternalServerError, "Delete Error", response, errors.New("error while deleting article"))
		return
	}
	if result.DeletedCount == 0 {
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
	response.Write([]byte(`{"message":"` + fmt.Sprintf("Article with id: %s is successfully deleted!", articleID) + `"}`))
}

func UpdateArticleHandler(response http.ResponseWriter, request *http.Request, claims jwt.MapClaims) {
	response.Header().Set("Content-Type", "application/json")
	var article model.Article

	err := json.NewDecoder(request.Body).Decode(&article)
	if err != nil {
		authorization.WriteError(http.StatusBadRequest, "DECODE ERROR", response, err)
		return
	}

	if article.Body == "" {
		authorization.WriteError(http.StatusBadRequest, "Invalid payload or nil body parameter", response, errors.New("invalid request body"))
		return
	}

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

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	client := db.InitializeDatabase()
	defer client.Disconnect(context.Background())

	articleCollection := client.Database(utils.Database).Collection(utils.ArticleCollection)

	filter := bson.M{"company_id": companyID, "_id": articleID}
	result := articleCollection.FindOne(ctx, filter)
	if result.Err() != nil {
		authorization.WriteError(http.StatusBadRequest, "wrong article_id, no article present with this id in database", response, result.Err())
		return
	}

	userID := fmt.Sprintf("%v", claims["user_id"])

	roleInterface, ok := claims["role"]
	if !ok {
		authorization.WriteError(http.StatusInternalServerError, "Decode Error", response, errors.New("unable to get email from claims"))
		return
	}
	role := fmt.Sprintf("%v", roleInterface)

	authorized := authorization.IsAuthorizedForArticle(companyID, userID, role, request.Method, articleID)
	if !authorized {
		authorization.WriteError(http.StatusUnauthorized, "UNAUTHORIZED", response, errors.New("unauthorized"))
		return
	}

	// Updating the article with given id
	articlecollection := client.Database(utils.Database).Collection(utils.ArticleCollection)
	filter2 := primitive.M{"_id": articleID, "company_id": companyID}
	opts := options.Update().SetUpsert(false)
	update := primitive.M{"$set": primitive.M{"body": article.Body}}
	_, err = articlecollection.UpdateOne(ctx, filter2, update, opts)
	if err != nil {
		authorization.WriteError(http.StatusBadRequest, "update error", response, err)
		return
	}
	response.WriteHeader(http.StatusOK)
	response.Write([]byte(`{"message":"` + fmt.Sprintf("Article with id: %s is successfully updated!", articleID) + `"}`))
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
