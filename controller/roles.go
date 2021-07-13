package controller

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/atul-wankhade/Spec-Center/authorization"
	"github.com/atul-wankhade/Spec-Center/db"
	"github.com/atul-wankhade/Spec-Center/model"
	"github.com/atul-wankhade/Spec-Center/utils"
	"github.com/gorilla/mux"

	"github.com/dgrijalva/jwt-go"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func UpdateCompanyRoleHandler(w http.ResponseWriter, r *http.Request, claims jwt.MapClaims) {
	w.Header().Set("Content-Type", "application/json")
	params := mux.Vars(r)
	companyID, ok := params["company_id"]
	if !ok {
		authorization.WriteError(http.StatusInternalServerError, "Decode Error", w, errors.New("unable to get companyid from parameters"))
		return
	}
	userID, ok := params["user_id"]
	if !ok {
		authorization.WriteError(http.StatusInternalServerError, "Decode Error", w, errors.New("unable to get user id from parameters"))
		return
	}

	var role model.UserRole
	err := json.NewDecoder(r.Body).Decode(&role)
	if err != nil {
		authorization.WriteError(http.StatusBadRequest, "DECODE ERROR", w, err)
		return
	}

	role.CompanyId = companyID
	role.UserID = userID

	superadminID := fmt.Sprintf("%v", claims["user_id"])
	log.Println("superadminID---->", superadminID)

	if superadminID == userID {
		authorization.WriteError(http.StatusBadRequest, "Cannot change own role", w, errors.New("cannot update own role"))
		return
	}

	valid := CheckRole(role.Role)
	if !valid {
		authorization.WriteError(http.StatusBadRequest, "Invalid user role provided", w, errors.New("wrong role"))
		return
	}

	userObjectID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		authorization.WriteError(http.StatusBadRequest, "Invalid user_id!", w, errors.New("wrong user_id"))
		return
	}

	client := db.InitializeDatabase()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	defer client.Disconnect(ctx)

	userExists := UserExists(userObjectID)
	if !userExists {
		authorization.WriteError(http.StatusBadRequest, "User does not exists in database!", w, errors.New("wrong user_id"))
		return
	}

	collection := client.Database(utils.Database).Collection(utils.CompanyRolesCollection)
	filter := primitive.M{"user_id": role.UserID, "company_id": role.CompanyId}
	var opts *options.UpdateOptions
	if r.Method == "PUT" {
		opts = options.Update().SetUpsert(false)
		update := bson.D{{"$set", bson.D{{"role", role.Role}}}}
		result, err := collection.UpdateOne(ctx, filter, update, opts)

		if err != nil {
			authorization.WriteError(http.StatusInternalServerError, "update error", w, err)
			return
		}
		if result.ModifiedCount == 0 && r.Method == "PUT" {
			authorization.WriteError(http.StatusBadRequest, "Trying to update with same role.", w, errors.New("BAD REQUEST, User not present"))
			return
		}

		go updateUserArticleRoles(role.UserID, role.CompanyId, role.Role)

		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"message":"` + fmt.Sprintf("Role for user with id :%s  is changed to: %s", role.UserID, role.Role) + `"}`))
	} else {
		opts = options.Update().SetUpsert(true)
		_, err = collection.InsertOne(ctx, role)
		if err != nil {
			authorization.WriteError(http.StatusConflict, "Duplicate User Entry!", w, errors.New("user already in company"))
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"message":"` + fmt.Sprintf("User with id :%s is added to comapny with company_id: %s with role:%s", role.UserID, companyID, role.Role) + `"}`))
	}
}

func UpdateArticleRoleHandler(w http.ResponseWriter, r *http.Request, claims jwt.MapClaims) {
	w.Header().Set("Content-Type", "application/json")

	params := mux.Vars(r)

	companyID, ok := params["company_id"]
	if !ok {
		authorization.WriteError(http.StatusInternalServerError, "Decode Error", w, errors.New("unable to get companyid from parameters"))
		return
	}
	userID, ok := params["user_id"]
	if !ok {
		authorization.WriteError(http.StatusInternalServerError, "Decode Error", w, errors.New("unable to get user id from parameters"))
		return
	}

	articleID, ok := params["article_id"]
	if !ok {
		authorization.WriteError(http.StatusInternalServerError, "Decode Error", w, errors.New("unable to get article_id from parameters"))
		return
	}

	var articleRole model.ArticleRole

	err := json.NewDecoder(r.Body).Decode(&articleRole)
	if err != nil {
		authorization.WriteError(http.StatusBadRequest, "DECODE ERROR", w, err)
		return
	}

	articleRole.ArticleId = articleID
	articleRole.CompanyId = companyID
	articleRole.UserID = userID

	superadminID := fmt.Sprintf("%v", claims["user_id"])

	if superadminID == userID {
		authorization.WriteError(http.StatusBadRequest, "Cannot change role on article for superadmin", w, errors.New("cannot update article role for superadmin"))
		return
	}

	valid := CheckRole(articleRole.Role)
	if !valid {
		authorization.WriteError(http.StatusBadRequest, "Invalid user role provided", w, errors.New("wrong role"))
		return
	}

	var role model.UserRole

	client := db.InitializeDatabase()
	defer client.Disconnect(context.Background())
	roleCollection := client.Database(utils.Database).Collection(utils.CompanyRolesCollection)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	err = roleCollection.FindOne(ctx, primitive.M{"user_id": userID, "company_id": companyID}).Decode(&role)
	if err != nil {
		authorization.WriteError(http.StatusBadRequest, "Please provide correct user id..!, user not present", w, err)
		return
	}
	if role.Role == "anonymous" {
		authorization.WriteError(http.StatusBadRequest, "BAD REQUEST:- anonymous user can't be given article access", w, errors.New("BAD REQUEST"))
		return
	}

	articleCollection := client.Database(utils.Database).Collection(utils.ArticleCollection)

	articleObjectID, err := primitive.ObjectIDFromHex(articleID)
	if err != nil {
		authorization.WriteError(http.StatusBadRequest, "wrong article_id.", w, err)
		return
	}

	filter := bson.M{"company_id": companyID, "_id": articleObjectID}
	result := articleCollection.FindOne(ctx, filter)
	if result.Err() != nil {
		authorization.WriteError(http.StatusBadRequest, "wrong article_id, no article present with this id in database", w, result.Err())
		return
	}

	articleRoleCollection := client.Database(utils.Database).Collection(utils.ArticleRoleCollection)
	opts := options.Update().SetUpsert(true)
	filter2 := primitive.M{"user_id": userID, "company_id": companyID, "article_id": articleID}
	update := primitive.M{"$set": primitive.M{"role": articleRole.Role}}

	_, err = articleRoleCollection.UpdateOne(ctx, filter2, update, opts)
	if err != nil {
		authorization.WriteError(http.StatusInternalServerError, "update error", w, err)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"message":"` + fmt.Sprintf("Role for user with id:%s for articleid: %s is changed to: %s", userID, articleRole.ArticleId, articleRole.Role) + `"}`))
}

// for deleting user role on special article when company role for user is changed, so we can used default role on all article
func updateUserArticleRoles(userID, companyID string, updatedRole string) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	client := db.InitializeDatabase()
	defer client.Disconnect(context.Background())
	collection := client.Database(utils.Database).Collection(utils.ArticleRoleCollection)
	filter := primitive.M{"user_id": userID, "company_id": companyID, "role": updatedRole}
	_, err := collection.DeleteMany(ctx, filter)
	if err != nil {
		log.Println("Error while deleting user article roles", err)
	}
}

func CheckRole(userRole string) bool {
	if userRole == "superadmin" {
		return false
	}
	client := db.InitializeDatabase()
	defer client.Disconnect(context.Background())
	rolecollection := client.Database(utils.Database).Collection(utils.RolesCollection)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	result := rolecollection.FindOne(ctx, primitive.M{"name": userRole})
	return result.Err() == nil
}

func UserExists(userID primitive.ObjectID) bool {
	client := db.InitializeDatabase()
	defer client.Disconnect(context.Background())
	userCollection := client.Database(utils.Database).Collection(utils.UserCollection)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	result := userCollection.FindOne(ctx, primitive.M{"_id": userID})
	if result.Err() != nil {
		return false
	}
	return true
}
