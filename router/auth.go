package router

import (
	"containerized-go-app/hash"
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

func AuthRoutes(app *fiber.App, db *mongo.Database) {
	auth := app.Group("/auth", func(c *fiber.Ctx) error {
		return c.Next()
	})
	Login(db, auth)
	Register(db, auth)
}

func Login(db *mongo.Database, auth fiber.Router) {
	auth.Post("/login", func(c *fiber.Ctx) error {
		var loginRequest models.User
		if err := c.BodyParser(&loginRequest); err != nil || loginRequest.Email == "" || loginRequest.Password == "" {
			return c.Status(http.StatusBadRequest).JSON(fiber.Map{
				"ok":    false,
				"error": "Bad Request",
			})
		}

		userCollection := db.Collection("User")

		// get user from db
		existingUser := models.User{}
		err := userCollection.FindOne(context.Background(), bson.M{"email": loginRequest.Email}).Decode(&existingUser)
		if err != nil {
			return c.Status(http.StatusUnauthorized).JSON(fiber.Map{
				"ok":    false,
				"error": "wrong credentials",
			})
		}

		// Check if the password is correct
		if !hash.CheckPasswordHash(loginRequest.Password, existingUser.Password) {
			return c.Status(http.StatusUnauthorized).JSON(fiber.Map{
				"ok":    false,
				"error": "wrong password",
			})
		}

		// Generate JWT token
		userID := existingUser.ID.Hex()
		token := jwt.GetToken(userID)
		if token == "" {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"ok":    false,
				"error": "Internal Server Error",
			})
		}

		return c.Status(http.StatusOK).JSON(fiber.Map{
			"ok": true,
			"data": fiber.Map{
				"token": token,
				"user": fiber.Map{
					"email":     existingUser.Email,
					"firstName": existingUser.FirstName,
					"lastName":  existingUser.LastName,
				},
			},
		})
	})
}

func Register(db *mongo.Database, auth fiber.Router) {
	auth.Post("/register", func(c *fiber.Ctx) error {
		var user models.User

		// parse body, return 400 if invalid
		if err := c.BodyParser(&user); err != nil || user.Email == "" ||
			user.Password == "" || user.FirstName == "" || user.LastName == "" {
			return c.Status(http.StatusBadRequest).JSON(fiber.Map{
				"ok":    false,
				"error": "Bad Request",
			})
		}

		// get user collection
		userCollection := db.Collection("User")

		// Check if user with the same email already exists
		existingUser := models.User{}
		err := userCollection.FindOne(context.Background(), bson.M{"email": user.Email}).Decode(&existingUser)
		if err == nil {
			return c.Status(http.StatusUnauthorized).JSON(fiber.Map{
				"ok":    false,
				"error": "User with the same email already exists",
			})
		}

		// Hash the password
		hashPassword, err := hash.HashPassword(user.Password)
		if err != nil {
			return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
				"ok":    false,
				"error": "Internal server error when hashing password",
			})
		}
		user.Password = hashPassword

		user.CreatedAt = time.Now()
		user.LastUpVote = time.Now().Add(-1 * time.Minute)

		// save new user in db
		res, err := userCollection.InsertOne(context.Background(), user)
		if err != nil {
			return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
				"ok":    false,
				"error": "Internal Server Error",
			})
		}

		// Generate JWT token
		userID := res.InsertedID.(primitive.ObjectID).Hex()
		token := jwt.GetToken(userID)
		if token == "" {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"ok":    false,
				"error": "Internal Server Error",
			})
		}

		return c.Status(http.StatusCreated).JSON(fiber.Map{
			"ok": true,
			"data": fiber.Map{
				"token": token,
				"user": fiber.Map{
					"email":     user.Email,
					"firstName": user.FirstName,
					"lastName":  user.LastName,
				},
			},
		})

	})
}
