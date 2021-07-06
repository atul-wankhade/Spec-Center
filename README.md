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

### Mongo
#### Database 
Name :- SPEC-CENTER <br/>
#### Collections
1. user
2. role
3. article
4. articlerole

#### Initial Data in database
1.Superadmin details for add company should need to insert initilly
<br/>2. Also, there role in role collection need to be added.
## APIs List

#### LOGIN & USER ADD
POST: localhost:8080/login/{companyid} <br/>
POST: localhost:8080/adduser <br/>

#### ARTICLE RELATED APIs
GET: localhost:8080/all_articles<br/>
GET: localhost:8080/article<br/>
PUT: localhost:8080/article<br/>
DELELTE: localhost:8080/article<br/>

#### ROLE CHANGE :- Only superadmin can change role of other user.<br/>
PUT: localhost:8080/articlerole/{articleid} <br/>
PUT: localhost:8080/role <br/>

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
<br/> 2. Once user is added in company, same time in role collection we are updating role of that user in company.
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
**Description:** To Get all articles for login company, only admin, member and superadmin can see all articles<br/>
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

#### Delete article by articleid
**API:** localhost:8080/article<br/>
**Method:** DELETE<br/>
**Params:**
<br/>Key : articleid, Value: int <br/>
**Description:**
<br/>1. superadmin,admin and member have access to this api, checked by "Cashbin".
<br/>2. After accesing the api, if user having admin or superadmin access on particular article that is checked by mongo articlerole collection, then only user allow to delete that article.
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
<br/>1. superadmin,admin and member have access to this api, checked by "Cashbin".
<br/>2. After accesing the api, if user having admin or superadmin access on particular article that is checked by mongo articlerole collection, then only user allow to update that article.
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

