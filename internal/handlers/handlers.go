package handlers

import (
	"fmt"
	"frontend-main/internal/models"
	"frontend-main/internal/repository"
	"frontend-main/internal/template"
	"math"
	"net/http"
	"strconv"

	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

var logger = logrus.New()

func IndexHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		template.ServeHTMLFile(w, r, "../page/index.html")
		return
	}

	logger.WithFields(logrus.Fields{
		"action": "index_handler",
		"status": "failure",
	}).Error("Method Not Allowed")

	http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
}

func SubmitHandler(w http.ResponseWriter, r *http.Request, db *gorm.DB) {
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

		userRepo := repository.NewUserRepository(db)
		userRepo.CreateUser(&models.User{Name: name, Email: email, Password: password})

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

func ErrorPageHandler(w http.ResponseWriter, r *http.Request) {
	template.ServeHTMLFile(w, r, "../page/error.html")
}

func SuccessPageHandler(w http.ResponseWriter, r *http.Request) {
	template.ServeHTMLFile(w, r, "../page/success.html")
}

func FilteredCoursesHandler(w http.ResponseWriter, r *http.Request, db *gorm.DB) {

	if r.Method == http.MethodGet {
		action := r.URL.Query().Get("action")
		pageStr := r.URL.Query().Get("page")
		perPage := 3 // Number of items per page

		var courses []models.AdditionalCourses
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
		query.Model(&models.AdditionalCourses{}).Count(&totalCount)

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

			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		totalPages := int(math.Ceil(float64(totalCount) / float64(perPage)))

		err = template.RenderCourses(w, models.PageData{
			Courses:     courses,
			CurrentPage: page,
			TotalPages:  totalPages,
			Sort:        sort,
			Filter:      filter,
		})

		return
	}

}
