package db

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/atul-wankhade/Spec-Center/model"
	"github.com/atul-wankhade/Spec-Center/utils"
	"go.mongodb.org/mongo-driver/bson"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// for preventing duplicate entry with same userid and articleid in user and article collection.
func Indexing() {
	client := InitializeDatabase()
	defer client.Disconnect(context.Background())
	userCollection := client.Database("SPEC-CENTER").Collection("user")
	_, err := userCollection.Indexes().CreateOne(context.Background(), mongo.IndexModel{
		Keys: bson.M{
			"id": 1,
		},
		Options: options.Index().SetUnique(true),
	})
	if err != nil {
		log.Fatal(err)
	}

	_, err = userCollection.Indexes().CreateOne(context.Background(), mongo.IndexModel{
		Keys: bson.M{
			"email": 1,
		},
		Options: options.Index().SetUnique(true),
	})
	if err != nil {
		log.Fatal(err)
	}

	articleCollection := client.Database("SPEC-CENTER").Collection("article")
	_, err = articleCollection.Indexes().CreateOne(context.Background(), mongo.IndexModel{
		Keys: bson.M{
			"articleid": 1,
		},
		Options: options.Index().SetUnique(true),
	})
	if err != nil {
		log.Fatal(err)
	}
	roleCollection := client.Database("SPEC-CENTER").Collection("role")
	_, err = roleCollection.Indexes().CreateOne(context.Background(), mongo.IndexModel{
		Keys: bson.M{
			"userid": 1,
		},
		Options: options.Index().SetUnique(true),
	})
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Indexing done..!")
}

// SuperadminEntry for entering  default superadmin and its role for each company in database.
func SuperadminEntry() {
	// retrieving password from env variables
	passSuperadminGSLAB := utils.GetEnvVariable("gslab_pass")
	passSuperadminIBM := utils.GetEnvVariable("ibm_pass")

	fmt.Println("!!!!!!!!!!!", passSuperadminIBM, passSuperadminGSLAB)

	client := InitializeDatabase()
	defer client.Disconnect(context.Background())
	userCollection := client.Database("SPEC-CENTER").Collection("user")
	roleCollection := client.Database("SPEC-CENTER").Collection("role")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var superadminGSLAB, superadminIBM model.User

	superadminGSLAB.ID = 1
	superadminGSLAB.FirstName = "atul"
	superadminGSLAB.LastName = "wankhade"
	superadminGSLAB.Email = "atul@gslab.com"
	superadminGSLAB.Password = utils.GetHash([]byte(passSuperadminGSLAB))

	superadminIBM.ID = 2
	superadminIBM.FirstName = "bhushan"
	superadminIBM.LastName = "gupta"
	superadminIBM.Email = "bhushan@ibm.com"
	superadminIBM.Password = utils.GetHash([]byte(passSuperadminIBM))

	_, err := userCollection.InsertMany(ctx, []interface{}{superadminGSLAB, superadminIBM})
	if err != nil {
		log.Println(err)
	}
	var roleForGSLAB, roleForIBM model.Roles
	roleForGSLAB.CompanyId = 1
	roleForGSLAB.UserId = 1
	roleForGSLAB.Role = "superadmin"

	roleForIBM.CompanyId = 2
	roleForIBM.UserId = 2
	roleForIBM.Role = "superadmin"
	_, err = roleCollection.InsertMany(ctx, []interface{}{roleForGSLAB, roleForIBM})

	if err != nil {
		log.Println(err, "role not added for superadmin user in database")
	}
	log.Println("Superadmin entries inserted")
}

func InitializeDatabase() *mongo.Client {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// for running on docker, mongoservice is docker container name mentioned in docker-compose.
	client, err := mongo.Connect(ctx, options.Client().ApplyURI("mongodb://mongoservice:27017"))

	// for running locally
	//client, err := mongo.Connect(ctx, options.Client().ApplyURI("mongodb://localhost:27017"))
	if err != nil {
		log.Fatal(err)
	} else {
		log.Println("Connected to Database")
	}
	return client
}
