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

	"github.com/dgrijalva/jwt-go"
)

func AddUser(response http.ResponseWriter, request *http.Request, claims jwt.MapClaims) {
	response.Header().Set("Content-Type", "application/json")
	var user model.User
	var userRole model.UserRole
	params := mux.Vars(request)
	companyID, ok := params["company_id"]
	if !ok {
		authorization.WriteError(http.StatusInternalServerError, "Decode Error", response, errors.New("unable to get companyid from claims"))
		return
	}

	type tempUser struct {
		FirstName string `json:"first_name" bson:"first_name"`
		LastName  string `json:"last_name" bson:"last_name"`
		Email     string `json:"email" bson:"email"`
		Password  string `json:"password" bson:"password"`
		Role      string `json: "role"`
	}

	var userHolder tempUser

	err := json.NewDecoder(request.Body).Decode(&userHolder)
	if err != nil {
		authorization.WriteError(http.StatusBadRequest, "BAD REQUEST BODY", response, err)
		return
	}

	user.ID = primitive.NewObjectID()
	user.FirstName = userHolder.FirstName
	user.LastName = userHolder.LastName
	user.Password = userHolder.Password
	user.Email = userHolder.Email

	if userHolder.FirstName == "" || userHolder.LastName == "" || userHolder.Password == "" || userHolder.Email == "" {
		authorization.WriteError(http.StatusBadRequest, "Invalid payload or nil body parameter", response, errors.New("invalid request body"))
		return
	}

	client := db.InitializeDatabase()
	defer client.Disconnect(context.Background())
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// checking provided role is valid or not
	valid := CheckRole(userHolder.Role)
	if !valid {
		authorization.WriteError(http.StatusBadRequest, "Invalid user role provided", response, errors.New("wrong role"))
		return
	}

	user.Password = utils.GetHash([]byte(user.Password))

	collection := client.Database(utils.Database).Collection(utils.UserCollection)
	_, err = collection.InsertOne(ctx, user)
	if err != nil {
		response.WriteHeader(http.StatusConflict)
		response.Write([]byte(`{"message":"` + fmt.Sprintf("User with  email : %s is already present in database!", user.Email) + `"}`))
		return
	}
	// user role insertion in roles collection
	userRole.UserEmail = user.Email
	userRole.CompanyId = companyID
	userRole.Role = userHolder.Role

	collection = client.Database(utils.Database).Collection(utils.CompanyRolesCollection)
	_, err = collection.InsertOne(ctx, userRole)
	if err != nil {
		response.WriteHeader(http.StatusInternalServerError)
		response.Write([]byte(`{"message":"` + err.Error() + `"}`))
		return
	}

	response.WriteHeader(http.StatusAccepted)
	response.Write([]byte(`{"message":"` + fmt.Sprintf("User with email: %s is added to company having id: %s with role: %s", user.Email, companyID, userHolder.Role) + `"}`))
}
