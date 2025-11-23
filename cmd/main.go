package main

import (
	"context"
	"github.com/Unitazavr/AvitoPR/internal/http"
	"github.com/Unitazavr/AvitoPR/internal/repository"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"log"
	"os"
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
	prRepo := repository.NewPrRepo(pool)

	//Джин
	if os.Getenv("GIN_MODE") == "release" {
		gin.SetMode(gin.ReleaseMode)
	}
	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(gin.Logger())

	//Роутинг, создание сервисов и контроллеров
	http.RegisterRoutes(router, userRepo, teamRepo, prRepo)

	addr := ":" + port
	log.Printf("starting server on %s", addr)

	//Обработка ошибки
	if err := router.Run(addr); err != nil {
		log.Fatalf("server stopped with error: %v", err)
	}

}
