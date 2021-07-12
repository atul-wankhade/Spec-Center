package worker

/*
// import (
// 	"context"
// 	"log"
// 	"os"
// 	"os/signal"
// 	"syscall"
// 	"time"

// 	"github.com/atul-wankhade/Spec-Center/db"
// 	"github.com/atul-wankhade/Spec-Center/model"
// 	"go.mongodb.org/mongo-driver/bson/primitive"
// 	"go.mongodb.org/mongo-driver/mongo"
// )

// func Worker() {
// 	tick := time.NewTicker(time.Minute * 3)
// 	done := make(chan bool)
// 	go Scheduler(tick, done)
// 	sigs := make(chan os.Signal, 1)
// 	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
// 	<-sigs

// 	done <- true

// }

// func Scheduler(tick *time.Ticker, done chan bool) {
// 	Task(time.Now())
// 	for {
// 		select {
// 		case t := <-tick.C:
// 			Task(t)
// 		case <-done:
// 			return
// 		}
// 	}
// }

// func Task(t time.Time) {
// 	client := db.InitializeDatabase()
// 	defer client.Disconnect(context.Background())
// 	database := client.Database("SPEC-CENTER")
// 	count, err := database.Collection("newentity").CountDocuments(context.Background(), primitive.M{}, nil)
// 	if err != nil || count == 0 {
// 		log.Printf("no new entities added, error: %w", err)
// 	}
// 	if count != 0 {
// 		cursor, err := database.Collection("newentity").Find(context.Background(), primitive.M{})
// 		if err != nil {
// 			log.Println(err)
// 		}
// 		defer cursor.Close(context.Background())
// 		for cursor.Next(context.Background()) {
// 			var entity model.NewEntity
// 			err = cursor.Decode(&entity)
// 			if err != nil {
// 				log.Println("decode error :", err)
// 			}
// 			if entity.Name == "article" {
// 				done := insertRolesForNewArticle(entity, database)
// 				if done {
// 					_, err := database.Collection("newentity").DeleteOne(context.Background(), primitive.M{"id": entity.ID})
// 					if err != nil {
// 						log.Println("unable to delete new entity job")
// 					}
// 				}
// 			}
// 		}
// 	}

// }

// func insertRolesForNewArticle(entity model.NewEntity, database *mongo.Database) bool {
// 	companyRoleCollection := database.Collection("role")
// 	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
// 	defer cancel()
// 	filter := primitive.M{"companyid": entity.CompanyID}
// 	cursor, err := companyRoleCollection.Find(ctx, filter)
// 	if err != nil {
// 		log.Println(err)
// 		return false
// 	}
// 	defer cursor.Close(ctx)

// 	articleRoleCollection := database.Collection("articlerole")
// 	var articleRole model.ArticleRole
// 	var userRole model.Roles

// 	for cursor.Next(ctx) {
// 		err := cursor.Decode(&userRole)
// 		if err != nil {
// 			log.Println(err)
// 			return false
// 		}
// 		articleRole.ArticleId = entity.ID
// 		articleRole.CompanyId = entity.CompanyID
// 		articleRole.Role = userRole.Role
// 		articleRole.UserId = userRole.UserId

// 		_, err = articleRoleCollection.InsertOne(ctx, articleRole)
// 		if err != nil {
// 			log.Printf("Failed to add article role for article id : %d, user id : %d, error : %w", entity.ID, userRole.UserId, err)
// 			return false
// 		}
// 		log.Printf("Role on new article with article id : %d, for user id : %d , for company id : %d added successfully", entity.ID, userRole.UserId, entity.CompanyID)
// 	}
// 	return true
// }
*/