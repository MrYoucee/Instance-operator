package main

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/alecthomas/kong"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/event"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"riskident.com/platform/gocrud/models"
)

var cli struct {
	MongoUri string `default:"mongodb://localhost:27017" help:"Mongodb uri to use" env:"MONGOURI"`
	DbName   string `default:"servers" help:"Mongodb database to use" env:"MONGODB"`
	ColName  string `default:"servers" help:"Mongodb collection to use" env:"COLNAME"`
}

func mongoConnect(mongouri string) (*mongo.Client, error) {
	log.WithFields(log.Fields{"context": "mongodb", "mongouri": mongouri}).Info("Connecting to mongodb")
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	cmdMonitor := &event.CommandMonitor{
		Started: func(_ context.Context, evt *event.CommandStartedEvent) {
			fmt.Printf("mongo: %+v\n", evt)
		},
	}
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(mongouri).SetMonitor(cmdMonitor))
	if err != nil {
		return client, err
	}
	err = client.Ping(ctx, readpref.Primary())
	return client, err
}
func mongoDisconnect(mongoClient *mongo.Client) error {
	log.WithFields(log.Fields{"context": "mongodb"}).Debug("Disconnecting from mongo")
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	return mongoClient.Disconnect(ctx)
}

func main() {
	kong.Parse(&cli)
	mongoClient, err := mongoConnect(cli.MongoUri)
	if err != nil {
		panic(err)
	}
	defer mongoDisconnect(mongoClient)

	collection := mongoClient.Database(cli.DbName).Collection(cli.ColName)

	router := gin.Default()
	v1 := router.Group("/v1")
	{
		v1.POST("/server", func(c *gin.Context) {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			var server models.Server
			if err := c.ShouldBindJSON(&server); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}
			res, err := collection.InsertOne(ctx, server)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}
			c.JSON(http.StatusOK, gin.H{"result": res})
		})

		v1.GET("/server/:id", func(c *gin.Context) {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			id := c.Param("id")
			var server models.Server
			objectId, err := primitive.ObjectIDFromHex(id)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}
			if err := collection.FindOne(ctx, bson.D{{"_id", objectId}}).Decode(&server); err != nil {
				c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
				return
			}
			c.JSON(http.StatusOK, gin.H{"result": server})
		})
	}
	router.Run()
}
