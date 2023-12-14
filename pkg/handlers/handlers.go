package handlers

import (
	"net/http"

	config "github.com/raindongz/booking-system/pkg/configs"
	"github.com/raindongz/booking-system/pkg/models"
	"github.com/raindongz/booking-system/pkg/render"
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
	render.RenderTemplate(w, "home.page.tmpl", &models.TemplateData{})
}

func (m *Repository) About(w http.ResponseWriter, r *http.Request){
	stringMap := make(map[string]string)
	stringMap["test"] = "Hello again"

	remoteIP := m.App.Session.GetString(r.Context(), "remote_ip")
	stringMap["remote_ip"] = remoteIP

	//fmt.Fprintf(w, fmt.Sprintf("the added value is %d", sum))
	render.RenderTemplate(w, "about.page.tmpl", &models.TemplateData{
		StringMap: stringMap,
	})
}