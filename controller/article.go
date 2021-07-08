package controller

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/atul-wankhade/Spec-Center/authorization"
	"github.com/atul-wankhade/Spec-Center/db"
	"github.com/atul-wankhade/Spec-Center/model"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/dgrijalva/jwt-go"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func CreateArticleHandler(w http.ResponseWriter, r *http.Request, claims jwt.MapClaims) {
	w.Header().Set("Content-Type", "application/json")

	var article model.Article
	err := json.NewDecoder(r.Body).Decode(&article)
	if err != nil {
		authorization.WriteError(http.StatusBadRequest, "BAD REQUEST BODY", w, err)
		return
	}

	userCompanyID, ok := claims["companyid"].(float64)
	if !ok {
		authorization.WriteError(http.StatusInternalServerError, "decode error", w, errors.New("unable to get companyid from claims"))
		return
	}

	if int(userCompanyID) != article.ComapanyID {
		authorization.WriteError(http.StatusBadRequest, "wrong company id", w, errors.New("wrong company id"))
		return
	}

	client := db.InitializeDatabase()
	defer client.Disconnect(context.Background())
	collection := client.Database("SPEC-CENTER").Collection("article")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	result, err := collection.InsertOne(ctx, article)
	if err != nil {
		w.WriteHeader(http.StatusConflict)
		w.Write([]byte(`{"message":"` + fmt.Sprintf("Article with articleid: %d is already present in database, please provide different articleid", article.ArticleID ) + `"}`))
		return
	}

	go insertRolesForNewArticle(article.ArticleID, article.ComapanyID)

	w.WriteHeader(http.StatusAccepted)
	json.NewEncoder(w).Encode(result)
}

func DeleteArticleHandler(response http.ResponseWriter, request *http.Request, claims jwt.MapClaims) {
	response.Header().Set("Content-Type", "application/json")
	dummyarticleID := request.URL.Query().Get("articleid")
	articleID, err := strconv.Atoi(dummyarticleID)
	if err != nil {
		authorization.WriteError(http.StatusInternalServerError, "String conversion error", response, errors.New("unable to convert articleid into int value"))
		return
	}

	companyID, ok := claims["companyid"].(float64)
	if !ok {
		authorization.WriteError(http.StatusInternalServerError, "Decode Error", response, errors.New("unable to get companyid from claims"))
		return
	}
	userID, ok := claims["userid"].(float64)
	if !ok {
		authorization.WriteError(http.StatusInternalServerError, "Decode Error", response, errors.New("unable to get userid from claims"))
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	//first finding out role on particular article from articlerole collection, using getUserArticleRole function

	fmt.Println(userID, companyID, articleID)

	articleRole, err  := getUserArticleRole(int(userID), int(companyID), articleID)
	if err != nil {
		authorization.WriteError(http.StatusBadRequest, "BAD REQUEST, please check request body and its value.", response, err)
		return
	}

	// checking role on article, if its other than admin, superadmin then user is unauthorized to delete article

	if articleRole != "admin" && articleRole != "superadmin" {
		authorization.WriteError(http.StatusUnauthorized, "UNAUTHORIZED", response, errors.New("unauthorized"))
		return
	}

	client := db.InitializeDatabase()
	defer client.Disconnect(context.Background())
	// deleting the article with given id
	collection := client.Database("SPEC-CENTER").Collection("article")
	_, err = collection.DeleteOne(ctx, primitive.M{"articleid": articleID})
	if err != nil {
		authorization.WriteError(http.StatusInternalServerError, "Delete Error", response, errors.New("error while deleting article"))
		return
	}

	// As article is deleted, deleting role on that articles from articlerole collection
	collection1 := client.Database("SPEC-CENTER").Collection("articlerole")
	_, err = collection1.DeleteMany(ctx, primitive.M{"articleid": articleID})
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
	articleID := article.ArticleID

	companyID, ok := claims["companyid"].(float64)
	if !ok {
		authorization.WriteError(http.StatusInternalServerError, "Decode Error", response, errors.New("unable to get companyid from claims"))
		return
	}

	if article.ComapanyID != int(companyID) {
		authorization.WriteError(http.StatusBadRequest, "Please provide correct company id", response, errors.New("wrong companyid"))
		return
	}
	userID, ok := claims["userid"].(float64)
	if !ok {
		authorization.WriteError(http.StatusInternalServerError, "Decode Error", response, errors.New("unable to get userid from claims"))
		return
	}

	//first finding out role on particular article from articlerole collection, using getUserArticleRole function
	articleRole, err := getUserArticleRole(int(userID), int(companyID), articleID)
	if err != nil {
		authorization.WriteError(http.StatusBadRequest, "BAD REQUEST, please check articleid and request body.", response, err)
		return
	}

	// checking role on article, if its other than admin, superadmin then user is unauthorized to delete article

	if articleRole != "admin" && articleRole != "superadmin" {
		authorization.WriteError(http.StatusUnauthorized, "UNAUTHORIZED", response, errors.New("unauthorized"))
		return
	}

	// Updating the article with given id
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)

	defer cancel()

	client := db.InitializeDatabase()
	defer client.Disconnect(context.Background())

	collection := client.Database("SPEC-CENTER").Collection("article")
	filter := primitive.M{"articleid": articleID, "companyid": int(companyID)}
	opts := options.Update().SetUpsert(true)
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
	flaotCompanyID, ok := claims["companyid"].(float64)
	if !ok {
		authorization.WriteError(http.StatusInternalServerError, "decode error", response, errors.New("unable to get companyid from claims"))
		return
	}
	// changing company id to int value
	companyID := int(flaotCompanyID)
	var articles []model.Article

	client := db.InitializeDatabase()
	defer client.Disconnect(context.Background())
	collection := client.Database("SPEC-CENTER").Collection("article")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cursor, err := collection.Find(ctx, bson.M{"companyid": companyID})
	if err != nil {
		response.WriteHeader(http.StatusInternalServerError)
		response.Write([]byte(`{"message":"` + "Please provide Details. " + err.Error() + `"}`))
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

// Inserting articlerole for newly created article for all user present in company with there role
func insertRolesForNewArticle(articleID, companyID int) {
	client := db.InitializeDatabase()
	defer client.Disconnect(context.Background())
	database := client.Database("SPEC-CENTER")

	companyRoleCollection := database.Collection("role")
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
	var userRole model.Roles

	for cursor.Next(ctx) {
		err := cursor.Decode(&userRole)
		if err != nil {
			log.Println(err)
		}
		articleRole.ArticleId = articleID
		articleRole.CompanyId = companyID
		articleRole.Role = userRole.Role
		articleRole.UserId = userRole.UserId

		_, err = articleRoleCollection.InsertOne(ctx, articleRole)
		if err != nil {
			log.Printf("Failed to add article role for article id : %d, user id : %d, error : %w", articleID, userRole.UserId, err)
		}
		log.Printf("Role on new article with article id : %d, for user id : %d , for company id : %d added successfully", articleID, userRole.UserId, companyID)
	}
}
