package controller

// import (
// 	"context"
// 	"encoding/json"
// 	"errors"
// 	"fmt"
// 	"log"
// 	"net/http"
// 	"strconv"
// 	"time"

// 	"github.com/atul-wankhade/Spec-Center/authorization"
// 	"github.com/atul-wankhade/Spec-Center/db"
// 	"github.com/atul-wankhade/Spec-Center/model"
// 	"github.com/gorilla/mux"

// 	"github.com/dgrijalva/jwt-go"
// 	"go.mongodb.org/mongo-driver/bson"
// 	"go.mongodb.org/mongo-driver/bson/primitive"
// 	"go.mongodb.org/mongo-driver/mongo/options"
// )

// func UpdateCompanyRoleHandler(w http.ResponseWriter, r *http.Request, claims jwt.MapClaims) {
// 	w.Header().Set("Content-Type", "application/json")
// 	var role model.Roles
// 	err := json.NewDecoder(r.Body).Decode(&role)
// 	if err != nil {
// 		authorization.WriteError(http.StatusBadRequest, "DECODE ERROR", w, err)
// 		return
// 	}
// 	params := mux.Vars(r)
// 	companyID, err := strconv.Atoi((params["company_id"]))
// 	if err != nil {
// 		authorization.WriteError(http.StatusInternalServerError, "Decode Error", w, errors.New("unable to get companyid from claims"))
// 		return
// 	}

// 	if role.CompanyId != companyID {
// 		authorization.WriteError(http.StatusBadRequest, "Please provide correct company id ", w, errors.New("wrong companyid"))
// 		return
// 	}

// 	superadminId := int(claims["userid"].(float64))

// 	if superadminId == role.UserId {
// 		authorization.WriteError(http.StatusBadRequest, "Cannot change own role", w, errors.New("cannot update own role"))
// 		return
// 	}

// 	if role.Role != "admin" && role.Role != "member" && role.Role != "anonymous" {
// 		authorization.WriteError(http.StatusBadRequest, "invalid role provided", w, errors.New("invalid role"))
// 		return
// 	}

// 	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
// 	defer cancel()
// 	client := db.InitializeDatabase()
// 	defer client.Disconnect(context.Background())
// 	collection := client.Database("SPEC-CENTER").Collection("role")
// 	filter := primitive.M{"userid": role.UserId, "companyid": role.CompanyId}
// 	opts := options.Update().SetUpsert(false)
// 	update := bson.D{{"$set", bson.D{{"role", role.Role}}}}

// 	result, err := collection.UpdateOne(ctx, filter, update, opts)
// 	if err != nil {
// 		authorization.WriteError(http.StatusInternalServerError, "update error", w, err)
// 		return
// 	}
// 	if result.ModifiedCount == 0 {
// 		authorization.WriteError(http.StatusBadRequest, "BAD REQUEST, User not present", w, errors.New("BAD REQUEST, User not present"))
// 		return
// 	}

// 	go updateUserArticleRoles(role.UserId, role.CompanyId, role.Role)

// 	w.WriteHeader(http.StatusOK)
// 	w.Write([]byte(`{"message":"` + fmt.Sprintf("Role for userid:%d  is changed to: %s", role.UserId, role.Role) + `"}`))
// }

func UpdateArticleRoleHandler(w http.ResponseWriter, r *http.Request, claims jwt.MapClaims) {
	w.Header().Set("Content-Type", "application/json")
	var articleRole model.ArticleRole

	err := json.NewDecoder(r.Body).Decode(&articleRole)
	if err != nil {
		authorization.WriteError(http.StatusBadRequest, "DECODE ERROR", w, err)
		return
	}
	params := mux.Vars(r)
	companyID, err := strconv.Atoi((params["company_id"]))
	if err != nil {
		authorization.WriteError(http.StatusInternalServerError, "Decode Error", w, errors.New("unable to get companyid from claims"))
		return
	}

	if articleRole.CompanyId != companyID {
		authorization.WriteError(http.StatusBadRequest, "Please provide correct company id ", w, errors.New("BAD REQUEST"))
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

	client := db.InitializeDatabase()
	defer client.Disconnect(context.Background())
	roleCollection := client.Database("SPEC-CENTER").Collection("role")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	err = roleCollection.FindOne(ctx, primitive.M{"userid": articleRole.UserId, "companyid": articleRole.CompanyId}).Decode(&role)
	if err != nil {
		authorization.WriteError(http.StatusBadRequest, "Please provide correct userid!, user not present", w, err)
		return
	}
	if role.Role == "anonymous" {
		authorization.WriteError(http.StatusBadRequest, "BAD REQUEST:- anonymous user can't be given article access", w, errors.New("BAD REQUEST"))
		return
	}

	articleRoleCollection := client.Database("SPEC-CENTER").Collection("articlerole")
	opts := options.Update().SetUpsert(true)
	filter := primitive.M{"userid": articleRole.UserId, "companyid": articleRole.CompanyId, "articleid": articleRole.ArticleId}
	update := bson.D{{"$set", bson.D{{"role", articleRole.Role}}}}

	result, err := articleCollection.UpdateOne(ctx, filter, update, opts)
	if err != nil {
		authorization.WriteError(http.StatusInternalServerError, "update error", w, err)
		return
	}
	if result.ModifiedCount == 0 {
		authorization.WriteError(http.StatusBadRequest, "BAD REQUEST, Wrong Article Id", w, errors.New("BAD REQUEST, article not present"))
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"message":"` + fmt.Sprintf("Role for userid:%d for articleid: %d is changed to: %s", articleRole.UserId, articleRole.ArticleId, articleRole.Role) + `"}`))
}

// // for changing user role on all article when company role for user is changed.
// func updateUserArticleRoles(userID, companyID int, updatedRole string) {
// 	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
// 	defer cancel()
// 	client := db.InitializeDatabase()
// 	defer client.Disconnect(context.Background())
// 	collection := client.Database("SPEC-CENTER").Collection("articlerole")
// 	filter := primitive.M{"userid": userID, "companyid": companyID}
// 	opts := options.Update().SetUpsert(false)
// 	update := bson.D{{"$set", bson.D{{"role", updatedRole}}}}

// 	_, err := collection.UpdateMany(ctx, filter, update, opts)
// 	if err != nil {
// 		log.Println("Error while updating user article roles", err)
// 	}
// }
