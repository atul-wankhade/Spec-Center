package controller

import (
	"context"
	"encoding/json"
	"github.com/atul-wankhade/Spec-Center/db"
	"github.com/atul-wankhade/Spec-Center/model"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"golang.org/x/crypto/bcrypt"
)

var SECRET_KEY = []byte("gosecretkey")

// var Client *mongo.Client

func LoginHandler(response http.ResponseWriter, request *http.Request) {
	response.Header().Set("Content-Type", "application/json")

	var user model.User
	var dbUser model.User
	var role model.Roles

	params := mux.Vars(request)
	companyID, _ := strconv.Atoi((params["companyid"]))

	json.NewDecoder(request.Body).Decode(&user)

	client := db.InitializeDatabase()
	defer client.Disconnect(context.Background())
	collection := client.Database("SPEC-CENTER").Collection("user")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err := collection.FindOne(ctx, bson.M{"email": user.Email}).Decode(&dbUser)
	if err != nil {
		response.WriteHeader(http.StatusInternalServerError)
		response.Write([]byte(`{"message":"` + "Please provide correct email. " + err.Error() + `"}`))
		return
	}

	collection = client.Database("SPEC-CENTER").Collection("role")
	err = collection.FindOne(ctx, primitive.M{"userid": dbUser.ID, "companyid": companyID}).Decode(&role)
	if err != nil {
		response.WriteHeader(http.StatusInternalServerError)
		response.Write([]byte(`{"message":"` + "Please provide correct companyid and other details. " + err.Error() + `"}`))
		return
	}
	userRole := role.Role
	dbUserId := dbUser.ID
	userPass := []byte(user.Password)
	dbPass := []byte(dbUser.Password)

	// checking companyid is correct or not
	//if companyID != role.CompanyId {
	//	authorization.WriteError(http.StatusBadRequest, "wrong company id", response, errors.New("wrong company"))
	//	return
	//}

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


