/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package main

import (
	_ "embed"
	"fmt"
	"log"
	"myapp/internal"
	"myapp/internal/middleware"
	"myapp/internal/router"
	"net/http"
	"os"

	"github.com/joho/godotenv"
)

//go:embed docs/openapi.yaml
var openAPISpec []byte

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables.")
	}

	internal.InitDB(dbConnString())

	serverPort := os.Getenv("PORT")
	if serverPort == "" {
		serverPort = os.Getenv("APP_PORT")
	}
	if serverPort == "" {
		serverPort = "3000"
	}

	handler := middleware.CORS(router.Setup(openAPISpec))

	log.Printf("🚀 Inventory Management System API started on :%s", serverPort)
	log.Printf("📖 Swagger UI: http://localhost:%s/swagger/", serverPort)
	log.Printf("🔐 Auth: POST /auth/login (default admin: admin@ims.local / admin123)")
	if os.Getenv("AUTH_DISABLED") == "true" {
		log.Println("⚠️  AUTH_DISABLED=true — all routes open without JWT")
	}
	if err := http.ListenAndServe(":"+serverPort, handler); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}

func dbConnString() string {
	if url := os.Getenv("DATABASE_URL"); url != "" {
		return url
	}

	sslmode := os.Getenv("DB_SSLMODE")
	if sslmode == "" {
		sslmode = "disable"
	}

	return fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_NAME"),
		sslmode,
	)
}
