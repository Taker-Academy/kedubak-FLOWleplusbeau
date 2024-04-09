package router

import (
	"containerized-go-app/jwt"
	"containerized-go-app/models"
	"context"
	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"net/http"
	"time"
)

func PostRoutes(app *fiber.App, db *mongo.Database) {
	post := app.Group("/post", func(c *fiber.Ctx) error {
		return c.Next()
	})
	GetPosts(db, post)
	GetMyPosts(db, post)
	CreatePost(db, post)
}

func CreatePost(db *mongo.Database, post fiber.Router) {
	post.Post("/", func(c *fiber.Ctx) error {
		// get user id and authorization from token
		userId, err := jwt.GetUserID(c.Get("Authorization"), db.Client())
		if err != nil {
			return c.Status(http.StatusUnauthorized).JSON(fiber.Map{
				"ok":    false,
				"error": "wrong token",
			})
		}

		// get post request
		var postRequest models.Post
		if err := c.BodyParser(&postRequest); err != nil || postRequest.Title == "" || postRequest.Content == "" {
			return c.Status(http.StatusBadRequest).JSON(fiber.Map{
				"ok":    false,
				"error": "Bad Request",
			})
		}

		// get user from db
		userCollection := db.Collection("User")
		objId, _ := primitive.ObjectIDFromHex(userId)
		user := models.User{}
		_ = userCollection.FindOne(context.Background(), bson.M{"_id": objId}).Decode(&user)

		postCollection := db.Collection("Post")

		// create new post
		newPost := models.Post{
			CreatedAt: time.Now(),
			UserId:    userId,
			FirstName: user.FirstName,
			Title:     postRequest.Title,
			Content:   postRequest.Content,
			Comments:  []models.Comment{},
			UpVotes:   []string{},
		}

		// insert post to db
		_, err = postCollection.InsertOne(context.Background(), newPost)
		if err != nil {
			return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
				"ok":    false,
				"error": "Internal Server Error",
			})
		}

		return c.Status(http.StatusCreated).JSON(fiber.Map{
			"ok": true,
			"data": fiber.Map{
				"createdAt": newPost.CreatedAt,
				"userId":    newPost.UserId,
				"firstName": newPost.FirstName,
				"title":     newPost.Title,
				"content":   newPost.Content,
				"comments":  newPost.Comments,
				"upVotes":   newPost.UpVotes,
			},
		})
	})
}

func GetPosts(db *mongo.Database, post fiber.Router) {
	post.Get("/", func(c *fiber.Ctx) error {
		_, err := jwt.GetUserID(c.Get("Authorization"), db.Client())
		if err != nil {
			return c.Status(http.StatusUnauthorized).JSON(fiber.Map{
				"ok":    false,
				"error": "wrong token",
			})
		}

		postCollection := db.Collection("Post")
		//get all posts
		cursor, err := postCollection.Find(context.Background(), bson.M{})
		if err != nil {
			return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
				"ok":    false,
				"error": "Internal Server Error",
			})
		}
		posts := []models.Post{}
		if err = cursor.All(context.Background(), &posts); err != nil {
			return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
				"ok":    false,
				"error": "Internal Server Error",
			})
		}

		// change nil array to empty array
		for i, post := range posts {
			if post.UpVotes == nil {
				posts[i].UpVotes = []string{}
			}
			if post.Comments == nil {
				posts[i].Comments = []models.Comment{}
			}
		}
		return c.Status(http.StatusOK).JSON(fiber.Map{
			"ok":   true,
			"data": posts,
		})
	})
}

func GetMyPosts(db *mongo.Database, post fiber.Router) {
	post.Get("/me", func(c *fiber.Ctx) error {
		userID, err := jwt.GetUserID(c.Get("Authorization"), db.Client())
		if err != nil {
			return c.Status(http.StatusUnauthorized).JSON(fiber.Map{
				"ok":    false,
				"error": "wrong token",
			})
		}

		postCollection := db.Collection("Post")
		//get all posts of the user
		cursor, err := postCollection.Find(context.Background(), bson.M{"userId": userID})
		if err != nil {
			return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
				"ok":    false,
				"error": "Internal Server Error",
			})
		}
		posts := []models.Post{}
		if err = cursor.All(context.Background(), &posts); err != nil {
			return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
				"ok":    false,
				"error": "Internal Server Error",
			})
		}

		// change nil array to empty array
		for i, post := range posts {
			if post.UpVotes == nil {
				posts[i].UpVotes = []string{}
			}
			if post.Comments == nil {
				posts[i].Comments = []models.Comment{}
			}
		}

		return c.Status(http.StatusOK).JSON(fiber.Map{
			"ok":   true,
			"data": posts,
		})
	})
}
