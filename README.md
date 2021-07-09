# Spec Center

# Requirement
1. Create a web application where there are three entities - Company, Article, Users and
   Roles.
2. Company has many Articles. User has and belongs to many Companies through Roles
   table.
3. Company will have one super admin
4. Super admin of the company be able to give user specific access to Company and
   Articles.
5. Accesses are like this.
6. A User can be an ADMIN of Company. He can see and edit all the articles that belongs
   to the company.
7. A User can be a MEMBER of a Company. The user can see the Articles, but cannot edit
   it.
8. If a User is a MEMBER of a Company, he can be given ADMIN access to an Article,
   which will let him to edit the Article.
9. A User with no access in Company shouldn’t see the articles.
10. Use Casbin for Authorisation Roles.
11. Write rest API to perform all operations. No need for html pages.
## Initial Setup
### Environment Variable Setup :-
#### Create .env file for all environment variable
#### 1. As we have two company right now so create password for superadmin in that file as follows :-
      ```
      ibm_pass=<password for ibm>
      gslab_pass=password for gslab>
      ```
#### 2. Also need to setup one more env variable in .env for jwt token secret as follows
      ```
      SECRET=<secret value>
      ```

### Mongo Setup
#### Database
#### Name :- SPEC-CENTER <br/>
#### Collections
#### 1. user :- Struct corresponding to database entity is as below.
   ```
   type User struct {
      ID        int    `json:"id" bson:"id"`
      FirstName string `json:"firstname" bson:"firstname"`
      LastName  string `json:"lastname" bson:"lastname"`
      Email     string `json:"email" bson:"email"`
      Password  string `json:"password" bson:"password"`
   }
   ```
#### 2. role :- Struct corresponding to database entity is as below.
   ```
   type Roles struct {
       UserId    int    `json:"userid" bson:"userid"`
       CompanyId int    `json:"companyid" bson:"companyid"`
       Role      string `json:"role" bson:"role"`
   }
   ```
#### 3. article:- Struct corresponding to database entity is as below.
   ```
   type Article struct {
      ComapanyID int   `json:"companyid" bson:"companyid"`
      ArticleID  int   `json:"articleid" bson:"articleid"`
      Body       string   `json:"body" bson:"body"`
   } 
   ```
#### 4. articlerole:- Struct corresponding to database entity is as below.
   ```
   type ArticleRole struct {
       UserId    int    `json:"userid" bson:"userid"`
       CompanyId int    `json:"companyid" bson:"companyid"`
       ArticleId int    `json:"articleid" bson:"articleid"`
       Role      string `json:"role" bson:"role"`
   }
   ```
#### 5. company:- Struct corresponding to database entity is as below.
   ```
   type Company struct {
       ID         string   `json:"id" bson:"id"`
       Name       string   `json:"name" bson:"name"`
   }
   ```

#### Initial Data in database
<br/>1. superadmin user details for each company should need to insert initially.
<br/>2. Also, same user with superadmin role need to be added in role collection.

## APIs List

#### Login in Company
```
   POST: localhost:8080/login/{companyid} 
```
#### To add User
```
   POST: localhost:8080/adduser
```
#### ARTICLE RELATED APIs
##### To get all article in company.
```
   GET: localhost:8080/all_articles
```
##### To create article for company.
```
   POST: localhost:8080/article
```
##### To update article in company.
```
   PUT: localhost:8080/article
```
##### To delete article in company.
```
   DELELTE: localhost:8080/article
```
### ROLE CHANGE APIs

#### To change role of user in company :- Only superadmin can change role of other user.<br/>
```
   PUT: localhost:8080/role <br/>
```
#### Only superadmin can change role of other user on particular article.<br/>
```
   PUT: localhost:8080/articlerole/{articleid} <br/>
```


## APIs

### User login

**API:** localhost:8080/login/{companyid}<br/>
**Method:** POST<br/>
**Payload**:
```
{
    "email": "<user email>",
    "password" :"<password>"
}
```

**Response:**
```
{
    "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJhdXRob3JpemVkIjp0cnVlLCJjb21wYW55aWQiOjEsImV4cCI6IjIwMjEtMDctMDZUMTA6MjQ6MTMuNzI5ODU2NTQ3KzA1OjMwIiwidXNlcmlkIjoxLCJ1c2Vycm9sZSI6InN1cGVyYWRtaW4ifQ.x8Ig1OU5JghF0pefemOWcbA_QwOVhqXETHStkhQnxjI"
}
```
#### Once user is logged inside company, comapanyid is taken from  login generated token.
#### following api is works for same logged in company.

### Add User
**API:** localhost:8080/adduser<br/>
**Method:** POST<br/>
**Description:** <br/>1. Only superadmin can add user in company, by giving role for user in json body.

<br/> 2. Once user is added in company, same time in role collection we are updating role of that user in company.<br/>

<br/> 3. After that, updating the articlerole collection for each article in company with newly user role on each article(articlerole).

**Payload**:
```
{
    "id" : <id>,
    "firstname":<firstname>,
    "lastname": <lastname>,
    "email": <email>
    "password" : <password>
    "role" : <role for new user in comapany>
}
```
**Response:**
```
{
    "message": "User with userid: 21  is added to company having id: 1 with role: admin"
}
```

#### GET all articles
**API:** localhost:8080/all_articles<br/>
**Method:** GET<br/>
**Description:** <br/>To Get all articles for login company, only admin, member and superadmin can see all articles<br/>
**Response:**
```
[
    {
        "companyid": 1,
        "articleid": 3,
        "body": "Blockchain Learning"
    },
    {
        "companyid": 1,
        "articleid": 10,
        "body": "Article by bhushan"
    },
    .
    .
    .
    .
]
```
#### Create article for company
**API:** localhost:8080/article<br/>
**Method:** POST<br/>
**Description:**
<br/>1. Only superadmin have access to this api, and it's checked by "Cashbin".<br/>
<br/>2. After adding article in article collection, role for that particular article for all user is added by performing insert query in articlerole collection.<br/>
**Payload**:
```
{
    "articleid":3,
    "companyid":1,
    "body": "Blockchain Learning"
}
```
**Response:**
```
{
    "InsertedID": "60e55c22b705d4c5af021b74"
}
```

#### Delete article by articleid
**API:** localhost:8080/article<br/>
**Method:** DELETE<br/>
**Params:**
<br/>Key : articleid, Value: int <br/>
**Description:**
<br/>1. superadmin,admin and member have access to this api, checked by "Cashbin".<br/>
<br/>2. After accesing the api, if user having admin or superadmin access on particular article that is checked by mongo articlerole collection, then only user allow to delete that article.<br/>
<br/>3. After deleting the article all entries related to that articleid in articlerole collection will be deleted<br/>

**Response:**
```
{
    "message": "Article with id: 1 is successfully deleted!"
}
```
#### Update article by articleid
**API:** localhost:8080/article<br/>
**Method:** PUT<br/>
**Description:**
<br/>1. superadmin,admin and member have access to this api, checked by "Cashbin".<br/>
<br/>2. After accessing the api, if user having admin or superadmin access on particular article that is checked by mongo articlerole collection, then only user allow to update that article.<br/>

**Payload**:
```
{
    "articleid":3,
    "companyid":1,
    "body": "Blockchain Learning"
}
```
**Response:**
```
{
    "message": "Article with id: 3 is successfully updated!"
}
```
#### Change company role of user.
**API:** localhost:8080/role<br/>
**Method:** PUT<br/>
**Description:**
<br/>1. Only superadmin have access to this api, and it's checked by "Cashbin".<br/>
<br/>2. After changing role of user, corresponding role is updated for each article in articlerole collection.<br/>
**Payload**:
```
{
    "userid": 3,
    "companyid": 1,
    "role": "anonymous"
}
```
**Response:**
```
{
    "message": "Role for userid:3  is changed to: anonymous"
}
```
#### Change role of user on particular article.
**API:** localhost:8080/articlerole<br/>
**Method:** PUT<br/>
**Description:**
<br/> Only superadmin have access to this api, and it's checked by "Cashbin" and can change role of user on particular article that will upated in articlerole collction. <br/>

**Payload**:
```
{
    "articleid": 1,
    "userid" : 4,
    "companyid": 1,
    "role" : "admin"
}
```
**Response:**
```
{
    "message": "Role for userid:3 for articleid: 1 is changed to: admin"
}
```