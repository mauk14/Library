package data

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"time"
)

type MongoDb struct {
	DB *mongo.Database
}

func NewMongo(client *mongo.Client, name string) *MongoDb {
	return &MongoDb{DB: client.Database(name)}
}

func (m *MongoDb) Insert(ctx context.Context, _ string, collection string, data interface{}) error {
	coll := m.DB.Collection(collection)
	switch data.(type) {
	case *Token:
		token := data.(*Token)
		_, err := coll.InsertOne(ctx, bson.M{
			"user_id": token.UserID,
			"expiry":  token.Expiry,
			"scope":   token.Scope,
			"hash":    token.Hash,
		})
		return err
	case *User:
		user := data.(*User)
		_, err := coll.InsertOne(ctx, bson.M{
			"id":         user.ID,
			"created_at": user.CreatedAt,
			"password":   user.Password.hash,
			"email":      user.Email,
			"activated":  user.Activated,
			"version":    user.Version,
		})
		fmt.Println(err)
		return err
	}
	_, err := coll.InsertOne(ctx, data)
	return err
}

func (m *MongoDb) Get(ctx context.Context, _ string, id interface{}, collection string, scope string) (interface{}, error) {
	coll := m.DB.Collection(collection)

	if collection == "books" {
		var result Book

		err := coll.FindOne(ctx, bson.D{{"id", id}}).Decode(&result)
		if err != nil {
			return nil, err
		}

		return result, nil
	} else if collection == "users" {
		var result User
		input := struct {
			ID        int64     `json:"id"`
			CreatedAt time.Time `json:"created_at"`
			Name      string    `json:"name"`
			Email     string    `json:"email"`
			Password  []byte    `json:"-"`
			Activated bool      `json:"activated"`
			Version   uuid.UUID `json:"-"`
		}{}

		switch id.(type) {
		case string:
			err := coll.FindOne(ctx, bson.D{{"email", id}}).Decode(&input)
			if err != nil {
				return nil, err
			}
			result.ID = input.ID
			result.CreatedAt = input.CreatedAt
			result.Name = input.Name
			result.Email = input.Email
			result.Password.hash = input.Password
			result.Activated = input.Activated
			result.Version = input.Version

			return result, nil
		}
		err := coll.FindOne(ctx, bson.D{{"id", id}}).Decode(&result)
		if err != nil {
			return nil, err
		}

		return result, nil
	} else if collection == "tokens" {
		input := struct {
			ID        int64     `json:"id"`
			CreatedAt time.Time `json:"created_at"`
			Name      string    `json:"name"`
			Email     string    `json:"email"`
			Password  []byte    `json:"-"`
			Activated bool      `json:"activated"`
			Version   uuid.UUID `json:"-"`
		}{}
		var result User
		var token Token
		var res bson.M

		opt := options.FindOne().SetProjection(bson.D{{"_id", 0}, {"user_id", 1}})
		err := coll.FindOne(ctx, bson.M{"hash": id, "scope": scope}, opt).Decode(&res)
		//fmt.Println(res)
		//fmt.Println(err)

		jsonData, err := json.MarshalIndent(res, "", "    ")
		if err != nil {
			return nil, err
		}

		err = json.Unmarshal(jsonData, &token)
		//fmt.Println(err)
		if err != nil {
			return nil, err
		}
		//fmt.Println(token.UserID)

		coll = m.DB.Collection("users")
		err = coll.FindOne(ctx, bson.M{"id": token.UserID}).Decode(&input)
		//fmt.Println(err)

		result.ID = input.ID
		result.CreatedAt = input.CreatedAt
		result.Name = input.Name
		result.Email = input.Email
		result.Password.hash = input.Password
		result.Activated = input.Activated
		result.Version = input.Version

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
	} else if collection == "users" {
		input := struct {
			ID        int64     `json:"id"`
			CreatedAt time.Time `json:"created_at"`
			Name      string    `json:"name"`
			Email     string    `json:"email"`
			Password  []byte    `json:"-"`
			Activated bool      `json:"activated"`
			Version   uuid.UUID `json:"-"`
		}{}
		err := coll.FindOne(ctx, filter, opts).Decode(&input)
		if err != nil {
			return 0, err
		}
		return input.ID, nil
	}
	return 0, errors.New("No collections in database")

}

func (m *MongoDb) GetAll(ctx context.Context, _ string, collection string, opt map[string]string, genres []string, filters Filters) ([]*Book, Metadata, error) {
	coll := m.DB.Collection(collection)
	var direct int
	if filters.sortDirection() == "ASC" {
		direct = 1
	} else {
		direct = -1
	}
	opts := options.Find().SetSort(bson.D{{filters.sortColumn(), direct}}).SetSkip(int64(filters.offset())).SetLimit(int64(filters.limit()))
	totalRecords, err := coll.CountDocuments(ctx, bson.D{})
	if err != nil {
		return nil, Metadata{}, err
	}
	result := make([]*Book, 0, 5)
	var cursor *mongo.Cursor

	if opt["title"] != "" && opt["author"] != "" && len(genres) != 0 {
		cursor, err = coll.Find(ctx, bson.M{
			"title":  bson.M{"$regex": opt["title"], "$options": "i"},
			"author": bson.M{"$regex": opt["author"], "$options": "i"},
			"genres": bson.D{{"$all", genres}},
		}, opts)
	} else if opt["title"] != "" && opt["author"] != "" {
		cursor, err = coll.Find(ctx, bson.D{{"title", bson.M{"$regex": opt["title"], "$options": "i"}}, {"author", bson.M{"$regex": opt["author"], "$options": "i"}}}, opts)
	} else if opt["title"] != "" && len(genres) != 0 {
		cursor, err = coll.Find(ctx, bson.D{{"title", bson.M{"$regex": opt["title"], "$options": "i"}}, {"genres", bson.D{{"$all", genres}}}}, opts)
	} else if opt["author"] != "" && len(genres) != 0 {
		cursor, err = coll.Find(ctx, bson.D{{"author", bson.M{"$regex": opt["author"], "$options": "i"}}, {"genres", bson.D{{"$all", genres}}}}, opts)
	} else if opt["title"] != "" {
		cursor, err = coll.Find(ctx, bson.D{{"title", bson.M{"$regex": opt["title"], "$options": "i"}}}, opts)
	} else if opt["author"] != "" {
		cursor, err = coll.Find(ctx, bson.D{{"author", bson.M{"$regex": opt["author"], "$options": "i"}}}, opts)
	} else if len(genres) != 0 {
		cursor, err = coll.Find(ctx, bson.D{{"genres", bson.D{{"$all", genres}}}}, opts)
	} else {
		cursor, err = coll.Find(ctx, bson.D{}, opts)
	}

	if err != nil {
		return nil, Metadata{}, err
	}

	for cursor.Next(ctx) {
		var book *Book
		if err = cursor.Decode(&book); err != nil {
			return nil, Metadata{}, err
		}
		result = append(result, book)
	}
	if err = cursor.Err(); err != nil {
		return nil, Metadata{}, err
	}

	metadata := calculateMetadata(int(totalRecords), filters.Page, filters.PageSize)

	return result, metadata, nil

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
	case *User:
		coll := m.DB.Collection("users")
		user := data.(*User)
		filter := bson.D{{"id", user.ID}, {"version", user.Version}}
		user.Version = uuid.New()

		update := bson.D{
			{"$set", bson.D{
				{"name", user.Name},
				{"email", user.Email},
				{"password", user.Password.hash},
				{"activated", user.Activated},
				{"version", user.Version},
			}}}
		_, err := coll.UpdateOne(ctx, filter, update)
		return err
	}
	return errors.New("No collections in database")

}

func (m *MongoDb) Delete(ctx context.Context, _ string, id int64, collection string, scope string) error {

	coll := m.DB.Collection(collection)
	if collection == "tokens" {
		_, err := coll.DeleteOne(ctx, bson.D{{"user_id", id}, {"scope", scope}})
		if err != nil {
			return err
		}
		return nil
	}

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
