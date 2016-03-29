package main

import (
	"fmt"
	"github.com/gorilla/mux"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"html/template"
	"log"
	"net/http"
	"path"
)

type MatchID struct {
	Id        bson.ObjectId `bson:"_id"`
	Serie     string        `bson:"serie"`
	Round     int           `bson:"round"`
	P1        string        `bson:"p1"`
	P2        string        `bson:"p2"`
	Timestamp string        `bson:"timestamp"`
	P1_score  int           `bson:"p1_score"`
	P2_score  int           `bson:"p2_score"`
}

type hookedResponseWriter struct {
	http.ResponseWriter
	ignore bool
}

func (hrw *hookedResponseWriter) WriteHeader(status int) {
	hrw.ResponseWriter.WriteHeader(status)
	if status == 404 {
		hrw.ignore = true
		// Write custom error here to hrw.ResponseWriter
	}
}

func (hrw *hookedResponseWriter) Write(p []byte) (int, error) {
	if hrw.ignore {
		return len(p), nil
	}
	return hrw.ResponseWriter.Write(p)
}

type NotFoundHook struct {
	h http.Handler
}

func (nfh NotFoundHook) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	nfh.h.ServeHTTP(&hookedResponseWriter{ResponseWriter: w}, r)
}

func Bracket(rw http.ResponseWriter, r *http.Request) {
	fp := path.Join("templates", "index.html")
	tmpl, err := template.ParseFiles(fp)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return
	}
	if err := tmpl.Execute(rw, ""); err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return
	}

}

func BracketShowHandler(rw http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	serie := params["serie"]

	log.Println("serie" + serie)
	session, err := mgo.Dial("localhost")
	if err != nil {
		panic(err)
	}
	defer session.Close()

	//Optional. Switch the session to monotonic behavior
	session.SetMode(mgo.Monotonic, true)

	c := session.DB("spdb").C("match")

	query := c.Find(bson.M{"serie": serie, "round": 1})
	var matchid []MatchID
	if err := query.All(&matchid); err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return
	}
	fp := path.Join("templates", "match.html")
	tmpl, err := template.ParseFiles(fp)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return
	}
	if err := tmpl.Execute(rw, matchid); err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return
	}
}

func MainPage(rw http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(rw, "Basic information on the app.")
}

func main() {
	r := mux.NewRouter()
	r.PathPrefix("/public/").Handler(NotFoundHook{http.StripPrefix("/public/", http.FileServer(http.Dir("public")))})

	r.HandleFunc("/bracket/{serie}", BracketShowHandler)
	r.HandleFunc("/", Bracket)

	log.Println("Server started...")
	http.ListenAndServe(":3099", r)
}
