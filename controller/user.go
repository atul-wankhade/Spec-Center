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
	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/dgrijalva/jwt-go"
)

func AddUser(response http.ResponseWriter, request *http.Request, claims jwt.MapClaims) {
	response.Header().Set("Content-Type", "application/json")
	var user model.User

	err := json.NewDecoder(request.Body).Decode(&user)
	if err != nil {
		authorization.WriteError(http.StatusBadRequest, "BAD REQUEST BODY", response, err)
		return
	}

	user.ID = primitive.NewObjectID()

	if user.FirstName == "" || user.LastName == "" || user.Password == "" || user.Email == "" {
		authorization.WriteError(http.StatusBadRequest, "Invalid payload or nil body parameter", response, errors.New("invalid request body"))
		return
	}

	client := db.InitializeDatabase()
	defer client.Disconnect(context.Background())
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	user.Password = utils.GetHash([]byte(user.Password))

	collection := client.Database(utils.Database).Collection(utils.UserCollection)
	_, err = collection.InsertOne(ctx, user)
	if err != nil {
		response.WriteHeader(http.StatusConflict)
		response.Write([]byte(`{"message":"` + fmt.Sprintf("User with  email : %s is already present in database!", user.Email) + `"}`))
		return
	}
	response.WriteHeader(http.StatusAccepted)
	response.Write([]byte(`{"message":"` + fmt.Sprintf("User with email: %s is added to database with id :- %s", user.Email, user.ID.Hex()) + `"}`))
}
