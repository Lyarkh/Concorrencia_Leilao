package auction

import (
	"context"
	"fullcycle-auction_go/internal/entity/auction_entity"
	"os"
	"testing"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func setupTestDB(t *testing.T) (*mongo.Database, func()) {
	t.Helper()

	mongoURL := os.Getenv("MONGODB_URL")
	if mongoURL == "" {
		mongoURL = "mongodb://admin:admin@mongodb:27017/auctions_test?authSource=admin"
	}

	ctx := context.Background()
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(mongoURL))
	if err != nil {
		t.Fatalf("failed to connect to mongodb: %v", err)
	}

	if err := client.Ping(ctx, nil); err != nil {
		t.Fatalf("failed to ping mongodb: %v", err)
	}

	db := client.Database("auctions_test")

	cleanup := func() {
		db.Drop(ctx)
		client.Disconnect(ctx)
	}

	return db, cleanup
}

func TestAuctionAutoClose(t *testing.T) {
	os.Setenv("AUCTION_DURATION", "2s")
	defer os.Unsetenv("AUCTION_DURATION")

	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewAuctionRepository(db)
	ctx := context.Background()

	auction, err := auction_entity.CreateAuction(
		"Test Product",
		"Tech",
		"Test auction for auto close verification",
		auction_entity.New,
	)
	if err != nil {
		t.Fatalf("failed to create auction entity: %v", err.Error())
	}

	if ierr := repo.CreateAuction(ctx, auction); ierr != nil {
		t.Fatalf("failed to insert auction: %v", ierr.Error())
	}

	// Verify auction is active right after creation
	var result AuctionEntityMongo
	filter := bson.M{"_id": auction.Id}
	if err := repo.Collection.FindOne(ctx, filter).Decode(&result); err != nil {
		t.Fatalf("failed to find auction: %v", err)
	}

	if result.Status != auction_entity.Active {
		t.Fatalf("expected status Active (%d), got %d", auction_entity.Active, result.Status)
	}

	// Wait for AUCTION_DURATION + buffer
	time.Sleep(3 * time.Second)

	// Verify auction is now closed
	var closed AuctionEntityMongo
	if err := repo.Collection.FindOne(ctx, filter).Decode(&closed); err != nil {
		t.Fatalf("failed to find auction after duration: %v", err)
	}

	if closed.Status != auction_entity.Completed {
		t.Fatalf("expected status Completed (%d), got %d", auction_entity.Completed, closed.Status)
	}
}

func TestGetAuctionDuration(t *testing.T) {
	tests := []struct {
		name     string
		envValue string
		expected time.Duration
	}{
		{"valid duration", "10s", 10 * time.Second},
		{"empty env", "", 5 * time.Minute},
		{"invalid env", "invalid", 5 * time.Minute},
		{"minutes", "2m", 2 * time.Minute},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Setenv("AUCTION_DURATION", tt.envValue)
			defer os.Unsetenv("AUCTION_DURATION")

			got := getAuctionDuration()
			if got != tt.expected {
				t.Errorf("getAuctionDuration() = %v, want %v", got, tt.expected)
			}
		})
	}
}
