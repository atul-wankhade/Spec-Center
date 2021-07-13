package controller

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/atul-wankhade/Spec-Center/db"
	"github.com/atul-wankhade/Spec-Center/model"
	"github.com/atul-wankhade/Spec-Center/utils"

	"github.com/dgrijalva/jwt-go"
	"go.mongodb.org/mongo-driver/bson"
	"golang.org/x/crypto/bcrypt"
)

var SECRET = utils.GetEnvVariable("SECRET")
var SECRET_KEY = []byte(SECRET)

func LoginHandler(response http.ResponseWriter, request *http.Request) {
	response.Header().Set("Content-Type", "application/json")

	var user model.User
	var dbUser model.User

	json.NewDecoder(request.Body).Decode(&user)

	client := db.InitializeDatabase()
	defer client.Disconnect(context.Background())
	collection := client.Database(utils.Database).Collection(utils.UserCollection)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err := collection.FindOne(ctx, bson.M{"email": user.Email}).Decode(&dbUser)
	if err != nil {
		response.WriteHeader(http.StatusInternalServerError)
		response.Write([]byte(`{"message":"` + "Please provide correct email. " + err.Error() + `"}`))
		return
	}

	userEmail := dbUser.Email
	userPass := []byte(user.Password)
	dbPass := []byte(dbUser.Password)

	passErr := bcrypt.CompareHashAndPassword(dbPass, userPass)
	if passErr != nil {
		log.Println(passErr)
		response.Write([]byte(`{"response":"Wrong Password!"}`))
		return
	}
	var superAdminTag bool

	userRoleCollection := client.Database(utils.Database).Collection(utils.CompanyRolesCollection)
	userIsSuperadmin := userRoleCollection.FindOne(ctx, bson.M{"user_id": dbUser.ID.Hex(), "role": "superadmin"})
	if userIsSuperadmin.Err() != nil {
		superAdminTag = false
	}
	superAdminTag = true

	jwtToken, err := GenerateJWT(userEmail, dbUser.ID.Hex(), superAdminTag)
	if err != nil {
		response.WriteHeader(http.StatusInternalServerError)
		response.Write([]byte(`{"message":"` + err.Error() + `"}`))
		return
	}
	response.Write([]byte(`{"token":"` + jwtToken + `"}`))
}

func GenerateJWT(userEmail string, userID string, superAdminTag bool) (string, error) {
	token := jwt.New(jwt.SigningMethodHS256)
	claims := token.Claims.(jwt.MapClaims)
	claims["authorized"] = true
	claims["user_email"] = userEmail
	claims["user_id"] = userID
	claims["superadmin"] = superAdminTag
	claims["exp"] = time.Now().Add(time.Minute * 10).Unix()
	tokenString, err := token.SignedString(SECRET_KEY)
	if err != nil {
		log.Println("Error in JWT token generation")
		return "", err
	}
	return tokenString, nil
}
