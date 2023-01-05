package main

import (
	"context"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"personal-web/connection"
	"personal-web/middleware"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	"golang.org/x/crypto/bcrypt"
)

type MetaData struct {
	Id          int
	ProjectName string
	IsLogin     bool
	UserName    string
	FlashData   string
}

var Data = MetaData{}

// STRUCT TEMPLATE
type Project struct {
	ID              int
	ProjectName     string
	StartDate       time.Time
	EndDate         time.Time
	StartDateString string
	EndDateString   string
	Duration        string
	Description     string
	Technologies    []string
	Image           string
	UserID          string
}

type User struct {
	Id       int
	Name     string
	Email    string
	Password string
}

// LOCAL DATABASE
var ProjectList = []Project{}

func main() {
	route := mux.NewRouter()

	// CONNECTION TO DATABASE
	connection.DatabaseConnect()

	// route path folder for public folder
	route.PathPrefix("/public/").Handler(http.StripPrefix("/public/", http.FileServer(http.Dir("./public"))))
	route.PathPrefix("/uploads/").Handler(http.StripPrefix("/uploads/", http.FileServer(http.Dir("./uploads"))))

	route.HandleFunc("/", home).Methods("GET")

	// CONTACT
	route.HandleFunc("/contact", contact).Methods("GET")

	// CREATE PROJECT
	route.HandleFunc("/project", project).Methods("GET")
	route.HandleFunc("/edit-project/{id}", editForm).Methods("GET")
	route.HandleFunc("/edited-project/{id}", middleware.UploadFile(editProject)).Methods("POST")
	route.HandleFunc("/project-detail/{id}", ProjectDetails).Methods("GET")
	route.HandleFunc("/project/addproject", middleware.UploadFile(addproject)).Methods("POST")
	route.HandleFunc("/project-details/{id}", ProjectDetails).Methods("GET")
	route.HandleFunc("/delete-project/{id}", DeleteProject).Methods("GET")
	route.HandleFunc("/register", formRegister).Methods("GET")
	route.HandleFunc("/register", register).Methods("POST")
	route.HandleFunc("/login", formLogin).Methods("GET")
	route.HandleFunc("/login", login).Methods("POST")
	route.HandleFunc("/logout", logout).Methods("GET")

	fmt.Println("Server running on port 8000")
	http.ListenAndServe("localhost:8000", route)
}

// RENDER
func home(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	tmpl, err := template.ParseFiles("views/index.html")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("message : " + err.Error()))
		return
	} else {
		var renderData []Project
		var item = Project{}

		rows, _ := connection.Conn.Query(context.Background(), `SELECT tb_project.id, project_name, start_date, end_date,description, image, technologies, tb_user.name FROM tb_project LEFT JOIN tb_user ON tb_project.user_id=tb_user.id`)
		for rows.Next() {

			err := rows.Scan(&item.ID, &item.ProjectName, &item.StartDate, &item.EndDate, &item.Description, &item.Image, &item.Technologies, &item.UserID)

			if err != nil {
				fmt.Println(err.Error())
				return
			} else {
				// PARSING DATE
				item := Project{
					ID:           item.ID,
					ProjectName:  item.ProjectName,
					Duration:     GetDuration(item.StartDate, item.EndDate),
					Description:  item.Description,
					Technologies: item.Technologies,
					Image:        item.Image,
					UserID:       item.UserID,
				}
				renderData = append(renderData, item)
			}
		}
		response := map[string]interface{}{
			"renderData": renderData,
			"Data":       Data,
		}

		store := sessions.NewCookieStore([]byte("SESSIONS_ID"))
		session, _ := store.Get(r, "SESSIONS_ID")

		if session.Values["IsLogin"] != true {
			Data.IsLogin = false
		} else {
			Data.IsLogin = session.Values["IsLogin"].(bool)
			Data.UserName = session.Values["Name"].(string)
			Data.Id = session.Values["Id"].(int)
		}

		fm := session.Flashes("message")
		var flashes []string
		if len(fm) > 0 {
			session.Save(r, w)

			for _, fl := range fm {
				flashes = append(flashes, fl.(string))
			}
			Data.FlashData = strings.Join(flashes, "")
		}

		w.WriteHeader(http.StatusOK)
		tmpl.Execute(w, response)
	}

}

func contact(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	var tmpl, err = template.ParseFiles("views/contact.html")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Message : " + err.Error()))
		return
	}

	w.WriteHeader(http.StatusOK)
	tmpl.Execute(w, nil)
}

// CREATE PROJECT
func project(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	tmpl, err := template.ParseFiles("views/project.html")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Message : " + err.Error()))
		return
	}

	w.WriteHeader(http.StatusOK)
	tmpl.Execute(w, nil)
}

func addproject(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	const (
		layoutISO = "2006-01-02"
	)

	if err != nil {
		log.Fatal(err)
	} else {
		ProjectName := r.PostForm.Get("ProjectName")
		Description := r.PostForm.Get("description")
		StartDate, _ := time.Parse(layoutISO, r.PostForm.Get("date-start"))
		EndDate, _ := time.Parse(layoutISO, r.PostForm.Get("date-end"))
		Technologies := r.Form["technologies"]

		store := sessions.NewCookieStore([]byte("SESSIONS_ID"))
		session, _ := store.Get(r, "SESSIONS_ID")
		user := session.Values["Id"].(int)

		dataContext := r.Context().Value("dataFile")
		image := dataContext.(string)

		_, err = connection.Conn.Exec(context.Background(), `INSERT INTO public.tb_project( project_name, start_date, end_date, description, technologies, image, user_id)
			VALUES ( $1, $2, $3, $4, $5, $6, $7)`, ProjectName, StartDate, EndDate, Description, Technologies, image, user)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("message : " + err.Error()))
			return
		}

		// ProjectList = append(ProjectList, newProject)

		http.Redirect(w, r, "/", http.StatusMovedPermanently)
	}
}

func ProjectDetails(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	var tmpl, err = template.ParseFiles("views/project-detail.html")
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Message : " + err.Error()))
		return
	} else {
		id, _ := strconv.Atoi(mux.Vars(r)["id"])
		renderDetails := Project{}

		// GET ID FROM DATABASE
		err = connection.Conn.QueryRow(context.Background(), `SELECT id, project_name, start_date, end_date, description, technologies, image
		FROM public.tb_project WHERE id=$1`, id).Scan(&renderDetails.ID, &renderDetails.ProjectName, &renderDetails.StartDate, &renderDetails.EndDate, &renderDetails.Description, &renderDetails.Technologies, &renderDetails.Image)

		// ERROR HANDLING
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("message : " + err.Error()))
		} else {
			// PARSING DATE
			renderDetails := Project{
				ID:              renderDetails.ID,
				ProjectName:     renderDetails.ProjectName,
				StartDateString: FormatDate(renderDetails.StartDate),
				EndDateString:   FormatDate(renderDetails.EndDate),
				Duration:        GetDuration(renderDetails.StartDate, renderDetails.EndDate),
				Description:     renderDetails.Description,
				Technologies:    renderDetails.Technologies,
				Image:           renderDetails.Image,
			}
			response := map[string]interface{}{
				"renderDetails": renderDetails,
			}
			w.WriteHeader(http.StatusOK)
			tmpl.Execute(w, response)
		}
	}
}

// UPDATE PROJECT
func editForm(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-type", "text/html; charset=utf-8")
	tmpt, err := template.ParseFiles("views/edit-project.html")

	if err != nil {
		w.Write([]byte("Message: " + err.Error()))
		return
	} else {

		id, _ := strconv.Atoi(mux.Vars(r)["id"])

		var resultData = Project{}

		err = connection.Conn.QueryRow(context.Background(), `SELECT id, project_name, start_date, end_date, description
		FROM public.tb_project WHERE id=$1`, id).Scan(
			&resultData.ID, &resultData.ProjectName, &resultData.StartDate, &resultData.EndDate, &resultData.Description)

		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("Message: " + err.Error()))
			return
		}

		data := map[string]interface{}{
			"Project": resultData,
		}

		resultData.StartDateString = resultData.StartDate.Format("2 January 2006")
		resultData.EndDateString = resultData.EndDate.Format("2 January 2006")

		tmpt.Execute(w, data)
	}
}

func editProject(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	const (
		layoutISO = "2006-01-02"
	)

	if err != nil {
		log.Fatal(err)
	} else {
		ID, _ := strconv.Atoi(mux.Vars(r)["id"])
		ProjectName := r.PostForm.Get("ProjectName")
		Description := r.PostForm.Get("description")
		StartDate, _ := time.Parse(layoutISO, r.PostForm.Get("date-start"))
		EndDate, _ := time.Parse(layoutISO, r.PostForm.Get("date-end"))
		Technologies := r.Form["technologies"]
		// Image := r.PostForm.Get("upload-image")

		dataContext := r.Context().Value("dataFile")
		image := dataContext.(string)

		_, err = connection.Conn.Exec(context.Background(), `UPDATE public.tb_project
		SET project_name=$1, description=$2, start_date=$3, end_date=$4, technologies=$5, image=$6
		WHERE id=$7;`, ProjectName, Description, StartDate, EndDate, Technologies, image, ID)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("message : " + err.Error()))
			return
		}

		// ProjectList = append(ProjectList, newProject)

		http.Redirect(w, r, "/", http.StatusMovedPermanently)
	}
}

func formRegister(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	tmpl, err := template.ParseFiles("views/formregister.html")

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("message : " + err.Error()))
	}

	w.WriteHeader(http.StatusOK)
	tmpl.Execute(w, nil)

}
func register(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()

	if err != nil {
		log.Fatal(err)
	} else {
		name := r.PostForm.Get("name")
		email := r.PostForm.Get("email")

		password := r.PostForm.Get("password")
		passwordHash, _ := bcrypt.GenerateFromPassword([]byte(password), 10) //10 adalah random string

		_, err = connection.Conn.Exec(context.Background(), `INSERT INTO public.tb_user( name, email, password)
			VALUES ( $1, $2, $3)`, name, email, passwordHash)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("message : " + err.Error()))
			return
		}

		store := sessions.NewCookieStore([]byte("SESSIONS_ID"))
		session, _ := store.Get(r, "SESSIONS_ID")

		session.AddFlash("Successfully Login !!!", "message")
		session.Save(r, w)

		http.Redirect(w, r, "/", http.StatusMovedPermanently)
	}

}

func formLogin(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	tmpl, err := template.ParseFiles("views/formLogin.html")

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("message : " + err.Error()))
	}

	w.WriteHeader(http.StatusOK)
	tmpl.Execute(w, nil)
}

func login(w http.ResponseWriter, r *http.Request) {
	store := sessions.NewCookieStore([]byte("SESSIONS_ID")) //sessions_id untuk menyimpan id dari sesi nya
	session, _ := store.Get(r, "SESSIONS_ID")               //untuk mengecek sudah login atau belum

	err := r.ParseForm()
	if err != nil {
		log.Fatal(err)
	}

	email := r.PostForm.Get("email")
	password := r.PostForm.Get("password")

	user := User{}

	err = connection.Conn.QueryRow(context.Background(), `SELECT id, name, email, password
		FROM public.tb_user WHERE email=$1`, email).Scan(&user.Id, &user.Name, &user.Email, &user.Password)

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("message : " + err.Error()))
		return
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)) //mengecek value string random password dan yg dimasukkan
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("message : " + err.Error()))
		return
	}

	session.Values["IsLogin"] = true   //value session_id disimpan pada IsLogin (bernilai [tru] jika berhasil login)
	session.Values["Name"] = user.Name //untuk mengetahui siapa yang login
	session.Values["Id"] = user.Id
	session.Options.MaxAge = 7200 //umur dari Cookies selama 2 jam

	session.AddFlash("Successfully Login !!!", "message") //message dibelakang jika tidak berhasil login
	session.Save(r, w)                                    //untuk menyimpan codingan dari atas

	http.Redirect(w, r, "/", http.StatusMovedPermanently)
}

func logout(w http.ResponseWriter, r *http.Request) {
	store := sessions.NewCookieStore([]byte("SESSIONS_ID"))
	session, _ := store.Get(r, "SESSIONS_ID")
	session.Options.MaxAge = -1
	session.Save(r, w)

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

// DELETE PROJECT
func DeleteProject(w http.ResponseWriter, r *http.Request) {

	id, _ := strconv.Atoi(mux.Vars(r)["id"])

	_, errQuery := connection.Conn.Exec(context.Background(), "DELETE FROM tb_project WHERE id=$1", id)

	if errQuery != nil {
		fmt.Println("Message : " + errQuery.Error())
		return
	}

	// projects = append(projects[:index], projects[index+1:]...)

	http.Redirect(w, r, "/", http.StatusFound)
}

// ADDITIONAL FUNCTION

// GET DURATION
func GetDuration(startDate time.Time, endDate time.Time) string {

	margin := endDate.Sub(startDate).Hours() / 24
	var duration string

	if margin >= 30 {
		if (margin / 30) == 1 {
			duration = "1 Month"
		} else {
			duration = strconv.Itoa(int(margin/30)) + " Months"
		}
	} else {
		if margin <= 1 {
			duration = "1 Day"
		} else {
			duration = strconv.Itoa(int(margin)) + " Days"
		}
	}

	return duration
}

// CHANGE DATE FORMAT
func FormatDate(InputDate time.Time) string {

	Formated := InputDate.Format("02 January 2006")

	return Formated
}

// RETURN DATE FORMAT
func ReturnDate(InputDate time.Time) string {

	Formated := InputDate.Format("2006-01-02")

	return Formated
}
