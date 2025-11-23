package main

import (
	"context"
	"github.com/Unitazavr/AvitoPR/internal/http"
	"github.com/Unitazavr/AvitoPR/internal/repository"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	"log"
	"os"
)

func main() {
	//Конфиги

	err := godotenv.Load()
	if err != nil {
		log.Println("No .env file found, using system environment variables")
	}
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
	router.Use(cors.New(cors.Config{
		AllowAllOrigins:  true,
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}))
	//Роутинг, создание сервисов и контроллеров
	http.RegisterRoutes(router, userRepo, teamRepo, prRepo)

	addr := ":" + port
	log.Printf("starting server on %s", addr)

	//Обработка ошибки
	if err := router.Run(addr); err != nil {
		log.Fatalf("server stopped with error: %v", err)
	}

}
