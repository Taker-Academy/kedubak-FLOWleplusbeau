package router

import (
	"containerized-go-app/jwt"
	"containerized-go-app/models"
	"context"
	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"net/http"
)

func PostRoutes(app *fiber.App, db *mongo.Database) {
	post := app.Group("/post", func(c *fiber.Ctx) error {
		return c.Next()
	})
	GetPosts(db, post)
	GetMyPosts(db, post)
}

func GetPosts(db *mongo.Database, post fiber.Router) {
	post.Get("/", func(c *fiber.Ctx) error {
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
		// Get the token from the header Authorization
		token := c.Get("Authorization")
		userID, err := jwt.GetUserID(token, db.Client())
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
