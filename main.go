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
	"strconv"
	"html/template"
	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/bson/primitive"
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
	Id           primitive.ObjectID `bson:"_id"  json:id"`
	Name		string	`json:"name"`
	Offering	int	`json:"offering"`
}

var templates map[string]*template.Template
//Compile view templates
func init() {
	if templates == nil {
		templates = make(map[string]*template.Template)
	}
	templates["index"] = template.Must(template.ParseFiles("templates/index.html","templates/base.html"))
	templates["addMember"] = template.Must(template.ParseFiles("templates/addMember.html","templates/base.html"))
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

/* save new members */
func PostSaveMember(w http.ResponseWriter, r *http.Request) {
        var member Member

	client, err := CreateConnection()
	Check(err)

	c := client.Database(database).Collection(collection)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	//create the new member
	r.ParseForm()
        member.Id = primitive.NewObjectID()
	member.Name = r.PostFormValue("name")
	offering := r.FormValue("offering")
	member.Offering,err = strconv.Atoi(offering)
	Check(err)
	_, err = c.InsertOne(ctx, member)
	Check(err)

	// set headers
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Method", "GET")
	w.WriteHeader(http.StatusCreated)

	http.Redirect(w, r, "/", 302)
}

/* form view */
func MemberForm(w http.ResponseWriter,r *http.Request){
	//render template
	RenderTemp(w,"addMember","base",nil)
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
	RenderTemp(w,"index","base",members)
}

/* function render template */
//Render templates for the given name, template definition and data object
func RenderTemp(w http.ResponseWriter, name string, template string, viewModel interface{}) {
	// Ensure the template exists in the map.
	tmpl, ok := templates[name]
	if !ok {
		http.Error(w, "The template does not exist.", http.StatusInternalServerError)
	}
	err := tmpl.ExecuteTemplate(w, template, viewModel)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
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
	r.HandleFunc("/add",MemberForm).Methods("GET","OPTIONS")
	r.HandleFunc("/save",PostSaveMember).Methods("POST","OPTIONS")
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
