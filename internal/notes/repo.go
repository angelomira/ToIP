package notes

import (
	"context"
	"errors"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var ErrNotFound = errors.New("note not found")

type Repo struct {
	col *mongo.Collection
}

func NewRepo(database *mongo.Database) (*Repo, error) {
	collection := database.Collection("notes")
	indexCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_, err := collection.Indexes().CreateOne(indexCtx, mongo.IndexModel{
		Keys:    bson.D{{Key: "title", Value: 1}},
		Options: options.Index().SetUnique(true).SetName("uniq_note_title"),
	})
	if err != nil {
		return nil, err
	}
	return &Repo{col: collection}, nil
}

func (r *Repo) Create(ctx context.Context, title, content string) (Note, error) {
	now := time.Now().UTC()
	note := Note{Title: title, Content: content, CreatedAt: now, UpdatedAt: now}
	result, err := r.col.InsertOne(ctx, note)
	if err != nil {
		return Note{}, err
	}
	id, ok := result.InsertedID.(primitive.ObjectID)
	if !ok {
		return Note{}, errors.New("MongoDB returned an unexpected identifier type")
	}
	note.ID = id
	return note, nil
}

func (r *Repo) ByID(ctx context.Context, idHex string) (Note, error) {
	id, err := primitive.ObjectIDFromHex(idHex)
	if err != nil {
		return Note{}, ErrNotFound
	}
	var note Note
	err = r.col.FindOne(ctx, bson.M{"_id": id}).Decode(&note)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return Note{}, ErrNotFound
	}
	if err != nil {
		return Note{}, err
	}
	return note, nil
}

func (r *Repo) List(ctx context.Context, query string, limit, skip int64) ([]Note, error) {
	filter := bson.M{}
	if query != "" {
		filter["title"] = bson.M{"$regex": query, "$options": "i"}
	}
	findOptions := options.Find().
		SetLimit(limit).
		SetSkip(skip).
		SetSort(bson.D{{Key: "createdAt", Value: -1}})

	cursor, err := r.col.Find(ctx, filter, findOptions)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	notes := make([]Note, 0)
	for cursor.Next(ctx) {
		var note Note
		if err := cursor.Decode(&note); err != nil {
			return nil, err
		}
		notes = append(notes, note)
	}
	if err := cursor.Err(); err != nil {
		return nil, err
	}
	return notes, nil
}

func (r *Repo) Update(ctx context.Context, idHex string, title, content *string) (Note, error) {
	id, err := primitive.ObjectIDFromHex(idHex)
	if err != nil {
		return Note{}, ErrNotFound
	}
	set := bson.M{"updatedAt": time.Now().UTC()}
	if title != nil {
		set["title"] = *title
	}
	if content != nil {
		set["content"] = *content
	}

	updateOptions := options.FindOneAndUpdate().SetReturnDocument(options.After)
	var updated Note
	err = r.col.FindOneAndUpdate(ctx, bson.M{"_id": id}, bson.M{"$set": set}, updateOptions).Decode(&updated)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return Note{}, ErrNotFound
	}
	if err != nil {
		return Note{}, err
	}
	return updated, nil
}

func (r *Repo) Delete(ctx context.Context, idHex string) error {
	id, err := primitive.ObjectIDFromHex(idHex)
	if err != nil {
		return ErrNotFound
	}
	result, err := r.col.DeleteOne(ctx, bson.M{"_id": id})
	if err != nil {
		return err
	}
	if result.DeletedCount == 0 {
		return ErrNotFound
	}
	return nil
}

// GetStats demonstrates an aggregation pipeline: count documents and average content length.
func (r *Repo) GetStats(ctx context.Context) (Stats, error) {
	pipeline := mongo.Pipeline{
		{{Key: "$group", Value: bson.M{
			"_id":                  nil,
			"count":                bson.M{"$sum": 1},
			"averageContentLength": bson.M{"$avg": bson.M{"$strLenCP": "$content"}},
		}}},
	}
	cursor, err := r.col.Aggregate(ctx, pipeline)
	if err != nil {
		return Stats{}, err
	}
	defer cursor.Close(ctx)
	var results []Stats
	if err := cursor.All(ctx, &results); err != nil {
		return Stats{}, err
	}
	if len(results) == 0 {
		return Stats{}, nil
	}
	return results[0], nil
}
