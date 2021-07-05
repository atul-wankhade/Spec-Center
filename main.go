package main

import (
	"Spec-Center/authorization"
	"Spec-Center/model"
	"errors"
	"fmt"
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


// var Key *ecdsa.PrivateKey

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

	//setup routes
	router.HandleFunc("/login/{companyid}", LoginHandler).Methods("POST")
	router.Handle("/adduser/{role}", authorization.IsAuthorized(authEnforcer,AddUser)).Methods("POST")
	router.Handle("/all_articles", authorization.IsAuthorized(authEnforcer, GetArticlesHandler)).Methods("GET")


	router.Handle("/{company}/article/{articleid}", authorization.IsAuthorized(authEnforcer, GetArticlesHandler)).Methods("GET")
	router.Handle("/{company}/article/{articleid}", authorization.IsAuthorized(authEnforcer, DeleteArticleHandler)).Methods("DELETE")
	router.Handle("/{company}/articlerole/{articleid}", authorization.IsAuthorized(authEnforcer, UpdateArticleRoleHandler)).Methods("PATCH")
	// router.HandleFunc("/article/{company}", CreateArticleHandler).Methods("POST")
	// router.HandleFunc("/article/{company}", DeleteArticleHandler).Methods("DELETE")
	// router.Use(authorization.Authorizer(authEnforcer))

	log.Print("Server started on localhost:8000")
	log.Fatal(http.ListenAndServe(":8000", router))
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
		response.Write([]byte(`{"message":"` + "Please provide Details. "+ err.Error() + `"}`))
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

	//checking companyid is correct or not
	if companyID != 1 && companyID != 2 && companyID != 3 {
		response.Write([]byte(`{"response":"Wrong Company Id!"}`))
		return
	}


	json.NewDecoder(request.Body).Decode(&user)

	collection := Client.Database("SPEC-CENTER").Collection("user")
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)

	err := collection.FindOne(ctx, bson.M{"email": user.Email}).Decode(&dbUser)
	if err != nil {
		response.WriteHeader(http.StatusInternalServerError)
		response.Write([]byte(`{"message":"` + "Please provide correct Details. "+ err.Error() + `"}`))
		return
	}

	collection = Client.Database("SPEC-CENTER").Collection("role")
	err = collection.FindOne(ctx, primitive.M{"userid": dbUser.ID, "companyid": companyID}).Decode(&role)
	if err != nil {
		response.WriteHeader(http.StatusInternalServerError)
		response.Write([]byte(`{"message":"` + "Please provide Details. "+ err.Error() + `"}` ))
		return
	}
	userRole := role.Role
	dbUserId := dbUser.ID
	userPass := []byte(user.Password)
	dbPass := []byte(dbUser.Password)
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

func AddUser(response http.ResponseWriter, request *http.Request, claims jwt.MapClaims) {
	var user model.User
	var role model.Roles
	companyID := int(claims["companyid"].(float64))
	response.Header().Set("Content-Type", "application/json")

	params := mux.Vars(request)
	userRole := params["role"]

	//setting default value for  role
	if userRole !="admin" && userRole !="member" && userRole != "anonymous" {
		userRole = "anonymous"
	}

	json.NewDecoder(request.Body).Decode(&user)

	user.Password = getHash([]byte(user.Password))
	collection := Client.Database("SPEC-CENTER").Collection("user")
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	result, err := collection.InsertOne(ctx, user)
	if err != nil {
		response.WriteHeader(http.StatusInternalServerError)
		response.Write([]byte(`{"message":"` + err.Error() + `"}`))
		return
	}

	// role collection insertion
	role.UserId = user.ID
	role.CompanyId = companyID
	role.Role = userRole

	collection = Client.Database("SPEC-CENTER").Collection("role")
	_ , err = collection.InsertOne(ctx, role)
	if err != nil {
		response.WriteHeader(http.StatusInternalServerError)
		response.Write([]byte(`{"message":"` + err.Error() + `"}`))
		return
	}
	// for logs
	fmt.Println("IMP", companyID, user,role)
	json.NewEncoder(response).Encode(result)
}

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

func DeleteArticleHandler(w http.ResponseWriter, r *http.Request, claims jwt.MapClaims) {
	vars := mux.Vars(r)
	articleID := vars["articleid"]
	companyID, ok := claims["company_id"].(float64)
	if !ok {
		authorization.WriteError(http.StatusInternalServerError, "Decode Error", w, errors.New("unable to get companyid from claims"))
		return
	}
	userID, ok := claims["userid"].(float64)
	if !ok {
		authorization.WriteError(http.StatusInternalServerError, "Decode Error", w, errors.New("unable to get userid from claims"))
		return
	}
	articleRole := model.ArticleRole{}
	ctx, _ := context.WithTimeout(context.Background(), 30*time.Second)
	database := Client.Database("SPEC-CENTER")
	collection := database.Collection("articlerole")
	// filter := []primitive.M{{"articleid": articleID}, {"userid": userID}, {"companyid": companyID}}
	filter := primitive.M{"articleid": articleID, "userid": userID, "companyid": companyID}
	fmt.Println(filter)
	err := collection.FindOne(context.Background(), filter).Decode(&articleRole)
	if err != nil {
		authorization.WriteError(http.StatusInternalServerError, "Decode Error", w, errors.New("error while decoding article role"))
		return
	}

	if articleRole.Role != "admin" || articleRole.Role != "superadmin" {
		authorization.WriteError(http.StatusUnauthorized, "UNAUTHORIZED", w, errors.New("unauthorized"))
		return
	}

	articleColl := database.Collection("article")
	_, err = articleColl.DeleteOne(ctx, primitive.M{"id": articleID})
	if err != nil {
		authorization.WriteError(http.StatusInternalServerError, "Delete Error", w, errors.New("error while deleting article"))
		return
	}

	_, err = collection.DeleteMany(ctx, primitive.M{"id": articleID})
	if err != nil {
		authorization.WriteError(http.StatusInternalServerError, "Delete Error", w, errors.New("error while deleting article roles"))
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"message: successfully  deleted article"}`))
	return
}

