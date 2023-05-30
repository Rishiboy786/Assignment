package main

import (
	dbconnection "assignment/dbconnection"
	"fmt"
	"net/http"
	"path/filepath"
	"regexp"
	"time"

	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
)

type User struct {
	ID        int
	FirstName string
	LastName  string
	DOB       time.Time
	Email     string
	Phone     string
	CVPath    string
}

func main() {
	// Set up the Gin router
	router := gin.Default()

	// Set the static files directory
	router.Static("/uploads", "./uploads")

	// Load the HTML templates
	router.LoadHTMLGlob("templates/*")

	// Routes
	router.GET("/", showRegistrationForm)
	router.POST("/register", registerUser)
	router.GET("/users", showUsers)

	// Start the server
	router.Run(":8080")
}

func showRegistrationForm(c *gin.Context) {
	c.HTML(http.StatusOK, "register.html", nil)
}

func registerUser(c *gin.Context) {
	// Parse form data
	firstName := c.PostForm("first_name")
	lastName := c.PostForm("last_name")
	dob := c.PostForm("dob")
	email := c.PostForm("email")
	phone := c.PostForm("phone")
	cv, err := c.FormFile("cv")
	
	if err != nil {
		c.String(http.StatusBadRequest, "Failed to upload CV")
		return
	}

	if firstName == "" || lastName == "" || dob == "" || email == "" || phone == "" {
		c.String(http.StatusBadRequest, "Please fill in all fields")
		return
	}

	dobTime, err := time.Parse("2006-01-02", dob)
	if err != nil {
		c.String(http.StatusBadRequest, "Invalid date format")
		return
	}
	age := time.Since(dobTime) / (365 * 24 * time.Hour)
	if age < 18 {
		c.String(http.StatusBadRequest, "Age should be at least 18 years")
		return
	}

	// Perform phone number validation
	if !validatePhoneNumber(phone) {
		c.String(http.StatusBadRequest, "Invalid phone number")
		return
	}

	// Save CV to disk
	cvPath := filepath.Join("uploads", cv.Filename)
	if err := c.SaveUploadedFile(cv, cvPath); err != nil {
		c.String(http.StatusInternalServerError, "Failed to save CV")
		return
	}

	if err := saveFormData(firstName, lastName, dobTime, email, phone, cvPath); err != nil {
		
		c.String(http.StatusInternalServerError, "Failed to save form data")
		return
	}

	// Send email to the form submitter
	if err := sendEmail(email); err != nil {
		c.String(http.StatusInternalServerError, "Failed to send email")
		return
	}

	c.String(http.StatusOK, "Registration successful")
}

func showUsers(c *gin.Context) {
	users, err := getUsers()
	if err != nil {
		c.String(http.StatusInternalServerError, "Failed to fetch users")
		return
	}

	c.HTML(http.StatusOK, "users.html", gin.H{
		"Users": users,
	})
}

func saveFormData(firstName, lastName string, dob time.Time, email, phone, cvPath string) error {
	
	db, err := dbconnection.SetupDB()
	if err != nil {
		return err
	}
	defer db.Close()
	fmt.Print(firstName, lastName, dob, email, cvPath, phone)

	row, err := db.Query("INSERT INTO users (first_name, last_name, dob, email, phone,cvpath) VALUES (?, ?, ?, ?, ?, ?)",
		firstName, lastName, dob.Format("2006-01-02"), email, phone)
	if err != nil {
		return err
	}
	fmt.Print(row)

	return nil
}

func getUsers() ([]User, error) {
	db, err := dbconnection.SetupDB()
	if err != nil {
		return nil, err
	}
	defer db.Close()

	rows, err := db.Query("SELECT * FROM users")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	users := make([]User, 0)
	for rows.Next() {
		var user User
		err := rows.Scan(&user.ID, &user.FirstName, &user.LastName, &user.DOB, &user.Email, &user.Phone, &user.CVPath)
		if err != nil {
			return nil, err
		}
		users = append(users, user)
	}

	return users, nil
}

func sendEmail(email string) error {

	fmt.Println("Sending email to:", email)
	return nil
}

func validatePhoneNumber(phone string) bool {
	regex := `^(?:\+91|0)?(?:6\d{9}|7\d{9}|8\d{9}|9\d{9})$`

	match, _ := regexp.MatchString(regex, phone)
	return match
}
