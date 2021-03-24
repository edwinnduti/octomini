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
	"strings"
	"html/template"
	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
	"net/http"
	"os"
	"fmt"
	"time"
)

// templ
var (
	dir = "assets/"
	templ = template.Must(template.ParseGlob("templates/*.html"))
)

// Member type and all offerings
type Member struct{
	Id		primitive.ObjectID	`bson:"_id"  json:id"`
	Name		string			`json:"name"`
	AllOffering	[]TodaysOffering	`json:"allOfferings"`
	Total		int			`json:"total"`
}

// one day offering to store a map of date and offering
type TodaysOffering struct{
	Date		string	`json:"date"`
	Offering	int	`json:"todaysOffering"`
}


var templates map[string]*template.Template
//Compile view templates
func init() {
	if templates == nil {
		templates = make(map[string]*template.Template)
	}
	templates["index"] = template.Must(template.ParseFiles("templates/index.html","templates/base.html"))
	templates["addMember"] = template.Must(template.ParseFiles("templates/addMember.html","templates/base.html"))
	templates["profilePage"] = template.Must(template.ParseFiles("templates/profilePage.html","templates/base.html"))
	templates["editMember"] = template.Must(template.ParseFiles("templates/editMember.html","templates/base.html"))
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
	member.Id = primitive.NewObjectID()

	if r.Method == "POST"{
		r.ParseForm()
		member.Name = r.FormValue("name")

		var today TodaysOffering

		todaysOffering,err := strconv.Atoi(r.FormValue("offering"))
		Checkf("string not converted",err)

		today.Date = time.Now().Format("02-01-2006")
		today.Offering = todaysOffering

		member.AllOffering = append(member.AllOffering,today)
		member.Total = todaysOffering
		result, err := c.InsertOne(ctx, member)
		Check(err)
		fmt.Println("added new object of Id ",result.InsertedID.(primitive.ObjectID))

		// set headers
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Method", "POST")

		http.Redirect(w, r, "/", http.StatusFound)
	}
	if r.Method == "GET"{
		// set headers
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Method","GET")
		w.WriteHeader(http.StatusOK)
		RenderTemp(w,"addMember","base",nil)
	}
}

/* form view */
func MemberForm(w http.ResponseWriter,r *http.Request){
	// set headers
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Method", "GET")
	w.WriteHeader(http.StatusOK)

	//render template
	RenderTemp(w,"addMember","base",nil)
}

/* show member profile */
func MemberProfile(w http.ResponseWriter,r *http.Request){
	// get tableid
	vars := mux.Vars(r)
	id := vars["userid"]
	id = Between(id,"ObjectID(\"","\")")
	userid,err := primitive.ObjectIDFromHex(id)
	Check(err)

	var user Member

	// create connection
	client, err := CreateConnection()
	Check(err)

	// select db and collection
        cl := client.Database(database).Collection(collection)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)

	defer cancel()                                                                                              /*  USER DOC */
	// find table document
	err = cl.FindOne(ctx, bson.M{"_id": userid}).Decode(&user)
	Check(err)

	// set headers
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Method", "GET")
	w.WriteHeader(http.StatusOK)

	//render template
	RenderTemp(w,"profilePage","base",user)
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

// update form 
func UpdateForm(w http.ResponseWriter,r *http.Request){
	// get tableid
	vars := mux.Vars(r)
	id := vars["userid"]
	id = Between(id,"ObjectID(\"","\")")
	userid,err := primitive.ObjectIDFromHex(id)
	Check(err)

	var user Member
	// create connection
	client, err := CreateConnection()
	Check(err)

	// select db and collection
	cl := client.Database(database).Collection(collection)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)

	defer cancel()

	/*  USER DOC */
	// find table document
	err = cl.FindOne(ctx, bson.M{"_id": userid}).Decode(&user)
	Check(err)

	// set headers
	w.Header().Set("Access-Control-Allow-Origin","*")
	w.Header().Set("Access-Control-Allow-Method","GET")
	w.WriteHeader(http.StatusOK)

	//render template
	RenderTemp(w,"editMember","base",user)
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

func Checkf(line string,err error){
	if err != nil{
		log.Fatalln(line," : ",err)
	}
}

// check whats between
func Between(value string, a string, b string) string {
    // Get substring between two strings.
    posFirst := strings.Index(value, a)
    if posFirst == -1 {
	    return ""
    }
    posLast := strings.Index(value, b)
    if posLast == -1 {
        return ""
    }
    posFirstAdjusted := posFirst + len(a)
    if posFirstAdjusted >= posLast {
        return ""
    }
    return value[posFirstAdjusted:posLast]
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
	r.HandleFunc("/{userid}",MemberProfile).Methods("GET","OPTIONS")
	r.HandleFunc("{userid}/edit",MemberProfile).Methods("GET","OPTIONS")
	//r.HandleFunc("/update/{userid}",MemberProfile).Methods("PUT","OPTIONS")
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
