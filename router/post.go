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
	GetPostById(db, post)
	CreatePost(db, post)
	DeletePostById(db, post)
	VotePostById(db, post)
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

		_, err = userCollection.UpdateOne(context.Background(), bson.M{"_id": objId}, bson.M{"$set": bson.M{"comments": []models.Comment{}}})
		if err != nil {
			return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
				"ok":    false,
				"error": "Internal Server Error",
			})
		}

		_, err = userCollection.UpdateOne(context.Background(), bson.M{"_id": objId}, bson.M{"$set": bson.M{"upVotes": []string{}}})
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

		// get all posts
		postCollection := db.Collection("Post")
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

func GetPostById(db *mongo.Database, post fiber.Router) {
	post.Get("/:id", func(c *fiber.Ctx) error {
		// get user id and authorization from token
		_, err := jwt.GetUserID(c.Get("Authorization"), db.Client())
		if err != nil {
			return c.Status(http.StatusUnauthorized).JSON(fiber.Map{
				"ok":    false,
				"error": "wrong token",
			})
		}

		// get post by id
		postCollection := db.Collection("Post")
		postId := c.Params("id")
		objId, _ := primitive.ObjectIDFromHex(postId)
		post := models.Post{}

		// get post from db
		err = postCollection.FindOne(context.Background(), bson.M{"_id": objId}).Decode(&post)
		if err != nil {
			return c.Status(http.StatusNotFound).JSON(fiber.Map{
				"ok":    false,
				"error": "Post not found",
			})
		}

		// change nil array to empty array
		if post.UpVotes == nil {
			post.UpVotes = []string{}
		}
		if post.Comments == nil {
			post.Comments = []models.Comment{}
		}

		return c.Status(http.StatusOK).JSON(fiber.Map{
			"ok": true,
			"data": fiber.Map{
				"createdAt": post.CreatedAt,
				"userId":    post.UserId,
				"firstName": post.FirstName,
				"title":     post.Title,
				"content":   post.Content,
				"comments":  post.Comments,
				"upVotes":   post.UpVotes,
			},
		})
	})
}

func DeletePostById(db *mongo.Database, post fiber.Router) {
	post.Delete("/:id", func(c *fiber.Ctx) error {
		// get user id and authorization from token
		UserId, err := jwt.GetUserID(c.Get("Authorization"), db.Client())
		if err != nil {
			return c.Status(http.StatusUnauthorized).JSON(fiber.Map{
				"ok":    false,
				"error": "wrong token",
			})
		}

		// get post by id
		postCollection := db.Collection("Post")
		postId := c.Params("id")
		objId, _ := primitive.ObjectIDFromHex(postId)
		post := models.Post{}

		// get post from db
		err = postCollection.FindOne(context.Background(), bson.M{"_id": objId}).Decode(&post)
		if err != nil {
			return c.Status(http.StatusNotFound).JSON(fiber.Map{
				"ok":    false,
				"error": "Post not found",
			})
		}

		// check if the post belongs to the user
		if post.UserId != UserId {
			return c.Status(http.StatusForbidden).JSON(fiber.Map{
				"ok":    false,
				"error": "Forbidden",
			})
		}

		// delete post
		_, err = postCollection.DeleteOne(context.Background(), bson.M{"_id": objId})
		if err != nil {
			return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
				"ok":    false,
				"error": "Internal Server Error",
			})
		}

		return c.Status(http.StatusOK).JSON(fiber.Map{
			"ok": true,
			"message": fiber.Map{
				"createdAt": post.CreatedAt,
				"userId":    post.UserId,
				"firstName": post.FirstName,
				"title":     post.Title,
				"content":   post.Content,
				"comments":  post.Comments,
				"upVotes":   post.UpVotes,
				"removed":   true,
			},
		})
	})
}

func VotePostById(db *mongo.Database, post fiber.Router) {
	post.Post("/vote/:id", func(c *fiber.Ctx) error {
		// get user id and authorization from token
		UserId, err := jwt.GetUserID(c.Get("Authorization"), db.Client())
		if err != nil {
			return c.Status(http.StatusUnauthorized).JSON(fiber.Map{
				"ok":    false,
				"error": "wrong token",
			})
		}

		// get post by id
		postCollection := db.Collection("Post")
		postId := c.Params("id")
		objId, _ := primitive.ObjectIDFromHex(postId)
		post := models.Post{}

		// check if id is valid
		if objId.IsZero() {
			return c.Status(http.StatusUnprocessableEntity).JSON(fiber.Map{
				"ok":    false,
				"error": "Invalid ID",
			})
		}

		// get post from db
		err = postCollection.FindOne(context.Background(), bson.M{"_id": objId}).Decode(&post)
		if err != nil {
			return c.Status(http.StatusNotFound).JSON(fiber.Map{
				"ok":    false,
				"error": "Post not found",
			})
		}

		//get user from db
		userCollection := db.Collection("User")
		objId, _ = primitive.ObjectIDFromHex(UserId)
		user := models.User{}
		err = userCollection.FindOne(context.Background(), bson.M{"_id": objId}).Decode(&user)
		if err != nil {
			return c.Status(http.StatusNotFound).JSON(fiber.Map{
				"ok":    false,
				"error": "User not found",
			})
		}

		// check if the user has already voted
		for _, upVote := range post.UpVotes {
			if upVote == UserId {
				return c.Status(http.StatusConflict).JSON(fiber.Map{
					"ok":    false,
					"error": "User has already voted",
				})
			}
		}

		//check if the user has voted in the last 1 minutes
		if time.Now().Sub(user.LastUpVote) < time.Minute {
			return c.Status(http.StatusForbidden).JSON(fiber.Map{
				"ok":    false,
				"error": "User has voted in the last 1 minute",
			})
		}

		// create upVotes field in post in db
		if post.UpVotes == nil {
			_, err = postCollection.UpdateOne(context.Background(), bson.M{"_id": objId}, bson.M{"$set": bson.M{"upVotes": []string{"id"}}})
			if err != nil {
				return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
					"ok":    false,
					"error": "Internal Server Error",
				})
			}
		} else {
			_, err = postCollection.UpdateOne(context.Background(), bson.M{"_id": objId}, bson.M{"$push": bson.M{"upVotes": UserId}})
			if err != nil {
				return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
					"ok":    false,
					"error": "Internal Server Error",
				})
			}
		}

		// reset user vote time
		_, err = userCollection.UpdateOne(context.Background(), bson.M{"_id": objId}, bson.M{"$set": bson.M{"lastUpVote": time.Now()}})
		if err != nil {
			return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
				"ok":    false,
				"error": "Internal Server Error",
			})
		}

		return c.Status(http.StatusOK).JSON(fiber.Map{
			"ok":      true,
			"message": "post upvoted",
		})
	})
}
