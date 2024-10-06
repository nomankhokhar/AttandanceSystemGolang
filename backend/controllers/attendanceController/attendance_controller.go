package attendanceController

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"

	"AttendanceSystem/controllers/userController"
	"AttendanceSystem/db"
)

// InsertAttendance handles the creation of a new attendance record
func InsertAttendance(c *gin.Context) {

	// Parse input data
	var input Attendance
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// check the user is authorized and exists
	var user *userController.User
	user, err := userController.FindUserByEmail(input.Email)

	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "email not found", "error_details": err.Error()})
		return
	}

	// Convert input date to time.Time
	date, err := time.Parse("2006-01-02", input.Date.Format("2006-01-02")) // Assuming date in format YYYY-MM-DD

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid date format"})
		return
	}

	// Create Attendance record
	newAttendance := Attendance{
		Email:          user.Email, // Use the logged-in user's email
		Date:           date,
		StartTime:      input.StartTime,
		FinishTime:     input.FinishTime,
		HoursNotWorked: input.HoursNotWorked,
		Reason:         input.Reason,
		Authorized:     input.Authorized,
		TimeToCatchUp:  input.TimeToCatchUp,
		CaughtUp:       input.CaughtUp,
		Sick:           input.Sick,
		TotalHours:     input.TotalHours,
		Task:           input.Task,
	}

	// Insert into database
	err = InsertAttendanceByEmail(&newAttendance)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to insert attendance record"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Attendance record added successfully", "record": newAttendance})
}

// GetAttendanceByEmail retrieves attendance records by email
func GetAttendanceByEmail(c *gin.Context) {
	// Get the logged-in user's email from context
    email := c.Query("email")

	if email == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "email is required"})
		return
	}
	
	// Fetch attendance records by email
	attendances, err := GetAttendanceEmail(email)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve attendance records"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"attendances": attendances})
}

// InsertAttendanceByEmail inserts a new attendance record for the provided email
func InsertAttendanceByEmail(attendance *Attendance) error {
	collection := db.GetDB().Collection("attendances")

	// Inserting attendance into the collection
	_, err := collection.InsertOne(context.Background(), attendance)
	if err != nil {
		return err
	}
	return nil
}

// GetAttendanceEmail retrieves attendance records for a specific email
func GetAttendanceEmail(email string) ([]Attendance, error) {
	var attendances []Attendance
	collection := db.GetDB().Collection("attendances")

	// Query the database for attendance records by email
	cursor, err := collection.Find(context.Background(), bson.M{"email": email})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(context.Background())

	// Decode the cursor results into the attendances slice
	if err = cursor.All(context.Background(), &attendances); err != nil {
		return nil, err
	}

	return attendances, nil
}
