package utils

import (
	// godot package to load/read the .env file and
	"log"
	"os"

	"github.com/joho/godotenv"
	"golang.org/x/crypto/bcrypt"
)

const(
	//Collection names
	// MongoUrl = "mongodb://mongoservice:27017"
	MongoUrl = "mongodb://localhost:27017"
	Database = "SPEC-CENTER"
	CompanyRolesCollection = "company_roles"
	UserCollection = "user"
	ArticleCollection = "article"
	ArticleRoleCollection = "article_role"
	RolesCollection = "role"
	NewEntityCollection = "new_entity"
	CompanyCollection = "company"

	//
)

func GetHash(pwd []byte) string {
	hash, err := bcrypt.GenerateFromPassword(pwd, bcrypt.MinCost)
	if err != nil {
		log.Println(err)
	}
	return string(hash)
}

// return the value of the key enviroment variable
func GetEnvVariable(key string) string {
	// load .env file
	err := godotenv.Load(".env")

	if err != nil {
		log.Fatalf("Error loading .env file")
	}
	return os.Getenv(key)
}
