package controller

import (
	"Spec-Center/authorization"
	"Spec-Center/db"
	"Spec-Center/model"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/dgrijalva/jwt-go"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

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

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	client := db.InitializeDatabase()
	defer client.Disconnect(context.Background())
	collection := client.Database("SPEC-CENTER").Collection("role")
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
	client := db.InitializeDatabase()
	defer client.Disconnect(context.Background())
	collection := client.Database("SPEC-CENTER").Collection("role")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
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
	collection1 := client.Database("SPEC-CENTER").Collection("articlerole")
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

// for changing user role on all article when company role for user is changed.
func updateUserArticleRoles(userID, companyID int, updatedRole string) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	client := db.InitializeDatabase()
	defer client.Disconnect(context.Background())
	collection := client.Database("SPEC-CENTER").Collection("articlerole")
	filter := primitive.M{"userid": userID, "companyid": companyID}
	opts := options.Update().SetUpsert(true)
	update := bson.D{{"$set", bson.D{{"role", updatedRole}}}}

	_, err := collection.UpdateMany(ctx, filter, update, opts)
	if err != nil {
		log.Fatal(err)
	}
}