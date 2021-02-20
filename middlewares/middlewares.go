package middlewares

import(
	"log"
	"net/http"
	"html/template"
)

//template folder
var (
	templ = template.Must(template.ParseGlob("../templates/*.html"))
)

func Home(w http.ResponseWriter,r *http.Request){
	//render template
	err := templ.ExecuteTemplate(w,"index.html",nil)
	Check(err)
}

/* log errors */
func Check(err error){
	if err != nil{
		log.Fatalln(err)
	}
}
