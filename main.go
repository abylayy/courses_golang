package main

import (
	"fmt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"log"
	"net/http"
)

type User struct {
	gorm.Model
	Name     string
	Email    string
	Password string
}

type UserRepository struct {
	DB *gorm.DB
}

// creates a new UserRepository
func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{DB: db}
}

// creates a new user
func (ur *UserRepository) CreateUser(user *User) {
	ur.DB.Create(user)
}

// retrieves a user by ID
func (ur *UserRepository) GetUserByID(id uint) *User {
	var user User
	ur.DB.First(&user, id)
	return &user
}

// updates the user's name by ID
func (ur *UserRepository) UpdateUserNameByID(id uint, newName string) {
	var user User
	ur.DB.First(&user, id)
	ur.DB.Model(&user).Update("Name", newName)
}

// deletes a user by ID
func (ur *UserRepository) DeleteUserByID(id uint) {
	var user User
	ur.DB.Delete(&user, id)
}

// retrieves a list of all users
func (ur *UserRepository) GetAllUsers() []User {
	var users []User
	ur.DB.Find(&users)
	return users
}

func main() {
	// Connect to the database
	dsn := "user=postgres password=123 dbname=courses sslmode=disable"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to the database")
	}

	db.AutoMigrate(&User{})

	http.HandleFunc("/", indexHandler)
	http.HandleFunc("/submit", func(w http.ResponseWriter, r *http.Request) {
		submitHandler(w, r, db)
	})
	http.HandleFunc("/error", errorPageHandler)
	http.HandleFunc("/success", successPageHandler)

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

		// Add the user to the database
		userRepo := NewUserRepository(db)
		userRepo.CreateUser(&User{Name: name, Email: email, Password: password})

		fmt.Printf("Name: %s\nEmail: %s\nPassword: %s\n", name, email, password)

		// Redirect to the success page
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
