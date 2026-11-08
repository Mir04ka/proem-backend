package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"proem-backend/internal/config"
	"proem-backend/internal/handler"
	"proem-backend/internal/middleware"
	"proem-backend/internal/repository"
	"proem-backend/internal/service"
	"proem-backend/migrations"
	"proem-backend/pkg/migrator"
	"proem-backend/pkg/token"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	cfg := config.Load()

	if cfg.AppEnv == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	log.Println("Applying database migrations...")
	if err := migrator.RunMigrations(migrations.FS, cfg.DatabaseUrl); err != nil {
		log.Fatalf("Database migration failed: %v", err)
	}

	poolConfig, err := pgxpool.ParseConfig(cfg.DatabaseUrl)
	if err != nil {
		log.Fatalf("Unable to parse DB config: %v", err)
	}
	poolConfig.MaxConns = 25
	poolConfig.MinConns = 5
	poolConfig.MaxConnIdleTime = 15 * time.Minute

	dbPool, err := pgxpool.NewWithConfig(context.Background(), poolConfig)
	if err != nil {
		log.Fatalf("Unable to connect to database: %v", err)
	}
	defer dbPool.Close()

	if err := dbPool.Ping(context.Background()); err != nil {
		log.Fatalf("Database ping failed: %v", err)
	}
	log.Println("Connected to PostgreSQL successfully")

	jwtManager := token.NewJWTManager(cfg.JWTSecret, cfg.TokenTTL)
	mailerService := service.NewMailerService(cfg)
	storageService, err := service.NewLocalStorageService(cfg)
	if err != nil {
		log.Fatalf("Storage service initialization failed: %v", err)
	}

	userRepo := repository.NewUserRepository(dbPool)
	articleRepo := repository.NewArticleRepository(dbPool)

	authService := service.NewAuthService(userRepo, jwtManager, mailerService)
	articleService := service.NewArticleService(articleRepo)

	authHandler := handler.NewAuthHandler(authService)
	articleHandler := handler.NewArticleHandler(articleService)
	uploadHandler := handler.NewUploadHandler(storageService)

	router := gin.Default()
	router.Use(gin.Recovery())
	router.Use(gin.Logger())
	router.Use(middleware.CORSMiddleware())

	// Раздача загруженных изображений
	router.Static("/uploads", cfg.UploadDir)

	authMiddleware := middleware.AuthMiddleware(jwtManager)

	api := router.Group("/api/v1")
	{
		// 1. Авторизация
		auth := api.Group("/auth")
		{
			auth.POST("/register", authHandler.Register)
			auth.GET("/verify", authHandler.VerifyEmail)
			auth.POST("/login", authHandler.Login)
			auth.POST("/change-password", authMiddleware, authHandler.ChangePassword)
		}

		// 2. Пользователи
		users := api.Group("/users")
		{
			users.GET("/:id", authHandler.GetProfile)
			users.GET("/:id/articles", articleHandler.ListByUser)
		}

		// 3. Статьи (регистрируем маршруты напрямую без вложенного Group(""))
		articles := api.Group("/articles")
		{
			articles.GET("", articleHandler.List)
			articles.GET("/:id", articleHandler.GetByID)
			articles.GET("/:id/comments", articleHandler.GetComments)

			// Закрытые маршруты статей
			articles.POST("", authMiddleware, articleHandler.Publish)
			articles.PUT("/:id", authMiddleware, articleHandler.Update)
			articles.POST("/:id/vote", authMiddleware, articleHandler.Vote)
			articles.POST("/:id/comments", authMiddleware, articleHandler.AddComment)
		}

		// 4. Загрузка картинок
		media := api.Group("/media")
		{
			media.POST("/upload", authMiddleware, uploadHandler.UploadImage)
		}
	}

	// Выводим все зарегистрированные маршруты в лог при старте
	log.Println("--- REGISTERED ROUTES ---")
	for _, r := range router.Routes() {
		log.Printf("[ROUTE] %-6s %s\n", r.Method, r.Path)
	}
	log.Println("-------------------------")

	srv := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: router,
	}

	go func() {
		log.Printf("Server listening on port %s", cfg.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Listen error: %s\n", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown:", err)
	}

	log.Println("Server exiting")
}
