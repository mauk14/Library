package data

import (
	"context"
	"errors"
	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type MongoDb struct {
	DB *mongo.Database
}

func NewMongo(client *mongo.Client, name string) *MongoDb {
	return &MongoDb{DB: client.Database(name)}
}

func (m *MongoDb) Insert(ctx context.Context, _ string, collection string, data interface{}) error {
	coll := m.DB.Collection(collection)
	_, err := coll.InsertOne(ctx, data)
	return err
}

func (m *MongoDb) Get(ctx context.Context, _ string, id int64, collection string) (interface{}, error) {
	coll := m.DB.Collection(collection)
	if collection == "books" {
		var result Book

		err := coll.FindOne(ctx, bson.D{{"id", id}}).Decode(&result)
		if err != nil {
			return nil, err
		}

		return result, nil
	}

	return nil, errors.New("No collections in database")
}

func (m *MongoDb) GetLastId(ctx context.Context, _ string, collection string) (int64, error) {
	coll := m.DB.Collection(collection)
	filter := bson.D{}
	opts := options.FindOne().SetSort(bson.D{{"id", -1}})

	if collection == "books" {
		var result Book
		err := coll.FindOne(ctx, filter, opts).Decode(&result)
		if err != nil {
			return 0, err
		}
		return result.ID, nil
	}
	return 0, errors.New("No collections in database")

}

func (m *MongoDb) GetAll(ctx context.Context, _ string, id int64, collection string) ([]interface{}, error) {
	return nil, nil
}

func (m *MongoDb) Update(ctx context.Context, _ string, data interface{}) error {
	switch data.(type) {
	case *Book:
		coll := m.DB.Collection("books")
		book := data.(*Book)
		filter := bson.D{{"id", book.ID}, {"version", book.Version}}
		book.Version = uuid.New()
		update := bson.D{
			{"$set", bson.D{
				{"title", book.Title},
				{"author", book.Author},
				{"size", book.Size},
				{"year", book.Year},
				{"genres", book.Genres},
				{"version", book.Version},
			}}}

		_, err := coll.UpdateOne(ctx, filter, update)
		return err
	}
	return errors.New("No collections in database")

}

func (m *MongoDb) Delete(ctx context.Context, _ string, id int64, collection string) error {
	coll := m.DB.Collection(collection)
	result, err := coll.DeleteOne(ctx, bson.D{{"id", id}})
	if err != nil {
		return err
	}
	rowsAffected := result.DeletedCount
	if rowsAffected == 0 {
		return ErrRecordNotFound
	}
	return nil
}
