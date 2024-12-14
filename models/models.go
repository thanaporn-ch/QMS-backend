package models

import (
	"src/helpers"
	"time"
)

type Subscription struct {
	StudentID string `json:"studentId"`
	Endpoint  string `json:"endpoint"`
	Auth      string `json:"auth"`
	P256dh    string `json:"p256dh"`
}

type Counter struct {
	ID         uint      `json:"id"`
	Counter    string    `json:"counter"`
	Status     bool      `json:"status"`
	TimeClosed time.Time `json:"timeClosed"`
}

type CounterWithUser struct {
	ID         int      `json:"id"`
	Counter    string   `json:"counter"`
	Status     bool     `json:"status"`
	TimeClosed string   `json:"timeClosed"`
	User       UserOnly `json:"user"`
}

type User struct {
	ID          uint    `json:"id"`
	FirstNameTH *string `json:"firstNameTH"`
	LastNameTH  *string `json:"lastNameTH"`
	FirstNameEN *string `json:"firstNameEN"`
	LastNameEN  *string `json:"lastNameEN"`
	Email       string  `json:"email"`
	CounterID   uint    `json:"counterId"`
	Counter     Counter `json:"counter"`
}

type UserOnly struct {
	ID          uint    `json:"id"`
	FirstNameTH *string `json:"firstNameTH"`
	LastNameTH  *string `json:"lastNameTH"`
	FirstNameEN *string `json:"firstNameEN"`
	LastNameEN  *string `json:"lastNameEN"`
	Email       string  `json:"email"`
}

type Topic struct {
	ID        uint    `json:"id"`
	Topic     string  `json:"topic"`
	CounterID uint    `json:"counterId"`
	Counter   Counter `json:"counter"`
}

type Queue struct {
	ID          uint           `json:"id"`
	No          string         `json:"no"`
	StudentID   string         `json:"studentId"`
	Firstname   string         `json:"firstName"`
	Lastname    string         `json:"lastName"`
	TopicID     uint           `json:"topicId"`
	Topic       Topic          `json:"topic"`
	Description string         `json:"description"`
	Status      helpers.STATUS `json:"status"`
	CreatedAt   time.Time      `json:"createdAt"`
}

type Feedback struct {
	ID        uint      `json:"id"`
	UserID    uint      `json:"userId"`
	User      User      `json:"user"`
	TopicID   uint      `json:"topicId"`
	Topic     Topic     `json:"topic"`
	Feedback  string    `json:"feedback"`
	CreatedAt time.Time `json:"createdAt"`
}
