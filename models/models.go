package models

import (
	"src/helpers"
	"time"
)

type Config struct {
	ID          int  `json:"id" gorm:"primaryKey"`
	LoginNotCmu bool `json:"loginNotCmu" gorm:"default:true;not null"`
}

type Subscription struct {
	FirstName string `json:"firstName" gorm:"primaryKey;size:100"`
	LastName  string `json:"lastName" gorm:"primaryKey;size:100"`
	Endpoint  string `json:"endpoint" gorm:"not null"`
	Auth      string `json:"auth" gorm:"not null"`
	P256dh    string `json:"p256dh" gorm:"not null"`
}

type Counter struct {
	ID         int     `json:"id" gorm:"primaryKey;autoIncrement"`
	Counter    string  `json:"counter" gorm:"unique;not null"`
	Status     bool    `json:"status" gorm:"default:false;not null"`
	TimeClosed string  `json:"timeClosed" gorm:"type:time(3);default:'16:00:00';not null"`
	User       *User   `json:"user" gorm:"foreignKey:CounterID;constraint:OnDelete:CASCADE"`
	Topics     []Topic `json:"topics" gorm:"many2many:counter_topics;constraint:OnDelete:CASCADE"`
}

type User struct {
	ID          int     `json:"id" gorm:"primaryKey;autoIncrement"`
	FirstNameTH *string `json:"firstNameTH" gorm:"size:100"`
	LastNameTH  *string `json:"lastNameTH" gorm:"size:100"`
	FirstNameEN *string `json:"firstNameEN" gorm:"size:100"`
	LastNameEN  *string `json:"lastNameEN" gorm:"size:100"`
	Email       string  `json:"email" gorm:"unique;size:100;not null"`
	CounterID   int     `json:"counterId" gorm:"foreignKey:CounterID;constraint:OnDelete:CASCADE"`
	Counter     Counter `json:"counter" gorm:"foreignKey:CounterID;constraint:OnDelete:CASCADE"`
}

type Topic struct {
	ID      int    `json:"id" gorm:"primaryKey;autoIncrement"`
	TopicTH string `json:"topicTH" gorm:"unique;not null"`
	TopicEN string `json:"topicEN" gorm:"unique;not null"`
	Code    string `json:"code" gorm:"unique;not null"`
}

type CounterTopic struct {
	CounterID int     `json:"counterId" gorm:"primaryKey;constraint:OnDelete:CASCADE"`
	TopicID   int     `json:"topicId" gorm:"primaryKey;constraint:OnDelete:CASCADE"`
	Counter   Counter `json:"counter" gorm:"foreignKey:CounterID;constraint:OnDelete:CASCADE"`
	Topic     Topic   `json:"topic" gorm:"foreignKey:TopicID;constraint:OnDelete:CASCADE"`
}

type Queue struct {
	ID        int            `json:"id" gorm:"primaryKey;autoIncrement"`
	No        string         `json:"no" gorm:"not null"`
	StudentID *string        `json:"studentId" gorm:"size:9"`
	Firstname string         `json:"firstName" gorm:"not null"`
	Lastname  string         `json:"lastName" gorm:"not null"`
	TopicID   int            `json:"topicId" gorm:"foreignKey:TopicID;constraint:OnDelete:CASCADE"`
	Topic     Topic          `json:"topic" gorm:"foreignKey:TopicID;constraint:OnDelete:CASCADE"`
	Note      *string        `json:"note" gorm:"size:255"`
	Status    helpers.STATUS `json:"status" gorm:"default:'WAITING';not null"`
	CounterID *int           `json:"counterId" gorm:"foreignKey:CounterID;constraint:OnDelete:CASCADE"`
	CreatedAt time.Time      `json:"createdAt" gorm:"default:current_timestamp"`
}

type Feedback struct {
	ID        int       `json:"id" gorm:"primaryKey;autoIncrement"`
	UserID    int       `json:"userId" gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE"`
	User      User      `json:"user" gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE"`
	TopicID   int       `json:"topicId" gorm:"foreignKey:TopicID;constraint:OnDelete:CASCADE"`
	Topic     Topic     `json:"topic" gorm:"foreignKey:TopicID;constraint:OnDelete:CASCADE"`
	Feedback  string    `json:"feedback" gorm:"not null"`
	CreatedAt time.Time `json:"createdAt" gorm:"default:current_timestamp"`
}

type UserWithoutCounter struct {
	ID          int     `json:"id"`
	FirstNameTH *string `json:"firstNameTH"`
	LastNameTH  *string `json:"lastNameTH"`
	FirstNameEN *string `json:"firstNameEN"`
	LastNameEN  *string `json:"lastNameEN"`
	Email       string  `json:"email"`
}

type CounterResponse struct {
	ID         int                `json:"id"`
	Counter    string             `json:"counter"`
	Status     bool               `json:"status"`
	TimeClosed string             `json:"timeClosed"`
	User       UserWithoutCounter `json:"user"`
	Topics     []Topic            `json:"topics"`
}
