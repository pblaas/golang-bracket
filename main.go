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
	"strings"
	"time"
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

type ScorecardID struct {
	Id               bson.ObjectId `bson:"_id"`
	Serie            string        `bson:"serie"`
	Scorecard        int           `bson:"scorecard"`
	Bracketgame      string        `bson:"bracketgame"`
	Bracketname      string        `bson:"bracketname"`
	Bracketorganizer string        `bson:"bracketorganizer"`
	Bracketapi       string        `bson:"bracketapi"`
	Raffle           int           `bson:"raffle"`
	completed        int           `bson:"completed"`
	Players          [][]string    `bson:"players"`
	Results          [][][]int     `bson:"results"`
}

type SignupForm struct {
	ID            bson.ObjectId `bson:"_id,omitempty"`
	Playername    string        `form:"Username"`
	Email         string        `form:"Email"`
	Signupaddress string
	Signupdate    time.Time
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

func SignUp(rw http.ResponseWriter, r *http.Request) {
	fp := path.Join("templates", "signup.html")
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

func CreateBracket(rw http.ResponseWriter, r *http.Request) {
	fp := path.Join("templates", "createbracket.html")
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

func GracketShowIndex(w http.ResponseWriter, r *http.Request) {
	fp := path.Join("templates", "gracketindex.html")
	tmpl, err := template.ParseFiles(fp)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if err := tmpl.Execute(w, ""); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func GracketShowHandler(w http.ResponseWriter, r *http.Request) {
	fp := path.Join("templates", "gracketmatch.html")
	tmpl, err := template.ParseFiles(fp)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if err := tmpl.Execute(w, ""); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func SignupSubmit(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	session, err := mgo.Dial("localhost")
	if err != nil {
		panic(err)
	}
	defer session.Close()

	//Optional. Switch the session to monotonic behavior
	session.SetMode(mgo.Monotonic, true)

	c := session.DB("spdb").C("match")
	if r.FormValue("Email") != "" && r.FormValue("Username") != "" {

		s := strings.Split(r.RemoteAddr, ":")
		ip := s[0]
		entry := &SignupForm{
			Playername:    r.FormValue("Username"),
			Email:         r.FormValue("Email"),
			Signupaddress: ip,
			Signupdate:    time.Now(),
		}
		err = c.Insert(entry)
		if err != nil {
			log.Fatal(err)

		}
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
	}

	http.Redirect(w, r, "/signup", http.StatusTemporaryRedirect)
}

func BracketShowHandler(w http.ResponseWriter, r *http.Request) {
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

	query := c.Find(bson.M{"serie": serie, "scorecard": 1})
	var scorecardid []ScorecardID
	if err := query.All(&scorecardid); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	fp := path.Join("templates", "match.html")
	tmpl, err := template.ParseFiles(fp)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if err := tmpl.Execute(w, scorecardid); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
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
	r.HandleFunc("/signup", SignUp)
	r.HandleFunc("/createbracket", CreateBracket)
	r.HandleFunc("/SignupSubmit", SignupSubmit)
	r.HandleFunc("/", Bracket)
	r.HandleFunc("/gracket", GracketShowIndex)
	r.HandleFunc("/gracket/{serie}", GracketShowHandler)

	log.Println("Server started...")
	http.ListenAndServe(":3099", r)
}
