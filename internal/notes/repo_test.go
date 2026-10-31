package notes

import (
	"context"
	"os"
	"testing"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Integration test. Set MONGO_URI to run it against a local MongoDB instance.
func TestCreateReadUpdateDelete(t *testing.T) {
	uri := os.Getenv("MONGO_URI")
	if uri == "" {
		t.Skip("MONGO_URI is not set; skipping MongoDB integration test")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = client.Disconnect(context.Background()) }()

	dbName := "pz8_test_" + time.Now().Format("20060102150405")
	database := client.Database(dbName)
	t.Cleanup(func() { _ = database.Drop(context.Background()) })

	repo, err := NewRepo(database)
	if err != nil {
		t.Fatal(err)
	}
	created, err := repo.Create(ctx, "Integration test note", "Initial content")
	if err != nil {
		t.Fatal(err)
	}
	got, err := repo.ByID(ctx, created.ID.Hex())
	if err != nil {
		t.Fatal(err)
	}
	if got.Title != "Integration test note" {
		t.Fatalf("unexpected title: %q", got.Title)
	}

	newContent := "Updated content"
	updated, err := repo.Update(ctx, created.ID.Hex(), nil, &newContent)
	if err != nil {
		t.Fatal(err)
	}
	if updated.Content != newContent {
		t.Fatalf("unexpected content: %q", updated.Content)
	}
	if err := repo.Delete(ctx, created.ID.Hex()); err != nil {
		t.Fatal(err)
	}
	if _, err := repo.ByID(ctx, created.ID.Hex()); err != ErrNotFound {
		t.Fatalf("expected ErrNotFound after deletion, got %v", err)
	}
}
