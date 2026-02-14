package postgres

import (
	"context"
	"log"
	"github.com/jackc/pgx/v5/pgxpool"
)

func NewPool() *pgxpool.Pool {
	dsn := "postgres://postgres@127.0.0.1:5433/pulseq_db?sslmode=disable"// DSN = Data Source Name like DB address

	//create pool
	pool, err := pgxpool.New(context.Background(), dsn)

	if err != nil {
		log.Fatal("Unable to connect to database:", err)
	}

	err = pool.Ping(context.Background())

	if err != nil {
		log.Fatal("Database not responding:", err)
	}

	log.Println("Connected to Postgres")

	return pool

}


