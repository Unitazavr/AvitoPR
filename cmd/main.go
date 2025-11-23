package main

import (
	"context"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"log"
	"os"

	"yourmodule/internal/repository"
	"yourmodule/internal/transport"
)

func main() {
	//Конфиги
	dsn := os.Getenv("POSTGRES_DSN")
	if dsn == "" {
		log.Fatal("POSTGRES_DSN is required (e.g. postgres://user:pass@db:5432/dbname?sslmode=disable)")
	}
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	//Подключение к БД
	pool, err := pgxpool.New(context.Background(), dsn)
	if err != nil {
		log.Fatalf("failed to create pgx pool: %v", err)
	}
	defer pool.Close()

	userRepo := repository.NewUserRepo(pool)
	teamRepo := repository.NewTeamRepo(pool)
	prRepo := repository.NewPRRepo(pool)

	//Джин
	if os.Getenv("GIN_MODE") == "release" {
		gin.SetMode(gin.ReleaseMode)
	}
	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(gin.Logger())

	//Роутинг
	transport.RegisterRoutes(router, userRepo, teamRepo)

	addr := ":" + port
	log.Printf("starting server on %s", addr)

	//Обработка ошибки
	if err := router.Run(addr); err != nil {
		log.Fatalf("server stopped with error: %v", err)
	}

}
