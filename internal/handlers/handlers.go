package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	config "github.com/raindongz/booking-system/internal/configs"
	"github.com/raindongz/booking-system/internal/models"
	"github.com/raindongz/booking-system/internal/render"
)

//used by handlers
var Repo *Repository

//repository is the repository type
type Repository struct{
	App *config.AppConfig
}

//reate new repository
func NewRepo(a *config.AppConfig) *Repository{
	return &Repository{
		App: a,
	}
}

//New Handlers set the repository for the handlers
func NewHandlers(r *Repository){
	Repo = r
}

func (m *Repository) Home(w http.ResponseWriter, r *http.Request){
	remoteIP := r.RemoteAddr
	Repo.App.Session.Put(r.Context(), "remote_ip", remoteIP)

	//fmt.Fprintf(w, "this is home page")
	render.RenderTemplate(w, "home.page.tmpl", r, &models.TemplateData{})
}

func (m *Repository) About(w http.ResponseWriter, r *http.Request){
	stringMap := make(map[string]string)
	stringMap["test"] = "Hello again"

	remoteIP := m.App.Session.GetString(r.Context(), "remote_ip")
	stringMap["remote_ip"] = remoteIP

	//fmt.Fprintf(w, fmt.Sprintf("the added value is %d", sum))
	render.RenderTemplate(w, "about.page.tmpl", r, &models.TemplateData{
		StringMap: stringMap,
	})
}

func (m *Repository) Reservation(w http.ResponseWriter, r *http.Request){
	render.RenderTemplate(w, "make-reservation.page.tmpl", r, &models.TemplateData{})
}

func (m *Repository) General(w http.ResponseWriter, r *http.Request){
	render.RenderTemplate(w, "general.page.tmpl", r, &models.TemplateData{})
}
func (m *Repository) Major(w http.ResponseWriter, r *http.Request){
	render.RenderTemplate(w, "majors.page.tmpl", r, &models.TemplateData{})
}

func (m *Repository) Book(w http.ResponseWriter, r *http.Request){
	render.RenderTemplate(w, "book.page.tmpl", r, &models.TemplateData{})
}
func (m *Repository) PostBook(w http.ResponseWriter, r *http.Request){
	start := r.Form.Get("start")
	end := r.Form.Get("end")

	w.Write([]byte(fmt.Sprintf("start date is %s and end date is %s", start, end)))
}

type JsonResponse struct{
	OK bool `json:"ok"`
	Message string `json:"message"`
}
func (m *Repository) PostBookJson(w http.ResponseWriter, r *http.Request){
	response := JsonResponse{
		OK : false,
		Message: "Availability",
	}

	out, err := json.MarshalIndent(response, "", "     ")
	if err != nil {
		log.Println(err)
	}
	log.Println(string(out))
	w.Header().Set("Content-Type", "application/json")
	w.Write(out)
}

func (m *Repository) Contact(w http.ResponseWriter, r *http.Request){
	render.RenderTemplate(w, "contact.page.tmpl", r, &models.TemplateData{})
}