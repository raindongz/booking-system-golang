package forms

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

//create a custon form struct, embeds a url.Value object
type Form struct{
	url.Values
	Errors errors
}


//initialize a form struct
func New(data url.Values) *Form{
	return &Form{
		data,
		errors(map[string][]string{}),
	}
}

func (f *Form) Required(fields ...string){
	for _, field := range fields{
		value := f.Get(field)
		if strings.TrimSpace(value) == "" {
			f.Errors.Add(field, "this field can not be blank")
		}
	}
}

//checks if form field is in post and not empty
func (f *Form) Has(field string, r *http.Request) bool{
	x := r.Form.Get(field)
	if x== ""{
		f.Errors.Add(field, "this field cannot be blank")
		return false
	}
	return true
}

//valid return true if there are no errors
func (f *Form) Valid() bool{
	return len(f.Errors) == 0
}


func (f *Form) MinLength(field string, length int, r *http.Request) bool{
	x := r.Form.Get(field)
	if len(x) < length {
		f.Errors.Add(field, fmt.Sprintf("This field must be at least %d characters long", length))
		return false
	} 
	return true
}