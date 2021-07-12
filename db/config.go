package db

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/atul-wankhade/Spec-Center/model"
	"github.com/atul-wankhade/Spec-Center/utils"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"gopkg.in/mgo.v2/bson"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// for preventing duplicate entry with same userid and articleid in user and article collection.
func Indexing() {
	client := InitializeDatabase()
	defer client.Disconnect(context.Background())

	// Index for user collection based on email
	userCollection := client.Database(utils.Database).Collection(utils.UserCollection)
	_, err := userCollection.Indexes().CreateOne(context.Background(), mongo.IndexModel{
		Keys: bson.M{
			"email": 1,
		},
		Options: options.Index().SetUnique(true),
	})
	if err != nil {
		log.Fatal(err)
	}

	// Index for role collection based on role name
	roleCollection := client.Database(utils.Database).Collection(utils.RolesCollection)
	_, err = roleCollection.Indexes().CreateOne(context.Background(), mongo.IndexModel{
		Keys: bson.M{
			"name": 1,
		},
		Options: options.Index().SetUnique(true),
	})
	if err != nil {
		log.Fatal(err)
	}

	// Index for company collection based on company name
	companyCollection := client.Database(utils.Database).Collection(utils.CompanyCollection)
	_, err = companyCollection.Indexes().CreateOne(context.Background(), mongo.IndexModel{
		Keys: bson.M{
			"name": 1,
		},
		Options: options.Index().SetUnique(true),
	})
	if err != nil {
		log.Fatal(err)
	}

	// 	articleCollection := client.Database(utils.Database).Collection("article")
	// 	_, err = articleCollection.Indexes().CreateOne(context.Background(), mongo.IndexModel{
	// 		Keys: bson.M{
	// 			"articleid": 1,
	// 		},
	// 		Options: options.Index().SetUnique(true),
	// 	})
	// 	if err != nil {
	// 		log.Fatal(err)
	// 	}
	// roleCollection := client.Database(utils.Database).Collection(utils.CompanyRolesCollection)
	// _, err = roleCollection.Indexes().CreateOne(context.Background(), mongo.IndexModel{
	// 	Keys:    []bson.M{{"email", int32(1)}, {"company_id", int32(1)}},
	// 	Options: options.Index().SetUnique(true),
	// })
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// 	companyCollection := client.Database(utils.Database).Collection("company")
	// 	_, err = companyCollection.Indexes().CreateOne(context.Background(), mongo.IndexModel{
	// 		Keys: bson.M{
	// 			"id": 1,
	// 		},
	// 		Options: options.Index().SetUnique(true),
	// 	})
	// 	if err != nil {
	// 		log.Fatal(err)
	// 	}
	// 	log.Println("Indexing done..!")
}

func AddRoles() {
	superadminRole := model.Role{ID: primitive.NewObjectID(), Name: "superadmin"}
	adminRole := model.Role{ID: primitive.NewObjectID(), Name: "admin"}
	memberRole := model.Role{ID: primitive.NewObjectID(), Name: "member"}
	anonymousRole := model.Role{ID: primitive.NewObjectID(), Name: "anonymous"}

	client := InitializeDatabase()
	defer client.Disconnect(context.Background())
	rolesCollection := client.Database(utils.Database).Collection(utils.RolesCollection)
	_, err := rolesCollection.InsertMany(context.Background(), []interface{}{superadminRole, adminRole, memberRole, anonymousRole})
	if err != nil {
		log.Println(err)
	}
	log.Println("roles entrys added")

}
func CheckRole(userRole string) bool {
	if userRole == "superadmin" {
		return false
	}
	client := InitializeDatabase()
	defer client.Disconnect(context.Background())
	rolecollection := client.Database(utils.Database).Collection(utils.RolesCollection)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	result := rolecollection.FindOne(ctx, primitive.M{"name": userRole})
	return result.Err() == nil
}

var gslabID, kpointID, gslabUserID, kpointUserID primitive.ObjectID

// SuperadminEntry for entering  default superadmin and its role for each company in database.
func SuperadminEntry() {
	// retrieving password from env variables
	passSuperadminGSLAB := utils.GetEnvVariable("gslab_pass")
	passSuperadminKpoint := utils.GetEnvVariable("kpoint_pass")

	fmt.Println("!!!!!!!!!!!", passSuperadminKpoint, passSuperadminGSLAB)

	client := InitializeDatabase()
	defer client.Disconnect(context.Background())
	userCollection := client.Database(utils.Database).Collection(utils.UserCollection)
	roleCollection := client.Database(utils.Database).Collection(utils.CompanyRolesCollection)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var superadminGSLAB, superadminKpoint model.User
	gslabID = primitive.NewObjectID()
	kpointID = primitive.NewObjectID()
	gslabUserID = primitive.NewObjectID()
	kpointUserID = primitive.NewObjectID()

	superadminGSLAB.ID = gslabUserID
	superadminGSLAB.FirstName = "atul"
	superadminGSLAB.LastName = "wankhade"
	superadminGSLAB.Email = "atul@gmail.com"
	superadminGSLAB.Password = utils.GetHash([]byte(passSuperadminGSLAB))

	superadminKpoint.ID = kpointUserID
	superadminKpoint.FirstName = "bhushan"
	superadminKpoint.LastName = "gupta"
	superadminKpoint.Email = "bhushan@gmail.com"
	superadminKpoint.Password = utils.GetHash([]byte(passSuperadminKpoint))

	_, err := userCollection.InsertMany(ctx, []interface{}{superadminGSLAB, superadminKpoint})
	if err != nil {
		log.Println(err)
	}
	var roleForGSLAB, roleForKpoint model.UserRole
	roleForKpoint.UserEmail = "bhushan@gmail.com"
	roleForKpoint.CompanyId = kpointID.Hex()
	roleForKpoint.Role = "superadmin"

	roleForGSLAB.CompanyId = gslabID.Hex()
	roleForGSLAB.Role = "superadmin"
	roleForGSLAB.UserEmail = "atul@gmail.com"
	_, err = roleCollection.InsertMany(ctx, []interface{}{roleForGSLAB, roleForKpoint})

	if err != nil {
		log.Println(err, "role not added for superadmin user in database")
	}
	log.Println("Superadmin entries inserted")
}

func CompanyEntry() {
	var gslab, kpoint model.Company
	gslab = model.Company{ID: gslabID, Name: "gslab"}
	kpoint = model.Company{ID: kpointID, Name: "kpoint"}
	client := InitializeDatabase()
	defer client.Disconnect(context.Background())
	companyCollection := client.Database(utils.Database).Collection(utils.CompanyCollection)
	_, err := companyCollection.InsertMany(context.Background(), []interface{}{gslab, kpoint})
	if err != nil {
		log.Println(err)
	}
	log.Println("company entrys added")
}

func InitializeDatabase() *mongo.Client {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(utils.MongoUrl))

	if err != nil {
		log.Fatal(err)
	} else {
		log.Println("Connected to Database")
	}
	return client
}
