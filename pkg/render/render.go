package render

import (
	"bytes"
	"log"
	"net/http"
	"path/filepath"
	"text/template"

	config "github.com/raindongz/booking-system/pkg/configs"
	"github.com/raindongz/booking-system/pkg/models"
)

//var tempMap = make(map[string]*template.Template)


var app config.AppConfig
func NewTemplate(a config.AppConfig){
	app = a
}

func AddDefaultData(td *models.TemplateData) *models.TemplateData{
	
	return td
}

func RenderTemplate(w http.ResponseWriter, tmpl string, td *models.TemplateData){
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
	td = AddDefaultData(td)
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
	pages, err := filepath.Glob("./templates/*.page.tmpl")
	if err != nil {
		return myCache, err
	}

	//range through all the files ending with 
	for _, page := range pages{
		name := filepath.Base(page)
		ts, err := template.New(name).ParseFiles(page)
		if err != nil {
			return myCache, err
		}

		matches, err := filepath.Glob("./templates/*.layout.tmpl")
		if err !=nil {
			return myCache, err
		}

		if len(matches) > 0{
			ts, err = ts.ParseGlob("./templates/*.layout.tmpl")
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