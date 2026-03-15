package tests

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	testcontainers "github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/mongodb"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
	"go.mongodb.org/mongo-driver/v2/mongo/options"

	"go.mongodb.org/mongo-driver/v2/mongo"
)

var (
	postgresClient *sql.DB
	mongoClient    *mongo.Client
)

func TestMain(m *testing.M) {
	ctx := context.Background()

	postgresC, err := postgres.Run(ctx, "postgres:latest",
		testcontainers.WithWaitStrategy(wait.ForLog("database system is ready to accept connections").WithOccurrence(2)),
	)
	if err != nil {
		panic(err)
	}
	postgresConnStr, err := postgresC.ConnectionString(ctx)
	if err != nil {
		panic(err)
	}
	fmt.Printf("PostgreSQL connection string: %s\n", postgresConnStr)
	postgresClient, err = pg(ctx, postgresConnStr)
	if err != nil {
		panic(err)
	}

	mongoC, err := mongodb.Run(ctx, "mongo:latest",
		testcontainers.WithWaitStrategy(wait.ForLog("Waiting for connections")),
	)
	if err != nil {
		panic(err)
	}
	mongoConnStr, err := mongoC.ConnectionString(ctx)
	if err != nil {
		panic(err)
	}
	mongoConnStr += "?directConnection=true"
	fmt.Printf("MongoDB connection string: %s\n", mongoConnStr)
	mongoClient, err = mongo.Connect(options.Client().ApplyURI(mongoConnStr))
	if err != nil {
		panic(err)
	}

	code := m.Run()
	if err := postgresClient.Close(); err != nil {
		panic(err)
	}
	if err := mongoClient.Disconnect(ctx); err != nil {
		panic(err)
	}
	os.Exit(code)
}

func pg(ctx context.Context, connString string) (*sql.DB, error) {
	config, err := pgxpool.ParseConfig(connString)
	if err != nil {
		return nil, fmt.Errorf("failed to parse connection string: %w", err)
	}
	db, err := sql.Open("pgx", stdlib.RegisterConnConfig(config.ConnConfig))
	if err != nil {
		return nil, fmt.Errorf("failed to create database pool: %w", err)
	}
	if err := db.PingContext(ctx); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}
	return db, nil
}
