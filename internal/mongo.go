package internal

import (
	"chatapp/model"
	"context"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const DefaultDatabase = "chatapp"

const CollectionName = "users"

type MongoHandler struct {
	client   *mongo.Client
	database string
}

func NewHandler(address string) *MongoHandler {
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	cl, _ := mongo.Connect(ctx, options.Client().ApplyURI(address))
	mh := &MongoHandler{
		client:   cl,
		database: DefaultDatabase,
	}
	return mh
}

func (mh *MongoHandler) GetOne(c *model.User, filter interface{}) error {
	collection := mh.client.Database(mh.database).Collection(CollectionName)
	ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
	err := collection.FindOne(ctx, filter).Decode(c)
	return err
}

func (mh *MongoHandler) Get(filter interface{}) []*model.User {
	collection := mh.client.Database(mh.database).Collection(CollectionName)
	ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)

	cur, err := collection.Find(ctx, filter)
	if err != nil {
		log.Fatal(err)
	}
	defer cur.Close(ctx)

	var result []*model.User
	for cur.Next(ctx) {
		users := &model.User{}
		er := cur.Decode(users)
		if er != nil {
			log.Fatal(er)
		}
		result = append(result, users)
	}

	return result
}

func (mh *MongoHandler) AddOne(c *model.User) (*mongo.InsertOneResult, error) {
	collection := mh.client.Database(mh.database).Collection(CollectionName)
	ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
	result, err := collection.InsertOne(ctx, c)
	return result, err
}

func (mh *MongoHandler) Update(filter interface{}, update interface{}) (*mongo.UpdateResult, error) {
	collection := mh.client.Database(mh.database).Collection(CollectionName)
	ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
	result, err := collection.UpdateMany(ctx, filter, update)
	return result, err
}

func (mh *MongoHandler) RemoveOne(filter interface{}) (*mongo.DeleteResult, error) {
	collection := mh.client.Database(mh.database).Collection(CollectionName)
	ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)

	result, err := collection.DeleteOne(ctx, filter)
	return result, err
}
