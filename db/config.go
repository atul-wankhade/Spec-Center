package db

import (
	"context"
	"go.mongodb.org/mongo-driver/bson"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// for preventing duplicate entry with same userid and articleid in user and article collection.
func Indexing(){
	client := InitializeDatabase()
	userCollection := client.Database("SPEC-CENTER").Collection("user")
	_, err := userCollection.Indexes().CreateOne(context.Background(), mongo.IndexModel{
		Keys: bson.M{
			"id": 1,
		},
		Options: options.Index().SetUnique(true),
	},)
	if err != nil {
		log.Fatal(err)
	}
	articleCollection := client.Database("SPEC-CENTER").Collection("article")
	_, err = articleCollection.Indexes().CreateOne(context.Background(), mongo.IndexModel{
		Keys: bson.M{
			"articleid": 1,
		},
		Options: options.Index().SetUnique(true),
	},)
	if err != nil {
		log.Fatal(err)
	}
}

func InitializeDatabase() *mongo.Client {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	client, err := mongo.Connect(ctx, options.Client().ApplyURI("mongodb://localhost:27017"))
	if err != nil {
		log.Fatal(err)
	} else {
		log.Println("Connected to Database")
	}
	return client
}
