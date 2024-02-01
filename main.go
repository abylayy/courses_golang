package main

import (
	"fmt"
	"html/template"
	"log"
	"math"
	"net/http"
	"strconv"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
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
	Courses     []AdditionalCourses
	CurrentPage int
	TotalPages  int
	Filter      string
	Sort        string
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
		sort := r.URL.Query().Get("sort")
		pageStr := r.URL.Query().Get("page")
		perPage := 3 // Number of items per page

		var courses []AdditionalCourses
		query := db

		if filter != "" {
			query = query.Where("course_name LIKE ?", "%"+filter+"%")
		}

		// Sorting logic based on the 'sort' parameter
		switch sort {
		case "name":
			query = query.Order("course_name")
		case "price":
			query = query.Order("price")
		case "date":
			query = query.Order("recorded_date")
		default:
			// You can set a default sorting option here
			// For example, sort by course name if 'sort' parameter is not provided
			query = query.Order("course_name")
		}

		// Get total count for pagination
		var totalCount int64
		query.Model(&AdditionalCourses{}).Count(&totalCount)

		// Calculate offset based on page number
		page, err := strconv.Atoi(pageStr)
		if err != nil || page < 1 {
			page = 1
		}
		offset := (page - 1) * perPage

		// Retrieve courses for the specified page
		result := query.Offset(offset).Limit(perPage).Find(&courses)
		if result.Error != nil {
			log.Printf("Error querying the database: %v", result.Error)
			http.Error(w, "Error querying the database", http.StatusInternalServerError)
			return
		}

		// Calculate total pages
		totalPages := int(math.Ceil(float64(totalCount) / float64(perPage)))

		// Execute template with pagination data
		err = renderCourses(w, PageData{
			Courses:     courses,
			CurrentPage: page,
			TotalPages:  totalPages,
			Filter:      filter,
			Sort:        sort,
		})

		return
	}

	http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
}

func renderCourses(w http.ResponseWriter, pageData PageData) error {
	tmpl, err := template.New("").ParseGlob("page/*.html")
	if err != nil {
		return err
	}

	// Execute template with pagination data
	return tmpl.ExecuteTemplate(w, "courses.html", pageData)
}
