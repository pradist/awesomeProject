package main

import (
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Entry struct {
	Entries []*GuestBook `bson:"entries" json:"entries"`
}

type GuestBook struct {
	ID       primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Name     string             `bson:"name" json:"name"`
	Email    string             `bson:"email" json:"email"`
	Comment  string             `bson:"comment" json:"comment"`
	Icon     string             `bson:"icon" json:"icon"`
	CreateAt time.Time          `bson:"createAt" json:"createAt"`
}

func main() {
	fmt.Printf("GREETING : %s \n", os.Getenv("DEMO_GREETING"))
	clientOptions := options.Client().ApplyURI(os.Getenv("DATABASE_URL"))
	//clientOptions := options.Client().ApplyURI("mongodb://localhost:27017")
	client, err := mongo.Connect(context.TODO(), clientOptions)

	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}

	err = client.Ping(context.TODO(), nil)

	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Connected to MongoDB!")

	coll := client.Database("guestbook").Collection("guestbook")

	r := gin.Default()
	config := cors.DefaultConfig()
	config.AllowOrigins = []string{"*"}
	r.Use(cors.New(config))

	v1 := r.Group("/guestbook")
	{
		v1.GET("/entries", wrapError(coll, listGuestBookHandler))
		v1.PUT("/entries", wrapError(coll, createGuestBookHandler))
		v1.DELETE("/entries/:id", wrapError(coll, deleteGuestBookHandler))
	}

	r.Run(":8000")
}

func wrapError(coll *mongo.Collection, h func(context.Context, *gin.Context, *mongo.Collection) error) func(*gin.Context) {
	return func(c *gin.Context) {
		ctx := c.Request.Context()
		err := h(ctx, c, coll)
		if err != nil {
			if err == mongo.ErrNoDocuments {
				c.AbortWithError(http.StatusNotFound, err)
			} else {
				c.AbortWithError(http.StatusInternalServerError, err)
			}
		}
	}
}

func listGuestBookHandler(ctx context.Context, c *gin.Context, coll *mongo.Collection) error {
	guestBooks, err := listGuestBook(ctx, coll)
	if err != nil {
		return err
	}
	c.JSON(http.StatusOK, guestBooks)
	return nil
}

func listGuestBook(ctx context.Context, coll *mongo.Collection) (*Entry, error) {
	cur, err := coll.Find(ctx, bson.D{})
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)
	guestBookList := []*GuestBook{}

	for cur.Next(ctx) {
		book := newGuestBook()
		if err := cur.Decode(book); err != nil {
			return nil, err
		}
		guestBookList = append(guestBookList, book)
	}
	if err := cur.Err(); err != nil {
		return nil, err
	}

	entry := &Entry{}
	entry.Entries = guestBookList
	return entry, nil
}

func newGuestBook() *GuestBook {
	return &GuestBook{}
}

func createGuestBookHandler(ctx context.Context, c *gin.Context, coll *mongo.Collection) error {
	guestBooks, err := createGuestBook(ctx, c, coll)
	if err != nil {
		return err
	}
	c.JSON(http.StatusOK, guestBooks)
	return nil
}

func createGuestBook(ctx context.Context, c *gin.Context, coll *mongo.Collection) (*GuestBook, error) {
	guestBook := newGuestBook()
	if err := c.Bind(&guestBook); err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return nil, err
	}
	guestBook.CreateAt = time.Now()
	result, err := coll.InsertOne(ctx, guestBook)
	if err != nil {
		return nil, err
	}
	guestBook.ID = result.InsertedID.(primitive.ObjectID)
	return guestBook, nil
}

func deleteGuestBookHandler(ctx context.Context, c *gin.Context, coll *mongo.Collection) error {
	err := deleteGuestBook(ctx, c, coll)
	if err != nil {
		return err
	}
	c.JSON(http.StatusOK, "")
	return nil
}

func deleteGuestBook(ctx context.Context, c *gin.Context, coll *mongo.Collection) error {
	id, _ := primitive.ObjectIDFromHex(c.Param("id"))
	_, err := coll.DeleteOne(ctx, bson.D{{"_id", id}})
	if err != nil {
		return err
	}
	return nil
}
