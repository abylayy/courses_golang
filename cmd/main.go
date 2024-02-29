package main

import (
	"frontend-main/internal/handlers"
	"frontend-main/internal/models"
	"frontend-main/internal/utils"

	"net/http"
	"os"

	"github.com/sirupsen/logrus"
	"golang.org/x/time/rate"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var limiter = rate.NewLimiter(1, 3) // Rate limit of 1 request per second with a burst of 3 requests
var logger = logrus.New()

func main() {
	logger := logrus.New()
	utils.InitLogger(logger)
	dsn := "user=postgres password=CoIrD857 dbname=courses sslmode=disable"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		logger.WithFields(logrus.Fields{
			"action": "database_connection",
			"status": "failure",
		}).Error("Failed to connect to the database")
		os.Exit(1)
	}

	db.AutoMigrate(&models.User{}, &models.AdditionalCourses{})

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
		handlers.IndexHandler(w, r)
	})

	http.HandleFunc("/submit", func(w http.ResponseWriter, r *http.Request) {
		handlers.SubmitHandler(w, r, db)
	})

	http.HandleFunc("/error", handlers.ErrorPageHandler)

	http.HandleFunc("/success", handlers.SuccessPageHandler)

	http.HandleFunc("/additional-courses", func(w http.ResponseWriter, r *http.Request) {
		handlers.FilteredCoursesHandler(w, r, db)
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
