package main

import (
	"net/http"
	"os"

	"github.com/sirupsen/logrus"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var logger = logrus.New()

func main() {
	dsn := "user=postgres password=CoIrD857 dbname=courses sslmode=disable"
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
