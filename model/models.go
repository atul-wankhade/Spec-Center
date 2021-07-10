package model

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type Database struct {
	Client *mongo.Client
}

type User struct {
	ID        primitive.ObjectID `json:"_id" bson:"_id"`
	FirstName string             `json:"first_name" bson:"first_name"`
	LastName  string             `json:"last_name" bson:"last_name"`
	Email     string             `json:"email" bson:"email"`
	Password  string             `json:"password" bson:"password"`
}

type Company struct {
	ID   primitive.ObjectID `json:"_id" bson:"_id"`
	Name string             `json:"name" bson:"name"`
}

type Roles struct {
	UserEmail string `json:"email" bson:"email"`
	CompanyId string `json:"company_id" bson:"company_id"`
	Role      string `json:"role" bson:"role"`
}

type ArticleRole struct {
	UserId    string `json:"user_id" bson:"user_id"`
	CompanyId string `json:"company_id" bson:"company_id"`
	ArticleId string `json:"article_id" bson:"article_id"`
	Role      string `json:"role" bson:"role"`
}

type Article struct {
	ID         primitive.ObjectID `json:"_id" bson:"_id"`
	ComapanyID int                `json:"company_id" bson:"company_id"`
	Body       string             `json:"body" bson:"body"`
}

type NewEntity struct {
	Name      string `json:"name" bson:"name"`
	ID        int    `json:"id" bson:"id"`
	CompanyID int    `json:"company_id" bson:"company_id"`
}

type Role struct {
	ID   primitive.ObjectID `json:"_id" bson:"_id"`
	Name string             `json:"name" bson:"name"`
}
