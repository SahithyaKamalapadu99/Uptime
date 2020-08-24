Uptime Monitor

    This system monitors the uptime of services. It accepts URLs along with parameters.

Setup :

Clone the "uptime" folder into your machine. Change the paths accordingly in the import statement.

Install go Environment on your machine.

Install the necessary packages provided in go.mod file


Install MySql on your machine and create a database named urls (preferably ,if not change the name in db.go file) and mention the username and password in db.go file in CreateConnection() function

About the application :

1.run the command : go run main.go 

2.Application creates a database and allows you to store urls and parameters 

3.It allows you to perform various requests  like GET,POST,PATCH,DELETE,ACTIVATE and DEACTIVATE

4. Following requests can be performed 


POST /urls/                             allows to add new url record into database along with the parameters

GET /urls/:id                           retrieves the url details and status of url with "id " if present in database  

DELETE /urls/:id                        deletes the url with id as "id" record from database                  

PATCH /urls/:id                         allows to update the details of respective "id" if url present in the database 

POST /urls/:id/activate                 activates the url with requested id ,if it is inactive 

POST /urls/:id/deactivate               Deactivates the url with requested id and stops crawling the site.

