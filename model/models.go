package model

type User struct {
	ID        int    `json:"userid" bson:"id"`
	FirstName string `json:"firstname" bson:"firstname"`
	LastName  string `json:"lastname" bson:"lastname"`
	Email     string `json:"email" bson:"email"`
	Password  string `json:"password" bson:"password"`
}

type Company struct {
	ID         string   `json:"id" bson:"id"`
	Name       string   `json:"name" bson:"name"`
	Superadmin int      `json:"superadmin" bson:"superadmin"`
	Admins     []string `json:"admins" bson:"admins"`
	Members    []string `json:"members" bson:"members"`
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
	ComapanyID string   `json:"companyid" bson:"companyid"`
	ArticleID  string   `json:"articleid" bson:"articleid"`
	Admins     []string `json:"admins" bson:"admins"`
	Members    []string `json:"members" bson:"members"`
	Body       string   `json:"body" bson:"body"`
}
