package template

import (
	"frontend-main/internal/models"
	"html/template"
	"net/http"

	"github.com/sirupsen/logrus"
)

var logger = logrus.New()

func renderCourses(w http.ResponseWriter, pageData models.PageData) error {
	tmpl, err := template.New("").Funcs(template.FuncMap{"seq": seq}).ParseGlob("page/*.html")
	if err != nil {
		logger.WithFields(logrus.Fields{
			"action": "render_courses",
			"status": "failure",
		}).Error("Error parsing templates: ", err)
		return err
	}

	err = tmpl.ExecuteTemplate(w, "courses.html", pageData)
	if err != nil {
		logger.WithFields(logrus.Fields{
			"action": "render_courses",
			"status": "failure",
		}).Error("Error rendering template: ", err)
	}

	logger.WithFields(logrus.Fields{
		"action": "render_courses",
		"status": "success",
	}).Info("Courses rendered successfully")

	return err
}

func serveHTMLFile(w http.ResponseWriter, r *http.Request, filename string) {
	http.ServeFile(w, r, filename)
}
