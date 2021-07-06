# Spec Center

#Requirement
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
####Once user is logged inside company, comapanyid is taken from  login generated token.
####following api is works for same logged in company.

### Add User
**API:** localhost:8080/adduser/{role}<br/>
**Method:** POST<br/>
**Description:** Only superadmin can add user in company, by giving role for user in header parameter<br/>
**Payload**:
```
{
    "id" : <id>,
    "firstname":<firstname>,
    "lastname": <lastname>,
    "email": <email>
    "password" : <password>
}
```
**Response:**
```
```

#### GET all articles
**API:** localhost:8080/all_articles<br/>
**Method:** GET<br/>
**Description:** To Get all articles for login company, only admin, member and superadmin can see all articles<br/>
**Response:**
```
```

#### Delete article by articleid
**API:** localhost:8080/article<br/>
**Method:** DELETE<br/>
**Params:**
<br/>Key : articleid, Value: int <br/>
**Description:**<br/> 1. superadmin,admin and member have access to this api, checked by "Cashbin"<br/>
<br/>2. After accesing the api, if user having admin or superadmin access on particular article that is checked by mongo articlerole collection, then only user allow to delete that article.
<br/>3. After deleting the article all entries related to that articleid in articlerole collection will be deleted<br/>
**Response:**
#### for articleid = 1
```
{
    "message": "Article with id: 1 is successfully deleted!"
}
```
### Get TODO by ID
#### PART-1
API response when provided todo id exists in database.<br/>
**API:** localhost:8080/api/v1/todo/2<br/>
**Method:** GET<br/>
**Response:**
```
{
    "id": "2",
    "value": "Todo Number one"
}
```

#### PART-2
API response when provided todo id does not exist in database.<br/>
**API:** localhost:8080/api/v1/todo/2<br/>
**Method:** GET<br/>
**Response:**
```
Record not found with given ID=20
```

### Delete TODO by ID
#### PART-1
API response when provided todo id exists in database.<br/>
**API:** localhost:8080/api/v1/todo/2<br/>
**Method:** DELETE<br/>
**Response:**
```
{
    "msg": "TODO deleted successfully"
}
```

#### PART-2
API response when provided todo id does not exist in database.<br/>
**API:** localhost:8080/api/v1/todo/2<br/>
**Method:** GET<br/>
**Response:**
```
Record not found with given ID=20
```
### Update TODO by ID
#### PART-1
API response when provided todo id exists in database.<br/>
**API:** localhost:8080/api/v1/todo/2<br/>
**Method:** PUT<br/>
**Response:**
```
{
    "msg": "TODO updated successfully"
}
```

#### PART-2
API response when provided todo id does not exist in database.<br/>
**API:** localhost:8080/api/v1/todo/2<br/>
**Method:** PUT<br/>
**Response:**
```
Record not found with given ID=20
```
