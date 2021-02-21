/*
[*] Copyright © 2020
[*] Dev/Author ->  Edwin Nduti
[*] Description:
	The code stores names in a mysql file.
    Written in pure Golang.
 */

package main

// libraries to use
import (
	"github.com/urfave/negroni"
	"context"
	"encoding/json"
	"html/template"
	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
	"net/http"
	"os"
	"time"
)

// templ
var (
	dir = "assets/"
	templ = template.Must(template.ParseGlob("templates/*.html"))
)

// Member type
type Member struct{
	Name		string	`json:"name"`
	Offering	string	`json:"offering"`
}

// database and collection names are statically declared
const database, collection = "pceanyaga", "offeringManagement"

// create connection to mongodb
func CreateConnection() (*mongo.Client, error) {
	// connect to mongodb
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// set MONGOURI
	MongoURI := os.Getenv("MONGOURI")
	// connect to mongodb
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(
		MongoURI,
	))
	Check(err)

	// return client and error
	return client, nil
}

func Home(w http.ResponseWriter,r *http.Request){
	//render template
	err := templ.ExecuteTemplate(w,"index.html",nil)
	Check(err)
}

/* GET all  data */
func GetAllHandler(w http.ResponseWriter,r *http.Request){
        var members []Member

        // create connection
        client, err := CreateConnection()
	Check(err)

	// select db and collection
	cHotel := client.Database(database).Collection(collection)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// get all documents
	cursor,err := cHotel.Find(ctx, bson.M{})
	Check(err)

	err = cursor.All(ctx,&members)
	Check(err)

	// set headers
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Method", "GET")
	w.WriteHeader(http.StatusOK)

	//render template
	err = templ.ExecuteTemplate(w,"index.html",members)
}

/* log errors */
func Check(err error){
	if err != nil{
		log.Fatalln(err)
	}
}


// Main function
func main() {
	/*
	   mgo.SetDebug(true)
	   mgo.SetLogger(log.New(os.Stdout,"err",6))

	   The above two lines are for debugging errors
	   that occur straight from accessing the mongo db
	*/

	//Register router{}
	r := mux.NewRouter().StrictSlash(false)

	// API routes,handlers and methods
	r.HandleFunc("/",GetAllHandler).Methods("GET","OPTIONS")
	r.PathPrefix("/assets/").Handler(http.StripPrefix("/assets/", http.FileServer(http.Dir(dir))))

	//Get port
	Port := os.Getenv("PORT")
	if Port == "" {
		Port = "8080"
	}

	// establish logger
	n := negroni.Classic()
	n.UseHandler(r)
	server := &http.Server{
		Handler: n,
		Addr   : ":"+Port,
	}
	log.Printf("Listening on PORT: %s",Port)
	server.ListenAndServe()
}
