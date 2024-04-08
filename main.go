package main

import (
	"containerized-go-app/jwt"
	"containerized-go-app/models"
	"context"
	"errors"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"golang.org/x/crypto/bcrypt"
	"log"
	"net/http"
	"os"
	"time"
)

func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 11)
	return string(bytes), err
}

func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

func connectToDB() (*mongo.Database, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	uri := os.Getenv("MONGO_URI")
	if uri == "" {
		return nil, errors.New("MONGO_URI is not set")
	}

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		return nil, err
	}

	db := client.Database("keduback")

	return db, err
}

func main() {
	err := run()
	if err != nil {
		panic(err)
	}
}

func run() error {
	// load env variables
	err := godotenv.Load()
	if err != nil {
		return err
	}

	//init database
	db, err := connectToDB()
	if err != nil {
		log.Fatal(err)
	}

	//defer disconnect
	defer db.Client().Disconnect(context.Background())

	app := fiber.New()

	app.Use(logger.New())
	app.Use(recover.New())

	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("Hello, World!")
	})

	app.Post("/auth/register", func(c *fiber.Ctx) error {
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
		hash, err := HashPassword(user.Password)
		if err != nil {
			return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
				"ok":    false,
				"error": "Internal server error when hashing password",
			})
		}
		user.Password = hash

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

	app.Listen(":8080")

	return nil
}
