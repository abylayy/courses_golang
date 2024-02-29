package models

import "gorm.io/gorm"

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
