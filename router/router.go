/*
++ The router package
++ Major routes are traces here.
++ Only routes are both GET methods.
*/

package router

import(
	"net/http"
	"github.com/gorilla/mux"
	"github.com/edwinnduti/octomini/middlewares"
)

/* store images,css,etc */
var dir = "assets/"

func Router() *mux.Router {
	//Register router
	r := mux.NewRouter().StrictSlash(false)

	// API routes,handlers and methods
	r.HandleFunc("/",middlewares.Home).Methods("GET","OPTIONS")
	r.PathPrefix("/assets/").Handler(http.StripPrefix("/assets/", http.FileServer(http.Dir(dir))))

	// return router
	return r
}
