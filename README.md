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
   gslab_pass=password for gslab> 
   kpoint_pass=<password for kpoint>
```
#### 2. Also need to setup one more env variable in .env for jwt token secret as follows
```
   SECRET=<secret value>
```

### Mongo Setup
#### Database
#### Name :- SPEC-CENTER <br/>
#### Collections
#### 1. **user** :- This is for all users availables in database.
#### Struct corresponding to database entity is as below.
   ```
   type User struct {
      ID        primitive.ObjectID `json:"_id" bson:"_id"`
      FirstName string             `json:"first_name" bson:"first_name"`
      LastName  string             `json:"last_name" bson:"last_name"`
      Email     string             `json:"email" bson:"email"`
      Password  string             `json:"password" bson:"password"`
   }
   ```
#### 2. **role** :- This is for predefined valid roles that can be used for all companies
#### Struct corresponding to database entity is as below.
   ```
   type Role struct {
      ID   primitive.ObjectID `json:"_id" bson:"_id"`
      Name string             `json:"name" bson:"name"`  
   }
   ```
#### 3. **user_roles** :- This is for role of all user corresponding to their company
#### Struct corresponding to database entity is as below.
   ```
   type UserRole struct {
      UserID    string `json:"user_id" bson:"user_id"`
      CompanyId string `json:"company_id" bson:"company_id"`
      Role      string `json:"role" bson:"role"`
   }
   ```
  
#### 4. **article**:- This is for all articles of all companies in database.
#### Struct corresponding to database entity is as below.
   ```
   type Article struct {
      ID        primitive.ObjectID `json:"_id" bson:"_id"`
      CompanyID string             `json:"company_id" bson:"company_id"`
      Body      string             `json:"body" bson:"body"`
   }
   ```
#### 5. **article_role**:- We are using this when user have other role than its company role on particular article, so we have more control based on particular article.
#### Struct corresponding to database entity is as below.
   ```
    type ArticleRole struct {
      UserID    string `json:"user_id" bson:"user_id"`
      CompanyId string `json:"company_id" bson:"company_id"`
      ArticleId string `json:"article_id" bson:"article_id"`
      Role      string `json:"role" bson:"role"`
   }
   ```
#### 6. **company**:- This is for storing details of all companies.
#### Struct corresponding to database entity is as below.
   ```
   type Company struct {
      ID   primitive.ObjectID `json:"_id" bson:"_id"`
      Name string             `json:"name" bson:"name"`
   }
   ```

### Initial Data in database 
#### we are handling all below requirements through code itself.

<br/>1. superadmin user details for each company should need to insert initially.
<br/>2. Also, same user with superadmin role need to be added in *user_role** collection.
<br/>3. Also, all companies details need to be insert in **company** collection.
<br/>4. Also, predefind role need to add in **role** collection.


## APIs List

### 1. Login in Company
```
   POST: localhost:8080/login
```
#### 2. To add user in user collectiob
```
   POST: localhost:8080/user
```
#### 3. To add role for user in particular company
```
   POST: localhost:8080/company/{company_id}/user/{user_id}/role
```
#### ARTICLE RELATED APIs
#### 4. To get all article in company.
```
   GET: localhost:8080/company/{company_id}/article
```
#### 5. To get single article in company.
```
   GET: localhost:8080/company/{company_id}/article/{article_id}
```
#### 6. To create article for company.
```
   POST: localhost:8080/company/{company_id}/article
```
#### 7. To update article in company.
```
   PUT: localhost:8080/company/{company_id}/article/{article_id}
```
#### 8. To delete article in company.
```
   DELELTE: localhost:8080/article/{article_id}
```
### ROLE CHANGE APIs

#### 9. To change role of user in company :- Only superadmin can change role of other user.<br/>
```
   PUT: localhost:8080/company/{company_id}/user/{email}/role
```
#### 10. To change role of user on particular article :- Only superadmin can change role of other user on particular article.<br/>
```
   PUT: localhost:8080/company/{company_id}/user/{email}/article/{article_id}/role
```

## APIs

### 1. User login

**API:** **localhost:8080/login**<br/>
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
### 2. Add User
**API:** **localhost:8080/user**<br/>
**Method:** POST<br/>
**Description:** <br/>1. Only superadmin of all company can add user in **user** collection, then that user can be add in multiple company with specific api.

<br/> 2.Here user **email** is unique, duplicate entry with same email is not allowed.<br/>

**Payload**:
```
   {
      "first_name":<firstname>,
      "last_name": <lastname>,
      "email": <email>
      "password" : <password>
   }
```
**Response:** (for reference only)
```
   {
      "message": "User with email: shubham@gmail.com is added to database  with id :- 60ed8b906f708a84e5dac774"
   }
```

### 3. To add role for user in  particular company
**API:** **localhost:8080/company/{company_id}/user/{user_id}/user/role**<br/>
**Method:** POST<br/>
**Description:** <br/>1. Only superadmin have permission to access this api. They can add user from **user** collection into **user_role** collection,if that user not already present in **user_role** collection for same company.
 
**Payload**:
```
   {
       "role": "member"
   }
```
**Response:** (for reference only)
```
   {
      "message": "User with id :60ed8b906f708a84e5dac774 is added to comapny with company_id: 60ebe75e02bcbdc4d7ae5b44 with role:member"
   }
```
#### 4. GET all articles
**API:** **localhost:8080/company/{company_id}/article**<br/>
**Method:** GET<br/>
**Description:** <br/>To Get all articles in provided company, only admin, member and superadmin can see all articles.<br/>
**Response:** (for reference only)
```
   [
      {
         "_id": "60ebc67056152e4ab5c6a5f7",
         "company_id": "60ebc51456152e4ab5c6a5e2",
         "body": "Welcome to GSLAB family!!!"
      },
      {
         "_id": "60ebc67756152e4ab5c6a5fa",
         "company_id": "60ebc51456152e4ab5c6a5e2",
         "body": "Blockchain is future!"
      }
   ]
```
#### 5. GET single article 
**API:** **localhost:8080/company/{company_id}/article/{article_id}**<br/>
**Method:** GET<br/>
**Description:** <br/>To Get single article by it's **article_id** in provided company, only admin, member and superadmin can read  article.<br/>
**Response:** (for reference only)
```
  {
        "_id": "60ebc67056152e4ab5c6a5f7",
        "company_id": "60ebc51456152e4ab5c6a5e2",
        "body": "Hello Teams"
  }
```
#### 6. Create article for company
**API:** **localhost:8080/company/{company_id}/article**<br/>
**Method:** POST<br/>
**Description:**
<br/> Only superadmin can add article in company.<br/>

**Payload**:
```
   {
    "body": "Welcome to Kpoint!!"
   }
```
**Response:** (for reference only)
```
   {
       "message": "Article with article id: 60ed8cdc6f708a84e5dac790 is added to company having id: 60ebe75e02bcbdc4d7ae5b44 "
   }
```

#### 7. Update article by articleid
**API:** **localhost:8080/company/{company_id}/article/{article_id}**<br/>
**Method:** PUT<br/>
**Description:**
<br/>1. **superadmin,admin** and **member** have access to this api, checked by **Casbin**.<br/>
<br/>2. User has specific access to the articles apart from there role, which is verified internally. Roles related to articles are stored in **article_role** collection.<br/>

**Payload**:
```
   {
     "body": "Updated article"
   }
```
**Response:**
```
  {
    "message": "Article with id: 60ebc67756152e4ab5c6a5fa is successfully updated!"
  }
```

#### 8. Delete article by articleid
**API:** **localhost:8080/article/{article_id}**<br/>
**Method:** DELETE<br/>
**Description:**
<br/>1. superadmin,admin and member have access to this api, checked by "Cashbin".<br/>
<br/>2. **superadmin** and  **admin** access on article level has given permission to delete the article.<br/>
<br/>3. After deleting the article all entries related to that article in **article_role** collection will be deleted<br/>

**Response:** (for reference only)
```
   {
      "message": "Article with id: ObjectID("60ed8ccb6f708a84e5dac78d") is successfully deleted!"
   }
```

#### 9. Change company role of user.
**API:** **localhost:8080/company/{company_id}/user/{email}/role**<br/>
**Method:** PUT<br/>
**Description:**
<br/>1. Only superadmin have access to this api, and it's checked by **Casbin**.<br/>
<br/>2. After changing role of user, if user have same access on particular article  then in article_role collecion entry with same role will be deleted.Now only entry other than than company role is present in **article_role collection**.<br/>
**Payload**:
```
   {
      "role": "admin"
   }
```
**Response:**
```
   {
      "message": "Role for user with id :60ed8b906f708a84e5dac774  is changed to: admin"
   }
```

#### 10. Change role of user on particular article.
**API:** ***localhost:8080/company/{company_id}/user/{email}/role**<br/>
**Method:** PUT<br/>
**Description:**
<br/> Only superadmin have access to this api, and it's checked by **Casbin** and can change special role of user on particular article that will upated in **article_role** collction if entry already present or it will be added to article role collection. <br/>

**Payload**:
```
   {
      "role" : "member"
   }
```
**Response:** (for reference only)
```
   {
      "message": "Role for user with email:shubham@gmail.com for articleid: 60ebc67056152e4ab5c6a5f7 is changed to: member"
   }
```
