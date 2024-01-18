package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	config "github.com/raindongz/booking-system/internal/configs"
	mydriver "github.com/raindongz/booking-system/internal/driver"
	"github.com/raindongz/booking-system/internal/forms"
	"github.com/raindongz/booking-system/internal/helpers"
	"github.com/raindongz/booking-system/internal/models"
	"github.com/raindongz/booking-system/internal/render"
	"github.com/raindongz/booking-system/internal/repository"
	"github.com/raindongz/booking-system/internal/repository/dbrepo"
)

//used by handlers
var Repo *Repository

//repository is the repository type
type Repository struct{
	App *config.AppConfig
	DB repository.DatabaseRepo
}

//reate new repository
func NewRepo(a *config.AppConfig, db *mydriver.DB) *Repository{
	return &Repository{
		App: a,
		DB: dbrepo.NewPostgresRepo(db.SQL, a),
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
	render.Template(w, "home.page.tmpl", r, &models.TemplateData{})
}

func (m *Repository) About(w http.ResponseWriter, r *http.Request){
	// stringMap := make(map[string]string)
	// stringMap["test"] = "Hello again"

	// remoteIP := m.App.Session.GetString(r.Context(), "remote_ip")
	// stringMap["remote_ip"] = remoteIP
	//fmt.Fprintf(w, fmt.Sprintf("the added value is %d", sum))
	render.Template(w, "about.page.tmpl", r, &models.TemplateData{})
}

func (m *Repository) Reservation(w http.ResponseWriter, r *http.Request){
	// var emptyReservation models.Reservation
	res, ok := m.App.Session.Get(r.Context(), "reservation").(models.Reservation)
	if !ok {
		helpers.ServerError(w, errors.New("can not get reservation from session"))
		return
	}

	room, err := m.DB.GetRoomByID(res.RoomID)
	if err != nil{
		helpers.ServerError(w, err)
		return
	}
	res.Room.RoomName = room.RoomName


	sd := res.StartDate.Format("2006-01-02")
	ed := res.EndDate.Format("2006-01-02")

	m.App.Session.Put(r.Context(), "reservation", res)

	stringMap := make(map[string]string)
	stringMap["start_date"] = sd
	stringMap["end_date"] = ed

	data := make(map[string]interface{})
	data["reservation"] = res


	render.Template(w, "make-reservation.page.tmpl", r, &models.TemplateData{
		Form: forms.New(nil),
		Data: data,
		StringMap: stringMap,
	})
}

func (m *Repository) PostReservation(w http.ResponseWriter, r *http.Request){
	reservation, ok := m.App.Session.Get(r.Context(), "reservation").(models.Reservation)
	if !ok {
		helpers.ServerError(w, errors.New("can't get from session"))
		return
	}


	err := r.ParseForm()
	// err = errors.New("hello this is an error message ")
	if err != nil{
		// log.Println(err)
		helpers.ServerError(w, err)
		return
	}

	// sd := r.Form.Get("start_date")
	// ed := r.Form.Get("end_date")

	// layout := "2006-01-02"
	// startDate, err := time.Parse(layout, sd)
	// if err != nil {
	// 	helpers.ServerError(w, err)
	// }

	// endDate, err := time.Parse(layout, ed)
	// if err != nil {
	// 	helpers.ServerError(w, err)
	// }

	// roomID, err := strconv.Atoi(r.Form.Get("room_id"))
	// if err != nil {
	// 	helpers.ServerError(w, err)
	// }

	reservation.FirstName =r.Form.Get("first_name") 
	reservation.LastName =r.Form.Get("last_name") 
	reservation.Phone =r.Form.Get("phone") 
	reservation.Email =r.Form.Get("email") 
	// reservation := models.Reservation{
	// 	FirstName: r.Form.Get("first_name"),
	// 	LastName: r.Form.Get("last_name"),
	// 	Phone: r.Form.Get("phone"),
	// 	Email: r.Form.Get("email"),
	// 	StartDate: startDate,
	// 	EndDate: endDate,
	// 	RoomID: roomID,
	// }

	form := forms.New(r.PostForm)

	// form.Has("first_name", r)
	form.Required("first_name", "last_name", "email", "phone")
	form.MinLength("first_name", 3, r)
	//form.IsEmail("email")

	if !form.Valid() {
		data := make(map[string]interface{})
		data["reservation"] = reservation
		render.Template(w, "make-reservation.page.tmpl", r, &models.TemplateData{
			Form: form,
			Data: data,
		})	
		return 
	}

	newReservationID, err := m.DB.InsertReservation(reservation)
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	m.App.Session.Put(r.Context(), "reservation", reservation)

	restriction := models.RoomRestriction{
		StartDate: reservation.StartDate,
		EndDate: reservation.EndDate,
		RoomID: reservation.RoomID,
		ReservationID: newReservationID,
		RestrictionID: 1,
		// Room: models.Room{},
		// Reservation: models.Reservation{},
		// Restriction: models.Restriction{},
	}

	err = m.DB.InsertRoomRestriction(restriction)
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	//send notifications
	htmlMessage := fmt.Sprintf(`
		<strong>Reservation Confirmation</strong><br>
		Dear %s:, <br>
		This is confirm your reservation from %s to %s.
	`, reservation.FirstName, reservation.StartDate.Format("2006-01-02"), reservation.EndDate.Format("2006-01-02"))
	
	msg := models.MailData{
		To: reservation.Email,
		From: "me@here.com",
		Subject: "Reservation Confirmation",
		Content: htmlMessage,
		Template: "basic.html",
	}

	 m.App.MailChan <- msg

	m.App.Session.Put(r.Context(), "reservation", reservation)
	
	//avoid subit form twice
	http.Redirect(w, r, "/reservation-summary", http.StatusSeeOther)
}

func (m *Repository) General(w http.ResponseWriter, r *http.Request){
	render.Template(w, "general.page.tmpl", r, &models.TemplateData{})
}
func (m *Repository) Major(w http.ResponseWriter, r *http.Request){
	render.Template(w, "majors.page.tmpl", r, &models.TemplateData{})
}

func (m *Repository) Book(w http.ResponseWriter, r *http.Request){
	render.Template(w, "book.page.tmpl", r, &models.TemplateData{})
}
func (m *Repository) PostBook(w http.ResponseWriter, r *http.Request){
	start := r.Form.Get("start")
	end := r.Form.Get("end")

	layout := "2006-01-02"
	startDate, err := time.Parse(layout, start)
	if err != nil{
		helpers.ServerError(w, err)
		return
	}
	endDate, err := time.Parse(layout, end)
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	rooms, err := m.DB.SearchAvailabilityForAllRooms(startDate, endDate)
	if err != nil{
		helpers.ServerError(w, err)
		return
	}

	
	//no availability
	if len(rooms) == 0{
		m.App.Session.Put(r.Context(), "error", "No availability")
		http.Redirect(w, r, "/book", http.StatusSeeOther)
		return
	}

	data := make(map[string]interface{})
	data["rooms"] = rooms

	res := models.Reservation{
		StartDate: startDate,
		EndDate: endDate,
	}
	m.App.Session.Put(r.Context(), "reservation", res)

	//w.Write([]byte(fmt.Sprintf("start date is %s and end date is %s", start, end)))
	render.Template(w, "choose-room.page.tmpl", r, &models.TemplateData{
		Data: data,
	})
}

type JsonResponse struct{
	OK bool `json:"ok"`
	Message string `json:"message"`
	RoomID string `json:"room_id"`
	StartDate string `json:"start_date"`
	EndDate string `json:"end_date"`
}
func (m *Repository) PostBookJson(w http.ResponseWriter, r *http.Request){

	sd := r.Form.Get("start")
	ed := r.Form.Get("end")
	layout := "2006-01-02"
	startDate, _ := time.Parse(layout, sd)
	endDate, _ := time.Parse(layout, ed)
	roomID, _ := strconv.Atoi(r.Form.Get("room_id"))

	avaliable, _ := m.DB.SearchAvaliabilityByDatesByRoomID(startDate, endDate, roomID)

	response := JsonResponse{
		OK : avaliable,
		Message: "",
		StartDate: sd,
		EndDate: ed,
		RoomID: strconv.Itoa(roomID),
	}


	out, err := json.MarshalIndent(response, "", "     ")
	if err != nil {
		helpers.ServerError(w, err)
		return
	}
	log.Println(string(out))
	w.Header().Set("Content-Type", "application/json")
	w.Write(out)
}

func (m *Repository) Contact(w http.ResponseWriter, r *http.Request){
	render.Template(w, "contact.page.tmpl", r, &models.TemplateData{})
}

func (m *Repository) ReservationSummary(w http.ResponseWriter, r *http.Request){
	reservation, ok := m.App.Session.Get(r.Context(), "reservation").(models.Reservation)
	if !ok {
		m.App.ErrorLog.Println("Can't get error from session")
		m.App.Session.Put(r.Context(), "error", "Can't get reservation form session")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	m.App.Session.Remove(r.Context(), "reservation")

	data := make(map[string]interface{})
	data["reservation"] = reservation

	sd := reservation.StartDate.Format("2006-01-02")
	ed := reservation.EndDate.Format("2006-01-02")
	stringMap := make(map[string]string)
	stringMap["start_date"] = sd
	stringMap["end_date"] = ed

	render.Template(w, "reservation-summary.page.tmpl", r, &models.TemplateData{
		Data: data,
		StringMap: stringMap,
	})
}

func (m *Repository) ChooseRoom(w http.ResponseWriter, r *http.Request){
	roomID, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil{
		helpers.ServerError(w, err)
		return
	}

	res, ok := m.App.Session.Get(r.Context(), "reservation").(models.Reservation)
	if !ok {
		helpers.ServerError(w, err)
		return
	}

	res.RoomID = roomID
	m.App.Session.Put(r.Context(), "reservation", res)
	http.Redirect(w, r, "/make-reservation", http.StatusSeeOther)
}

func (m *Repository) BookRoom(w http.ResponseWriter, r *http.Request){
	// id, s, e
	roomID, _ := strconv.Atoi(r.URL.Query().Get("id"))
	sd := r.URL.Query().Get("s")
	ed := r.URL.Query().Get("e")

	layout := "2006-01-02"
	startDate, _ := time.Parse(layout, sd)
	endDate, _ := time.Parse(layout, ed)
	
	var res models.Reservation

	room, err := m.DB.GetRoomByID(roomID)
	if err != nil{
		helpers.ServerError(w, err)
		return
	}

	res.Room.RoomName = room.RoomName
	res.RoomID = roomID
	res.StartDate = startDate
	res.EndDate = endDate

	m.App.Session.Put(r.Context(), "reservation", res)

	http.Redirect(w, r, "make-reservation", http.StatusSeeOther)
}

func (m *Repository) ShowLogin(w http.ResponseWriter, r *http.Request){
	render.Template(w, "login.page.tmpl", r, &models.TemplateData{
		Form: forms.New(nil),
	})
}

func (m *Repository) PostShowLogin(w http.ResponseWriter, r *http.Request){
	_ = m.App.Session.RenewToken(r.Context())

	err := r.ParseForm()
	if err != nil {
		log.Println(err)
	}

	email := r.Form.Get("email")
	password := r.Form.Get("password")
	form := forms.New(r.PostForm)
	form.Required("email", "password")
	if !form.Valid(){
		render.Template(w, "login.page.tmpl", r, &models.TemplateData{
			Form: form,
		})
		return
	}

	id, _ , err := m.DB.Authenticate(email, password)
	if err != nil{
		log.Println(err)
		m.App.Session.Put(r.Context(), "error", "Invalid login credentials")
		http.Redirect(w, r, "/user/login", http.StatusSeeOther)
		return
	}

	m.App.Session.Put(r.Context(), "user_id", id)
	m.App.Session.Put(r.Context(), "flash", "login sucessfully")
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (m *Repository) Logout(w http.ResponseWriter, r *http.Request){
	_ = m.App.Session.Destroy(r.Context())
	_ = m.App.Session.RenewToken(r.Context())

	http.Redirect(w, r, "/user/login", http.StatusSeeOther)
}

func (m *Repository) AdminDashboard(w http.ResponseWriter, r *http.Request){
	render.Template(w, "admin-dashboard.page.tmpl", r, &models.TemplateData{})
}

func (m *Repository) AdminNewReservations(w http.ResponseWriter, r *http.Request){
	reservations, err := m.DB.AllNewReservations()
	if err !=nil {
		helpers.ServerError(w, err)
		return
	}

	data := make(map[string]interface{})
	data["reservations"] = reservations

	render.Template(w, "admin-new-reservations.page.tmpl", r, &models.TemplateData{
		Data: data,
	})
}
func (m *Repository) AdminAllReservations(w http.ResponseWriter, r *http.Request){
	reservations, err := m.DB.AllReservations()
	if err !=nil {
		helpers.ServerError(w, err)
		return
	}

	data := make(map[string]interface{})
	data["reservations"] = reservations

	render.Template(w, "admin-all-reservations.page.tmpl", r, &models.TemplateData{
		Data: data,
	})
}

func (m *Repository) AdminShowReservation(w http.ResponseWriter, r *http.Request){
	exploded := strings.Split(r.RequestURI, "/")
	id, err := strconv.Atoi(exploded[4])
	if err != nil {
		helpers.ServerError(w, err)
		return
	}
	
	src := exploded[3]
	stringMap := make(map[string]string)
	stringMap["src"] = src

	year := r.URL.Query().Get("y")
	month := r.URL.Query().Get("m")
	stringMap["month"] = month
	stringMap["year"] = year

	res, err := m.DB.GetReservationByID(id)
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	data := make(map[string]interface{})
	data["reservation"] = res

	render.Template(w, "admin-reservations-show.page.tmpl", r, &models.TemplateData{
		StringMap: stringMap,
		Data: data,
		Form: forms.New(nil),
	})
}

func (m *Repository) AdminPostShowReservation(w http.ResponseWriter, r *http.Request){
	err := r.ParseForm()
	if err !=nil {
		helpers.ServerError(w, err)
		return
	}


	exploded := strings.Split(r.RequestURI, "/")
	id, err := strconv.Atoi(exploded[4])
	if err != nil {
		helpers.ServerError(w, err)
		return
	}
	
	src := exploded[3]
	stringMap := make(map[string]string)
	stringMap["src"] = src	

	res, err := m.DB.GetReservationByID(id)
	if err != nil {
		helpers.ServerError(w, err)
		return
	}	

	res.FirstName = r.Form.Get("first_name")
	res.LastName = r.Form.Get("last_name")
	res.Email = r.Form.Get("email")
	res.Phone = r.Form.Get("phone")

	err = m.DB.UpdateReservation(res)
	if err != nil {
		helpers.ServerError(w, err)
		return	
	}

	month := r.Form.Get("month")
	year := r.Form.Get("year")

	m.App.Session.Put(r.Context(), "flash", "change saved")
	if year == ""{
		http.Redirect(w, r, fmt.Sprintf("/admin/reservations-%s", src), http.StatusSeeOther)
	}else {
		http.Redirect(w, r, fmt.Sprintf("/admin/reservations-calendar?y=%s&m=%s", year, month), http.StatusSeeOther)
	}
}

func (m *Repository) AdminReservationsCalender(w http.ResponseWriter, r *http.Request){
	now := time.Now()

	if r.URL.Query().Get("y") != ""{
		year, _ := strconv.Atoi(r.URL.Query().Get("y"))
		month, _ := strconv.Atoi(r.URL.Query().Get("m"))
		now = time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC)
	}

	data := make(map[string]interface{})
	data["now"] = now

	next := now.AddDate(0, 1, 0)
	last := now.AddDate(0, -1, 0)

	nextMonth := next.Format("01")
	nextMonthYear := next.Format("2006")

	lastMonth := last.Format("01")
	lastMonthYear := last.Format("2006")

	stringMap := make(map[string]string)
	stringMap["next_month"] = nextMonth
	stringMap["next_month_year"] = nextMonthYear
	stringMap["last_month"] = lastMonth
	stringMap["last_month_year"] = lastMonthYear

	stringMap["this_month"] = now.Format("01")
	stringMap["this_month_year"] = now.Format("2006")

	//get first day and last day of month
	currentYear, currentMonth, _ := now.Date()
	currentLocation := now.Location()
	firstOfMonth := time.Date(currentYear, currentMonth, 1, 0, 0, 0, 0, currentLocation)
	lastOfMonth := firstOfMonth.AddDate(0, 1, -1)

	intMap := make(map[string]int)
	intMap["days_in_month"] = lastOfMonth.Day()

	rooms, err := m.DB.AllRooms()

	if err != nil{
		helpers.ServerError(w, err)
		return
	}
	data["rooms"] = rooms
	
	for _, x := range rooms{
		reservationMap := make(map[string]int)
		blockMap := make(map[string]int)

		for d := firstOfMonth; d.After(lastOfMonth) == false; d = d.AddDate(0, 0, 1){
			reservationMap[d.Format("2006-01-2")] = 0
			blockMap[d.Format("2006-01-2")] = 0
		}

		restrictions, err := m.DB.GetRestrictionsForRoomByDate(x.ID, firstOfMonth, lastOfMonth)
		if err !=nil {
			helpers.ServerError(w, err)
			return
		}

		for _, y := range restrictions{
			if y.ReservationID > 0 {
				// it is a reservation
				for d := y.StartDate; d.After(y.EndDate) == false; d = d.AddDate(0, 0, 1){
					reservationMap[d.Format("2006-01-2")] = y.ReservationID
				}
			}else{
				//it is a block
				blockMap[y.StartDate.Format("2006-01-2")] = y.ID
			}
		}

		data[fmt.Sprintf("reservation_map_%d", x.ID)] = reservationMap
		data[fmt.Sprintf("block_map_%d", x.ID)] = blockMap

		m.App.Session.Put(r.Context(), fmt.Sprintf("block_map_%d", x.ID), blockMap)

	}

	render.Template(w, "admin-reservations-calender.page.tmpl", r, &models.TemplateData{
		StringMap: stringMap,
		Data: data,
		IntMap: intMap,
	})
}


func (m *Repository) AdminProcessReservation(w http.ResponseWriter, r *http.Request){
	id, _ := strconv.Atoi(chi.URLParam(r, "id"))
	src := chi.URLParam(r, "src")
	err := m.DB.UpdateProcessedForReservation(id, 1)
	if err !=nil {
		m.App.Session.Put(r.Context(), "error", "error happend when processing")
		http.Redirect(w, r, fmt.Sprintf("/admin/reservations-%s", src), http.StatusSeeOther)
		helpers.ServerError(w, err)
		return	
	}

	year := r.URL.Query().Get("y")
	month := r.URL.Query().Get("m")

	m.App.Session.Put(r.Context(), "flash", "Reservation marked as processed")
	if year == ""{
		http.Redirect(w, r, fmt.Sprintf("/admin/reservations-#{src}"), http.StatusSeeOther)
	}else {
		http.Redirect(w, r, fmt.Sprintf("/admin/reservations-calendar?y=%s&m=%s", year, month), http.StatusSeeOther)
	}
}

func (m *Repository) AdminDeleteReservation(w http.ResponseWriter, r *http.Request){
	id, _ := strconv.Atoi(chi.URLParam(r, "id"))
	src := chi.URLParam(r, "src")
	err := m.DB.DeleteReservation(id)
	if err !=nil {
		m.App.Session.Put(r.Context(), "error", "error happend when processing")
		http.Redirect(w, r, fmt.Sprintf("/admin/reservations-%s", src), http.StatusSeeOther)
		helpers.ServerError(w, err)
		return	
	}

	year := r.URL.Query().Get("y")
	month := r.URL.Query().Get("m")

	m.App.Session.Put(r.Context(), "flash", "Reservation Deleted")
	if year == "" {
		http.Redirect(w, r, fmt.Sprintf("/admin/reservations-%s", src), http.StatusSeeOther)
	}else {
		http.Redirect(w, r, fmt.Sprintf("/admin/reservations-calendar?y=%s&m=%s", year, month), http.StatusSeeOther)
	}
}

func (m *Repository) AdminPostReservationsCalendar(w http.ResponseWriter, r *http.Request){
	err := r.ParseForm()
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	year, _ := strconv.Atoi(r.Form.Get("y"))
	month, _ := strconv.Atoi(r.Form.Get("m"))

	//prcess blocks
	rooms, err := m.DB.AllRooms()
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	form := forms.New(r.PostForm)
	for _, x := range rooms{
		// Get the block map from the session loop through map, if we have an entry in the map
		// that does not exist in our posted data,  and if the restriction id > 0, then it is a block
		// we need to remove
		curMap := m.App.Session.Get(r.Context(), fmt.Sprintf("block_map_%d", x.ID)).(map[string]int)
		for name, value := range curMap{
			// ok will be false if the value is not in the map
			if val, ok := curMap[name]; ok{
				//only pay attention to value >0, and that are not in the form post
				// the rest are just place holders for days without blocks
				if val > 0{
					if !form.Has(fmt.Sprintf("remove_block_%d_%s", x.ID, name), r){
						//delete the restriction by id
						err := m.DB.DeleteBlockForRoom(value)
						if err != nil {
							log.Println(err)
						}
					}
				}
			}
		}
	}

	//now handle new blocks
	for name, _ := range r.PostForm {
		if strings.HasPrefix(name, "add_block"){
			exploded := strings.Split(name, "_")
			roomID, _ := strconv.Atoi(exploded[2])

			t, _ := time.Parse("2006-01-02", exploded[3])
			//insert a new block
			err := m.DB.InsertBlockForRoom(roomID, t)
			if err != nil {
				log.Println(err)
			}
		}
	}

	m.App.Session.Put(r.Context(), "flash", "changes saved")
	http.Redirect(w, r, fmt.Sprintf("/admin/reservations-calendar?y=%d&m=%d", year, month), http.StatusSeeOther)
}