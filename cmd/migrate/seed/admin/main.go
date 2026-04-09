package main

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/joho/godotenv/autoload"
	"github.com/myselfBZ/satjade-backend/internal/postgres"
	"golang.org/x/crypto/bcrypt"
)

var AdminEmail string
var AdminPassword string

func init() {
	AdminEmail = os.Getenv("ADMIN_EMAIL")
	if AdminEmail == "" {
		panic("env variable ADMIN_EMAIL not set")
	}

	AdminPassword = os.Getenv("ADMIN_PASSWORD")

	if AdminPassword == "" {
		panic("env variable ADMIN_PASSWORD not set")
	}
}

func seedAdmin(pool *pgxpool.Pool, email, password string) error {
	q := `INSERT INTO 
			users(full_name, email, role, password_hash) 
		VALUES ('The Super User', $1, 'admin', $2)`
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute*5)
	defer cancel()
	_, err = pool.Exec(ctx, q, email, hash)
	if err != nil {
		return err
	}
	log.Println("Admin has been created successfully")
	return nil
}

func main() {
	pool, err := postgres.New(postgres.Config{
		Addr:        os.Getenv("DB"),
		MaxConns:    15,
		MinConns:    15,
		MaxIdleTime: "15m",
	})
	if err != nil {
		panic(err)
	}
	if err := seedAdmin(pool, AdminEmail, AdminPassword); err != nil {
		panic(err)
	}
}

