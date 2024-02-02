package main

import (
	"fmt"
	"html/template"
	"math"
	"net/http"
	"os"
	"strconv"

	"github.com/sirupsen/logrus"
	"golang.org/x/time/rate"
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

var limiter = rate.NewLimiter(1, 3) // Rate limit of 1 request per second with a burst of 3 requests

var logger = logrus.New()

func isSelected(currentSort, optionSort string) bool {
	return currentSort == optionSort
}

func init() {
	// Set log formatter to JSON for structured logging
	logger.SetFormatter(&logrus.JSONFormatter{})
	// Log to stdout
	logger.SetOutput(os.Stdout)
	// Set log level to Info
	logger.SetLevel(logrus.InfoLevel)
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

func seq(start, end int) []int {
	s := make([]int, end-start+1)
	for i := range s {
		s[i] = start + i
	}
	return s
}

func main() {
	dsn := "user=postgres password=123 dbname=courses sslmode=disable"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		logger.WithFields(logrus.Fields{
			"action": "database_connection",
			"status": "failure",
		}).Error("Failed to connect to the database")
		os.Exit(1)
	}

	db.AutoMigrate(&User{}, &AdditionalCourses{})

	// Log the migration status
	logger.WithFields(logrus.Fields{
		"action": "database_migration",
		"status": "success",
	}).Info("Database migration completed successfully")

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if !limiter.Allow() {
			// Exceeded request limit
			logger.WithFields(logrus.Fields{
				"action": "rate_limit_exceeded",
				"status": "failure",
			}).Error("Rate limit exceeded")
			http.Error(w, "Rate limit exceeded", http.StatusTooManyRequests)
			return
		}
		indexHandler(w, r)
	})

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

	logger.WithFields(logrus.Fields{
		"action": "server_start",
		"status": "success",
	}).Info("Server is running on :8080")

	http.ListenAndServe(":8080", nil)
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		serveHTMLFile(w, r, "page/index.html")
		return
	}

	logger.WithFields(logrus.Fields{
		"action": "index_handler",
		"status": "failure",
	}).Error("Method Not Allowed")

	http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
}

func submitHandler(w http.ResponseWriter, r *http.Request, db *gorm.DB) {
	if r.Method == http.MethodPost {
		name := r.FormValue("name")
		email := r.FormValue("email")
		password := r.FormValue("password")

		if name == "" || email == "" || password == "" {
			logger.WithFields(logrus.Fields{
				"action": "submit_handler",
				"status": "failure",
			}).Error("Invalid form data")
			http.Redirect(w, r, "/error", http.StatusSeeOther)
			return
		}

		userRepo := NewUserRepository(db)
		userRepo.CreateUser(&User{Name: name, Email: email, Password: password})

		logger.WithFields(logrus.Fields{
			"action": "user_created",
			"status": "success",
			"user":   name,
		}).Info("User created successfully")

		fmt.Printf("Name: %s\nEmail: %s\nPassword: %s\n", name, email, password)

		http.Redirect(w, r, "/success", http.StatusSeeOther)
		return
	}

	logger.WithFields(logrus.Fields{
		"action": "submit_handler",
		"status": "failure",
	}).Error("Method Not Allowed")

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
		action := r.URL.Query().Get("action")
		pageStr := r.URL.Query().Get("page")
		perPage := 3 // Number of items per page

		var courses []AdditionalCourses
		query := db

		sort := r.URL.Query().Get("sort")
		filter := r.URL.Query().Get("filter")

		if filter != "" {
			query = query.Where("course_name ILIKE ?", "%"+filter+"%") // Adjust this line based on your database column
		}

		if sort == "" {
			query = query.Order("course_name")
		}

		switch action {
		case "filter":
			categories := r.URL.Query()["categories"]
			if len(categories) > 0 {
				query = query.Joins("JOIN course_categories ON additional_courses.id = course_categories.course_id").
					Joins("JOIN categories ON course_categories.category_id = categories.id").
					Where("categories.name IN (?)", categories)
			}
		case "sort":
			// Sorting logic based on the 'sort' parameter
			switch sort {
			case "course_name":
				query = query.Order("course_name")
			case "price":
				query = query.Order("price")
			case "recorded_date":
				query = query.Order("recorded_date")

			default:
				query = query.Order("course_name")
			}
		case "search":
			searchTerm := r.URL.Query().Get("search")
			if searchTerm != "" {
				query = query.Where("course_name LIKE ?", "%"+searchTerm+"%")
			}
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

		query = query.Order(sort).Offset(offset).Limit(perPage).Find(&courses)

		if query.Error != nil {
			logger.WithFields(logrus.Fields{
				"action": "filtered_courses_handler",
				"status": "failure",
				"error":  query.Error.Error(),
			}).Error("Error executing database query")

			// Optionally, you can return an HTTP 500 Internal Server Error status
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		// Calculate total pages
		totalPages := int(math.Ceil(float64(totalCount) / float64(perPage)))

		// Execute template with pagination data
		err = renderCourses(w, PageData{
			Courses:     courses,
			CurrentPage: page,
			TotalPages:  totalPages,
			Sort:        sort,
			Filter:      filter,
		})

		return
	}

}

func renderCourses(w http.ResponseWriter, pageData PageData) error {
	tmpl, err := template.New("").Funcs(template.FuncMap{"seq": seq}).ParseGlob("page/*.html")
	if err != nil {
		logger.WithFields(logrus.Fields{
			"action": "render_courses",
			"status": "failure",
		}).Error("Error parsing templates: ", err)
		return err
	}

	// Execute template with pagination data
	err = tmpl.ExecuteTemplate(w, "courses.html", pageData)
	if err != nil {
		logger.WithFields(logrus.Fields{
			"action": "render_courses",
			"status": "failure",
		}).Error("Error rendering template: ", err)
	}

	// Log the render courses action
	logger.WithFields(logrus.Fields{
		"action": "render_courses",
		"status": "success",
	}).Info("Courses rendered successfully")

	return err
}
