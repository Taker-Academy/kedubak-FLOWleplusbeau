package models

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
	"time"
)

type User struct {
	ID         primitive.ObjectID `bson:"_id,omitempty"`
	CreatedAt  time.Time          `bson:"createdAt,omitempty"`
	Email      string             `bson:"email,omitempty"`
	FirstName  string             `bson:"firstName,omitempty"`
	LastName   string             `bson:"lastName,omitempty"`
	Password   string             `bson:"password,omitempty"`
	LastUpVote time.Time          `bson:"lastUpVote,omitempty"`
}

type Post struct {
	ID        primitive.ObjectID `json:"_id" bson:"_id,omitempty"`
	CreatedAt time.Time          `json:"createdAt" bson:"createdAt,omitempty"`
	UserId    string             `json:"userId" bson:"userId,omitempty"`
	FirstName string             `json:"firstName" bson:"firstName,omitempty"`
	Title     string             `json:"title" bson:"title,omitempty"`
	Content   string             `json:"content" bson:"content,omitempty"`
	Comments  []Comment          `json:"comments" bson:"comments,omitempty"`
	UpVotes   []string           `json:"upVotes" bson:"upVotes,omitempty"`
}

type Comment struct {
	CreatedAt time.Time `json:"createdAt" bson:"createdAt,omitempty"`
	ID        string    `json:"id" bson:"id,omitempty"`
	FirstName string    `json:"firstName" bson:"firstName,omitempty"`
	Content   string    `json:"content" bson:"content,omitempty"`
}
