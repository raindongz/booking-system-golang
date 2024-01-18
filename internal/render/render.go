package render

import (
	"bytes"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"path/filepath"
	"time"

	"github.com/justinas/nosurf"
	config "github.com/raindongz/booking-system/internal/configs"
	"github.com/raindongz/booking-system/internal/models"
)

//var tempMap = make(map[string]*template.Template)

var functions = template.FuncMap{
	"humanDate": HumanDate,
	"formatDate": FormatDate,
	"iterate":Iterate,
	"add": Add,
}

var app *config.AppConfig
var pathToTemplate = "./templates"

func Add(a, b int)int{
	return a + b
}

//returns slice of ints starting at 1 going to count 
func Iterate(count int)[]int{
	var i int
	var items []int
	for i = 0; i < count; i++ {
		items = append(items, i)
	}
	return items
}

func NewRenderer(a *config.AppConfig){
	app = a
}


func HumanDate(t time.Time) string{
	return t.Format("2006-01-02")
}

func FormatDate(t time.Time, f string)string{
	return t.Format(f)
}

func AddDefaultData(td *models.TemplateData, r *http.Request) *models.TemplateData{
	td.Flash = app.Session.PopString(r.Context(), "flash")
	td.Error = app.Session.PopString(r.Context(), "error")
	td.Warning = app.Session.PopString(r.Context(), "warning")
	td.CSRFToken = nosurf.Token(r)	
	if app.Session.Exists(r.Context(), "user_id"){
		td.IsAuthenticated =1
	}
	return td
}

func Template(w http.ResponseWriter, tmpl string, r *http.Request, td *models.TemplateData){
	var tc map[string]*template.Template
	if app.UseCache{
		//get the template cache from the app config
		tc = app.TemplateCache
	}else{
		tc, _ = CreateTemplateCache()
	}


	//get requested from cache
	temp, exist := tc[tmpl]
	if !exist{
		log.Fatal("could not create template from template cache")
	}

	buf := new (bytes.Buffer)
	td = AddDefaultData(td, r)
	err := temp.Execute(buf, td)
	if err != nil {
		log.Println(err)
	}

	//render the template
	_, err = buf.WriteTo(w)
	if err != nil {
		log.Println(err)
	}


	// parsedTemplate, _ := template.ParseFiles("./templates/" + tmpl, "./templates/base.layout.tmpl")
	// err := parsedTemplate.Execute(w, nil)
	// if err != nil{
	// 	fmt.Println("error parsing template", err)
	// 	return
	// }
}


func CreateTemplateCache()(map[string]*template.Template, error){
	myCache := map[string]*template.Template{}
	//get all the files named *.page.tmpl from ./template
	pages, err := filepath.Glob(fmt.Sprintf("%s/*.page.tmpl", pathToTemplate))
	if err != nil {
		return myCache, err
	}

	//range through all the files ending with 
	for _, page := range pages{
		name := filepath.Base(page)
		ts, err := template.New(name).Funcs(functions).ParseFiles(page)
		if err != nil {
			return myCache, err
		}

		matches, err := filepath.Glob(fmt.Sprintf("%s/*.layout.tmpl", pathToTemplate))
		if err !=nil {
			return myCache, err
		}

		if len(matches) > 0{
			ts, err = ts.ParseGlob(fmt.Sprintf("%s/*.layout.tmpl", pathToTemplate))
			if err !=nil {
				return myCache, err
			}
		}
		myCache[name] = ts
	}
	return myCache, nil
}

// func RenderTemplate(w http.ResponseWriter, t string){
// 	var tmpl *template.Template	
// 	var err error

// 	_, checkTmp := tempMap[t]
// 	if !checkTmp {
// 		//create template		
// 		log.Println("didn't found in" )
// 		err = createTmplCache(t)
// 		if err != nil {
// 			log.Println("error creating template")
// 		}
// 	}else{
// 		//exist in the cache
// 		log.Println("found in cache")
// 	}

// 	tmpl = tempMap[t]
// 	err = tmpl.Execute(w, nil)
// }

// func createTmplCache(t string) error{
// 	tmplPath := []string{fmt.Sprintf("./templates/%s", t), "./templates/base.layout.tmpl"}
// 	tmpl, err := template.ParseFiles(tmplPath...)

// 	if err != nil {
// 		return err
// 	}
// 	tempMap[t] = tmpl
// 	return nil
// }