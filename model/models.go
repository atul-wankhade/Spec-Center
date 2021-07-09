package model

import "go.mongodb.org/mongo-driver/mongo"

type Database struct {
	Client *mongo.Client
}

type User struct {
	ID        int    `json:"id" bson:"id"`
	FirstName string `json:"firstname" bson:"firstname"`
	LastName  string `json:"lastname" bson:"lastname"`
	Email     string `json:"email" bson:"email"`
	Password  string `json:"password" bson:"password"`
}

type Company struct {
	ID   int    `json:"id" bson:"id"`
	Name string `json:"name" bson:"name"`
}

type Roles struct {
	UserId    int    `json:"userid" bson:"userid"`
	CompanyId int    `json:"companyid" bson:"companyid"`
	Role      string `json:"role" bson:"role"`
}

type ArticleRole struct {
	UserId    int    `json:"userid" bson:"userid"`
	CompanyId int    `json:"companyid" bson:"companyid"`
	ArticleId int    `json:"articleid" bson:"articleid"`
	Role      string `json:"role" bson:"role"`
}

type Article struct {
	ComapanyID int    `json:"companyid" bson:"companyid"`
	ArticleID  int    `json:"articleid" bson:"articleid"`
	Body       string `json:"body" bson:"body"`
}

type NewEntity struct {
	Name      string `json:"name" bson:"name"`
	ID        int    `json:"id" bson:"id"`
	CompanyID int    `json:"companyid" bson:"companyid"`
}
