package main

import (
	"fmt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"html/template"
	"log"
	"net/http"
)

type User struct {
	gorm.Model
	Name     string
	Email    string
	Password string
}

type AdditionalCourses struct {
	ID           uint
	CourseName   string
	Description  string
	Price        float64
	Sessions     int64
	RecordedDate string
	TotalUsers   int64
}

type PageData struct {
	Courses []AdditionalCourses
}

type UserRepository struct {
	DB *gorm.DB
}

func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{DB: db}
}

func (ur *UserRepository) CreateUser(user *User) {
	ur.DB.Create(user)
}

func (ur *UserRepository) GetUserByID(id uint) *User {
	var user User
	ur.DB.First(&user, id)
	return &user
}

func (ur *UserRepository) UpdateUserNameByID(id uint, newName string) {
	var user User
	ur.DB.First(&user, id)
	ur.DB.Model(&user).Update("Name", newName)
}

func (ur *UserRepository) DeleteUserByID(id uint) {
	var user User
	ur.DB.Delete(&user, id)
}

func (ur *UserRepository) GetAllUsers() []User {
	var users []User
	ur.DB.Find(&users)
	return users
}

func main() {
	dsn := "user=postgres password=123 dbname=courses sslmode=disable"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to the database")
	}

	db.AutoMigrate(&User{}, &AdditionalCourses{})

	// Log the migration status
	log.Println("Database migration completed successfully")

	http.HandleFunc("/", indexHandler)
	http.HandleFunc("/submit", func(w http.ResponseWriter, r *http.Request) {
		submitHandler(w, r, db)
	})
	http.HandleFunc("/error", errorPageHandler)
	http.HandleFunc("/success", successPageHandler)
	http.HandleFunc("/additional-courses", func(w http.ResponseWriter, r *http.Request) {
		filteredCoursesHandler(w, r, db)
	})

	http.Handle("/styles/", http.StripPrefix("/styles/", http.FileServer(http.Dir("styles"))))
	http.Handle("/images/", http.StripPrefix("/images/", http.FileServer(http.Dir("images"))))
	http.Handle("/page/", http.StripPrefix("/page/", http.FileServer(http.Dir("page"))))
	http.Handle("/javascript/", http.StripPrefix("/javascript/", http.FileServer(http.Dir("javascript"))))

	fmt.Println("Server is running on :8080")
	http.ListenAndServe(":8080", nil)
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		serveHTMLFile(w, r, "page/index.html")
		return
	}

	http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
}

func submitHandler(w http.ResponseWriter, r *http.Request, db *gorm.DB) {
	if r.Method == http.MethodPost {
		name := r.FormValue("name")
		email := r.FormValue("email")
		password := r.FormValue("password")

		if name == "" || email == "" || password == "" {
			http.Redirect(w, r, "/error", http.StatusSeeOther)
			return
		}

		userRepo := NewUserRepository(db)
		userRepo.CreateUser(&User{Name: name, Email: email, Password: password})

		fmt.Printf("Name: %s\nEmail: %s\nPassword: %s\n", name, email, password)

		http.Redirect(w, r, "/success", http.StatusSeeOther)
		return
	}

	http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
}

func errorPageHandler(w http.ResponseWriter, r *http.Request) {
	serveHTMLFile(w, r, "page/error.html")
}

func successPageHandler(w http.ResponseWriter, r *http.Request) {
	serveHTMLFile(w, r, "page/success.html")
}

func serveHTMLFile(w http.ResponseWriter, r *http.Request, filename string) {
	http.ServeFile(w, r, filename)
}

func filteredCoursesHandler(w http.ResponseWriter, r *http.Request, db *gorm.DB) {
	if r.Method == http.MethodGet {
		filter := r.URL.Query().Get("filter")
		var courses []AdditionalCourses
		if filter != "" {
			result := db.Where("course_name LIKE ?", "%"+filter+"%").Find(&courses)
			if result.Error != nil {
				log.Printf("Error querying the database: %v", result.Error)
				http.Error(w, "Error querying the database", http.StatusInternalServerError)
				return
			}
		} else {
			result := db.Find(&courses)
			if result.Error != nil {
				log.Printf("Error querying the database: %v", result.Error)
				http.Error(w, "Error querying the database", http.StatusInternalServerError)
				return
			}
		}

		// Print the retrieved courses
		log.Printf("Retrieved courses: %+v", courses)

		renderCourses(w, courses)
		return
	}

	http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
}

func renderCourses(w http.ResponseWriter, courses []AdditionalCourses) {
	tmpl, err := template.New("").ParseGlob("page/*.html")
	if err != nil {
		log.Printf("Error parsing template: %v", err)
		http.Error(w, "Error rendering template", http.StatusInternalServerError)
		return
	}

	// Print the parsed templates
	log.Printf("Parsed templates: %s", tmpl.DefinedTemplates())

	err = tmpl.ExecuteTemplate(w, "courses.html", struct{ Courses []AdditionalCourses }{Courses: courses})
	if err != nil {
		log.Printf("Error executing template: %v", err)
		http.Error(w, "Error rendering template", http.StatusInternalServerError)
		return
	}
}
