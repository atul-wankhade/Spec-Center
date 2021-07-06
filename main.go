package main

import (
	"Spec-Center/authorization"
	"Spec-Center/model"
	"errors"
	"fmt"
	"io/ioutil"
	"strconv"

	"go.mongodb.org/mongo-driver/bson/primitive"

	// "casbin/casbin-http-role-example/authorization"
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/dgrijalva/jwt-go"

	"github.com/casbin/casbin"
	// "github.com/dgrijalva/jwt-go"
	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"golang.org/x/crypto/bcrypt"
)

var SECRET_KEY = []byte("gosecretkey")
var Client *mongo.Client

func getHash(pwd []byte) string {
	hash, err := bcrypt.GenerateFromPassword(pwd, bcrypt.MinCost)
	if err != nil {
		log.Println(err)
	}
	return string(hash)
}

func GenerateJWT(userID int, companyID int, userRole string) (string, error) {
	token := jwt.New(jwt.SigningMethodHS256)
	claims := token.Claims.(jwt.MapClaims)
	claims["authorized"] = true
	claims["userid"] = userID
	claims["companyid"] = companyID
	claims["userrole"] = userRole
	claims["exp"] = time.Now().Add(time.Minute * 30)
	tokenString, err := token.SignedString(SECRET_KEY)
	if err != nil {
		log.Println("Error in JWT token generation")
		return "", err
	}
	return tokenString, nil
}

// to get role of user on particular article
func getUserArticleRole(userID int, companyID int, articleID int) string {
	var articleRole model.ArticleRole
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	collection := Client.Database("SPEC-CENTER").Collection("articlerole")
	filter := primitive.M{"articleid": articleID, "userid": int(userID), "companyid": int(companyID)}
	err := collection.FindOne(ctx, filter).Decode(&articleRole)
	if err != nil {
		log.Fatal(err)
	}
	return articleRole.Role
}

// for changing user role on all article when company role for user is changed.
func updateUserArticleRoles(userID, companyID int, updatedRole string) {
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	collection := Client.Database("SPEC-CENTER").Collection("articlerole")
	filter := primitive.M{"userid": userID, "companyid": companyID}
	opts := options.Update().SetUpsert(true)
	update := bson.D{{"$set", bson.D{{"role", updatedRole}}}}

	_, err := collection.UpdateMany(ctx, filter, update, opts)
	if err != nil {
		log.Fatal(err)
	}
}

func main() {
	authorization.SECRET_KEY = SECRET_KEY
	authorization.Client = Client
	log.Print("Starting the application...")

	// setup casbin auth rules
	authEnforcer, err := casbin.NewEnforcerSafe("./auth_model.conf", "./policy.csv")
	if err != nil {
		log.Fatal(err)
	}

	router := mux.NewRouter()
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	Client, _ = mongo.Connect(ctx, options.Client().ApplyURI("mongodb://localhost:27017"))

	//LOGIN & USER ADD
	router.HandleFunc("/login/{companyid}", LoginHandler).Methods("POST")
	router.Handle("/adduser", authorization.IsAuthorized(authEnforcer, AddUser)).Methods("POST")

	// ARTICLE
	router.Handle("/all_articles", authorization.IsAuthorized(authEnforcer, GetArticlesHandler)).Methods("GET")
	router.Handle("/article", authorization.IsAuthorized(authEnforcer, DeleteArticleHandler)).Methods("DELETE")
	router.Handle("/article", authorization.IsAuthorized(authEnforcer, UpdateArticleHandler)).Methods("PUT")
	router.Handle("/article", authorization.IsAuthorized(authEnforcer, CreateArticleHandler)).Methods("POST")

	//ROLE CHANGE :- only superadmin can change role of other user.
	router.Handle("/articlerole/{articleid}", authorization.IsAuthorized(authEnforcer, UpdateArticleRoleHandler)).Methods("PUT")
	router.Handle("/role", authorization.IsAuthorized(authEnforcer, UpdateCompanyRoleHandler)).Methods("PUT")

	// router.Use(authorization.Authorizer(authEnforcer))
	log.Print("Server started on localhost:8040")
	log.Fatal(http.ListenAndServe(":8040", router))
}

func GetArticlesHandler(response http.ResponseWriter, request *http.Request, claims jwt.MapClaims) {
	response.Header().Set("Content-Type", "application/json")
	companyID := int(claims["companyid"].(float64))

	var articles []model.Article

	collection := Client.Database("SPEC-CENTER").Collection("article")
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)

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

func LoginHandler(response http.ResponseWriter, request *http.Request) {
	response.Header().Set("Content-Type", "application/json")

	var user model.User
	var dbUser model.User
	var role model.Roles

	params := mux.Vars(request)
	companyID, _ := strconv.Atoi((params["companyid"]))

	json.NewDecoder(request.Body).Decode(&user)

	collection := Client.Database("SPEC-CENTER").Collection("user")
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)

	err := collection.FindOne(ctx, bson.M{"email": user.Email}).Decode(&dbUser)
	if err != nil {
		response.WriteHeader(http.StatusInternalServerError)
		response.Write([]byte(`{"message":"` + "Please provide correct Details. " + err.Error() + `"}`))
		return
	}

	collection = Client.Database("SPEC-CENTER").Collection("role")
	err = collection.FindOne(ctx, primitive.M{"userid": dbUser.ID, "companyid": companyID}).Decode(&role)
	if err != nil {
		response.WriteHeader(http.StatusInternalServerError)
		response.Write([]byte(`{"message":"` + "Please provide Details. " + err.Error() + `"}`))
		return
	}
	userRole := role.Role
	dbUserId := dbUser.ID
	userPass := []byte(user.Password)
	dbPass := []byte(dbUser.Password)

	// checking companyid is correct or not
	if companyID != role.CompanyId {
		authorization.WriteError(http.StatusBadRequest, "wrong company id", response, errors.New("wrong company"))
		return
	}

	passErr := bcrypt.CompareHashAndPassword(dbPass, userPass)
	if passErr != nil {
		log.Println(passErr)
		response.Write([]byte(`{"response":"Wrong Password!"}`))
		return
	}
	jwtToken, err := GenerateJWT(dbUserId, companyID, userRole)
	if err != nil {
		response.WriteHeader(http.StatusInternalServerError)
		response.Write([]byte(`{"message":"` + err.Error() + `"}`))
		return
	}
	response.Write([]byte(`{"token":"` + jwtToken + `"}`))
}

func DeleteArticleHandler(response http.ResponseWriter, request *http.Request, claims jwt.MapClaims) {
	response.Header().Set("Content-Type", "application/json")
	dummyarticleID := request.URL.Query().Get("articleid")
	articleID, err := strconv.Atoi(dummyarticleID)
	if err != nil {
		authorization.WriteError(http.StatusInternalServerError, "String conversion error", response, errors.New("unable to convert articleid into int value"))
		return
	}
	fmt.Println("!!!!!!!!!!!!!!!!!!!!!", articleID)

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
	ctx, _ := context.WithTimeout(context.Background(), 30*time.Second)

	//first finding out role on particular article from articlerole collection, using getUserArticleRole function

	fmt.Println(userID, companyID, articleID)

	articleRole := getUserArticleRole(int(userID), int(companyID), articleID)

	//for logs
	fmt.Println("!!!!!!!!!!!", articleRole)

	// checking role on article, if its other than admin, superadmin then user is unauthorized to delete article

	if articleRole != "admin" && articleRole != "superadmin" {
		authorization.WriteError(http.StatusUnauthorized, "UNAUTHORIZED", response, errors.New("unauthorized"))
		return
	}

	// deleting the article with given id
	collection := Client.Database("SPEC-CENTER").Collection("article")
	_, err = collection.DeleteOne(ctx, primitive.M{"articleid": articleID})
	if err != nil {
		authorization.WriteError(http.StatusInternalServerError, "Delete Error", response, errors.New("error while deleting article"))
		return
	}

	// As article is deleted, deleting role on that articles from articlerole collection
	collection1 := Client.Database("SPEC-CENTER").Collection("articlerole")
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

	fmt.Println("!!!!!!!!!!!!!!!!!!!!!", articleID)

	companyID, ok := claims["companyid"].(float64)
	if !ok {
		authorization.WriteError(http.StatusInternalServerError, "Decode Error", response, errors.New("unable to get companyid from claims"))
		return
	}

	if article.ComapanyID != int(companyID) {
		authorization.WriteError(http.StatusBadRequest, "Please provide correct company id ", response, errors.New("wrong companyid"))
		return
	}
	userID, ok := claims["userid"].(float64)
	if !ok {
		authorization.WriteError(http.StatusInternalServerError, "Decode Error", response, errors.New("unable to get userid from claims"))
		return
	}

	//first finding out role on particular article from articlerole collection, using getUserArticleRole function

	fmt.Println(userID, companyID, articleID)

	articleRole := getUserArticleRole(int(userID), int(companyID), articleID)

	//for logs
	fmt.Println("!!!!!!!!!!!", articleRole)

	// checking role on article, if its other than admin, superadmin then user is unauthorized to delete article

	if articleRole != "admin" && articleRole != "superadmin" {
		authorization.WriteError(http.StatusUnauthorized, "UNAUTHORIZED", response, errors.New("unauthorized"))
		return
	}

	// Updating the article with given id
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	collection := Client.Database("SPEC-CENTER").Collection("article")
	filter := primitive.M{"articleid": articleID, "companyid": int(companyID)}
	opts := options.Update().SetUpsert(true)
	update := bson.D{{"$set", bson.D{{"body", article.Body}}}}

	_, err = collection.UpdateOne(ctx, filter, update, opts)
	if err != nil {
		authorization.WriteError(http.StatusInternalServerError, "update error", response, errors.New("update error"))
		return
	}
	response.WriteHeader(http.StatusOK)
	response.Write([]byte(`{"message":"` + fmt.Sprintf("Article with id: %d is successfully updated!", articleID) + `"}`))
}

func UpdateArticleRoleHandler(w http.ResponseWriter, r *http.Request, claims jwt.MapClaims) {
	w.Header().Set("Content-Type", "application/json")
	var articleRole model.ArticleRole

	err := json.NewDecoder(r.Body).Decode(&articleRole)
	if err != nil {
		authorization.WriteError(http.StatusBadRequest, "DECODE ERROR", w, err)
		return
	}

	companyId := int(claims["companyid"].(float64))

	if articleRole.CompanyId != companyId {
		authorization.WriteError(http.StatusBadRequest, "Please provide correct company id ", w, err)
		return
	}

	if articleRole.Role != "admin" && articleRole.Role != "member" && articleRole.Role != "anonymous" {
		authorization.WriteError(http.StatusBadRequest, "invalid role provided", w, errors.New("invalid role"))
		return
	}
	superadminId := int(claims["userid"].(float64))

	if superadminId == articleRole.UserId {
		authorization.WriteError(http.StatusBadRequest, "Cannot change role on article for superadmin", w, errors.New("cannot update own role"))
		return
	}

	var role model.Roles

	fmt.Println("#####################")
	collection := Client.Database("SPEC-CENTER").Collection("role")
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	err = collection.FindOne(ctx, primitive.M{"userid": articleRole.UserId, "companyid": articleRole.CompanyId}).Decode(&role)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"message":"` + "Please provide correct Details. " + err.Error() + `"}`))
		return
	}
	fmt.Println("!!!!!!!!!!!!!!!!!!!!!", role)
	if role.Role == "anonymous" {
		authorization.WriteError(http.StatusBadRequest, "BAD REQUEST:- anonymous user can't be given article access", w, errors.New("BAD REQUEST"))
		return
	}

	fmt.Println("%%%%%%%%%%%%%%")
	collection1 := Client.Database("SPEC-CENTER").Collection("articlerole")
	opts := options.Update().SetUpsert(true)
	filter := primitive.M{"userid": articleRole.UserId, "companyid": articleRole.CompanyId, "articleid": articleRole.ArticleId}
	update := bson.D{{"$set", bson.D{{"role", articleRole.Role}}}}

	_, err = collection1.UpdateOne(ctx, filter, update, opts)
	if err != nil {
		authorization.WriteError(http.StatusInternalServerError, "update error", w, errors.New("update error"))
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"message":"` + fmt.Sprintf("Role for userid:%d for articleid: %d is changed to: %s", articleRole.UserId, articleRole.ArticleId, articleRole.Role) + `"}`))
}

func UpdateCompanyRoleHandler(w http.ResponseWriter, r *http.Request, claims jwt.MapClaims) {
	var role model.Roles
	err := json.NewDecoder(r.Body).Decode(&role)
	if err != nil {
		authorization.WriteError(http.StatusBadRequest, "DECODE ERROR", w, err)
		return
	}
	companyId := int(claims["companyid"].(float64))

	if role.CompanyId != companyId {
		authorization.WriteError(http.StatusBadRequest, "Please provide correct company id ", w, errors.New("wrong companyid"))
		return
	}

	superadminId := int(claims["userid"].(float64))

	if superadminId == role.UserId {
		authorization.WriteError(http.StatusBadRequest, "Cannot change own role", w, errors.New("cannot update own role"))
		return
	}

	if role.Role != "admin" && role.Role != "member" && role.Role != "anonymous" {
		authorization.WriteError(http.StatusBadRequest, "invalid role provided", w, errors.New("invalid role"))
		return
	}

	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	collection := Client.Database("SPEC-CENTER").Collection("role")
	filter := primitive.M{"userid": role.UserId, "companyid": role.CompanyId}
	opts := options.Update().SetUpsert(true)
	update := bson.D{{"$set", bson.D{{"role", role.Role}}}}

	_, err = collection.UpdateOne(ctx, filter, update, opts)
	if err != nil {
		authorization.WriteError(http.StatusInternalServerError, "update error", w, errors.New("update error"))
		return
	}

	go updateUserArticleRoles(role.UserId, role.CompanyId, role.Role)

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"message":"` + fmt.Sprintf("Role for userid:%d  is changed to: %s", role.UserId, role.Role) + `"}`))
}

func CreateArticleHandler(w http.ResponseWriter, r *http.Request, claims jwt.MapClaims) {
	w.Header().Set("Content-Type", "application/json")

	var article model.Article
	err := json.NewDecoder(r.Body).Decode(&article)
	if err != nil {
		authorization.WriteError(http.StatusBadRequest, "decode error", w, err)
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

	collection := Client.Database("SPEC-CENTER").Collection("article")
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	result, err := collection.InsertOne(ctx, article)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"message":"` + err.Error() + `"}`))
		return
	}

	go insertRolesForNewArticle(article.ArticleID, article.ComapanyID)

	w.WriteHeader(http.StatusAccepted)
	json.NewEncoder(w).Encode(result)
}
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
		authorization.WriteError(http.StatusInternalServerError, "Decode Error", response, errors.New("Wrong userid"))
		return
	}

	user.ID =  int(userId)
	user.FirstName = fmt.Sprintf("%v", keyVal["firstname"])
	user.LastName = fmt.Sprintf("%v", keyVal["lastname"])
	user.Password = fmt.Sprintf("%v", keyVal["password"])
	user.Email = fmt.Sprintf("%v", keyVal["email"])

	// for logs
	fmt.Println("!@!@!@!@",keyVal, user)

	//setting default value for  role
	if userRole != "admin" && userRole != "member" && userRole != "anonymous" {
		userRole = "anonymous"
	}

	user.Password = getHash([]byte(user.Password))
	collection := Client.Database("SPEC-CENTER").Collection("user")
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	_, err = collection.InsertOne(ctx, user)
	if err != nil {
		response.WriteHeader(http.StatusInternalServerError)
		response.Write([]byte(`{"message":"` + err.Error() + `"}`))
		return
	}
	// user role insertion in roles collection
	role.UserId = user.ID
	role.CompanyId = companyID
	role.Role = userRole

	collection = Client.Database("SPEC-CENTER").Collection("role")
	_, err = collection.InsertOne(ctx, role)
	if err != nil {
		response.WriteHeader(http.StatusInternalServerError)
		response.Write([]byte(`{"message":"` + err.Error() + `"}`))
		return
	}

	go insertAllArticleRoleForNewUser(user.ID, companyID, userRole)

	response.WriteHeader(http.StatusAccepted)
	response.Write([]byte(`{"message":"` + fmt.Sprintf("User with userid: %d is added to company having id: %d with role: %s", int(userId),companyID,userRole) + `"}`))
}

// Inserting articlerole for newly created user for all article present in company
func insertAllArticleRoleForNewUser(userID int, companyID int, role string) {
	database := Client.Database("SPEC-CENTER")
	companyRoleCollection := database.Collection("article")
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	filter := primitive.M{"companyid": companyID}

	cursor, err := companyRoleCollection.Find(ctx, filter)
	if err != nil {
		log.Fatal(err)
	}
	defer cursor.Close(ctx)

	articleRoleCollection := database.Collection("articlerole")
	var articleRole model.ArticleRole
	var article model.Article

	for cursor.Next(ctx) {
		err := cursor.Decode(&article)
		if err != nil {
			log.Fatal(err)
		}
		articleRole.ArticleId = article.ArticleID
		articleRole.CompanyId = companyID
		articleRole.Role = role
		articleRole.UserId =  userID

		_, err = articleRoleCollection.InsertOne(ctx, articleRole)
		if err != nil {
			log.Fatalf("Failed to add article role for article id : %d, user id : %d, error : %w", article.ArticleID, userID, err)
		}
		log.Printf("ArticleRole for  article id : %d, for user id : %d , for company id : %d added successfully", article.ArticleID, userID, companyID)
	}
}

// Inserting articlerole for newly created article for all user present in company with there role
func insertRolesForNewArticle(articleID, companyID int) {
	database := Client.Database("SPEC-CENTER")

	companyRoleCollection := database.Collection("role")
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	filter := primitive.M{"companyid": companyID}

	cursor, err := companyRoleCollection.Find(ctx, filter)
	if err != nil {
		log.Fatal(err)
	}
	defer cursor.Close(ctx)

	articleRoleCollection := database.Collection("articlerole")
	var articleRole model.ArticleRole
	var userRole model.Roles

	for cursor.Next(ctx) {
		err := cursor.Decode(&userRole)
		if err != nil {
			log.Fatal(err)
		}
		articleRole.ArticleId = articleID
		articleRole.CompanyId = companyID
		articleRole.Role = userRole.Role
		articleRole.UserId = userRole.UserId

		_, err = articleRoleCollection.InsertOne(ctx, articleRole)
		if err != nil {
			log.Fatalf("Failed to add article role for article id : %d, user id : %d, error : %w", articleID, userRole.UserId, err)
		}
		log.Printf("Role on new article with article id : %d, for user id : %d , for company id : %d added successfully", articleID, userRole.UserId, companyID)
	}
}
