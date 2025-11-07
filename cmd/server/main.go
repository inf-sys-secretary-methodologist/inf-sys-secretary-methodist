package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/auth/application/usecases"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/auth/domain/repositories"
	persistence "github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/auth/infrastructure"
	authHandler "github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/auth/interfaces/http/handlers"
	authMiddleware "github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/auth/interfaces/http/middleware"
	appMiddleware "github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/shared/application/middleware"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/shared/infrastructure/cache"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/shared/infrastructure/config"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/shared/infrastructure/logging"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/shared/infrastructure/middleware"
)

func main() {
	// Handle version flag
	if len(os.Args) > 1 && os.Args[1] == "--version" {
		fmt.Println("inf-sys-secretary-methodist v0.1.0")
		return
	}

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Initialize logger
	logger := logging.NewLogger(cfg.Log.Level)
	securityLogger := logging.NewSecurityLogger(logger)
	auditLogger := logging.NewAuditLogger(logger)
	perfLogger := logging.NewPerformanceLogger(logger)

	logger.Info("Starting application", map[string]interface{}{
		"environment": cfg.Environment,
		"version":     cfg.Version,
	})

	// Initialize database
	db, err := initDatabase(cfg, logger)
	if err != nil {
		logger.Error("Failed to initialize database", map[string]interface{}{
			"error": err.Error(),
		})
		os.Exit(1)
	}
	defer db.Close()

	logger.Info("Database connected successfully", map[string]interface{}{
		"max_open_conns": cfg.Database.MaxOpenConns,
		"max_idle_conns": cfg.Database.MaxIdleConns,
	})

	// Initialize Redis cache
	redisCache, err := initRedisCache(cfg, logger)
	if err != nil {
		logger.Warn("Redis cache not available, running without cache", map[string]interface{}{
			"error": err.Error(),
		})
	}
	if redisCache != nil {
		defer redisCache.Close()
		logger.Info("Redis cache connected successfully", nil)
	}

	// Initialize auth module with all optimizations
	authUseCase, authHandlerInstance := initAuthModule(
		db,
		redisCache,
		securityLogger,
		auditLogger,
		perfLogger,
		cfg,
	)

	// Initialize shared middleware
	corsMiddleware := appMiddleware.NewCORSMiddleware(
		cfg.CORS.AllowedOrigins,
		cfg.CORS.AllowedMethods,
		cfg.CORS.AllowedHeaders,
	)
	loggingMiddleware := appMiddleware.NewLoggingMiddleware(logger)

	// Setup router with all middleware
	router := setupRoutes(
		authUseCase,
		authHandlerInstance,
		securityLogger,
		perfLogger,
		cfg,
		logger,
		corsMiddleware,
		loggingMiddleware,
		db,
		redisCache,
	)

	// Create HTTP server
	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Server.Port),
		Handler:      router,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
		IdleTimeout:  cfg.Server.IdleTimeout,
	}

	// Start server in goroutine
	go func() {
		logger.Info("Server starting", map[string]interface{}{
			"port":        cfg.Server.Port,
			"environment": cfg.Environment,
		})
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("Server failed to start", map[string]interface{}{
				"error": err.Error(),
			})
			os.Exit(1)
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Server shutting down...", nil)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		logger.Error("Server forced to shutdown", map[string]interface{}{
			"error": err.Error(),
		})
	}

	logger.Info("Server stopped", nil)
}

func initDatabase(cfg *config.Config, logger *logging.Logger) (*sql.DB, error) {
	dsn := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		cfg.Database.Host,
		cfg.Database.Port,
		cfg.Database.Username,
		cfg.Database.Password,
		cfg.Database.Database,
	)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Configure connection pool for optimal performance
	db.SetMaxOpenConns(cfg.Database.MaxOpenConns)
	db.SetMaxIdleConns(cfg.Database.MaxIdleConns)
	db.SetConnMaxLifetime(cfg.Database.ConnMaxLifetime)

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return db, nil
}

func initRedisCache(cfg *config.Config, logger *logging.Logger) (*cache.RedisCache, error) {
	redisAddr := fmt.Sprintf("%s:%d", cfg.Redis.Host, cfg.Redis.Port)
	redisCache, err := cache.NewRedisCache(redisAddr, cfg.Redis.Password, cfg.Redis.DB)
	if err != nil {
		return nil, err
	}
	return redisCache, nil
}

func initAuthModule(
	db *sql.DB,
	redisCache *cache.RedisCache,
	securityLog *logging.SecurityLogger,
	auditLog *logging.AuditLogger,
	perfLog *logging.PerformanceLogger,
	cfg *config.Config,
) (*usecases.AuthUseCase, *authHandler.AuthHandler) {
	// JWT secrets from config
	jwtSecret := []byte(cfg.JWT.AccessSecret)
	refreshSecret := []byte(cfg.JWT.RefreshSecret)

	// Initialize base user repository
	baseUserRepo := persistence.NewUserRepositoryPG(db)

	// Initialize session repository (will be used in future for refresh token management)
	_ = persistence.NewSessionRepositoryPG(db)

	// Wrap with caching if Redis is available
	var userRepo interface{} = baseUserRepo
	if redisCache != nil {
		userCache := cache.NewUserCache(redisCache, 5*time.Minute)
		userRepo = persistence.NewCachedUserRepository(baseUserRepo, userCache, perfLog)
	}

	// Initialize use case with full logging and session repository
	authUseCase := usecases.NewAuthUseCase(
		userRepo.(repositories.UserRepository),
		jwtSecret,
		refreshSecret,
		securityLog,
		auditLog,
	)

	// Initialize handler
	authHandlerInstance := authHandler.NewAuthHandler(authUseCase)

	return authUseCase, authHandlerInstance
}

func setupRoutes(
	authUseCase *usecases.AuthUseCase,
	authHandlerInstance *authHandler.AuthHandler,
	securityLog *logging.SecurityLogger,
	perfLog *logging.PerformanceLogger,
	cfg *config.Config,
	logger *logging.Logger,
	corsMiddleware *appMiddleware.CORSMiddleware,
	loggingMiddleware *appMiddleware.LoggingMiddleware,
	db *sql.DB,
	redisCache *cache.RedisCache,
) *gin.Engine {
	router := gin.New()

	// Global middleware stack (order matters!)
	router.Use(gin.Recovery())
	router.Use(middleware.RequestIDMiddleware()) // Request ID для трейсинга
	router.Use(middleware.RequestContextMiddleware())
	router.Use(corsMiddleware.Handler()) // CORS из конфига
	router.Use(authMiddleware.SecurityHeadersMiddleware())
	router.Use(loggingMiddleware.Handler()) // Логирование всех запросов
	router.Use(performanceMiddleware(perfLog))

	// Health check endpoint with dependency checks
	router.GET("/health", healthCheckHandler(db, redisCache))

	var rateLimiter *middleware.RateLimiter
	if redisCache != nil {
		rateLimiter = middleware.NewRateLimiter(
			redisCache.Client(), // <- прямой доступ к *redis.Client
			5,                   // 5 запросов
			15*time.Minute,      // за 15 минут
		)
	}

	// Public auth routes with rate limiting
	authGroup := router.Group("/api/auth")
	if rateLimiter != nil {
		authGroup.Use(rateLimiter.RateLimitMiddleware())
		authGroup.Use(rateLimitLogger(securityLog))
	}
	{
		authGroup.POST("/register", authHandlerInstance.Register)
		authGroup.POST("/login", authHandlerInstance.Login)
		authGroup.POST("/refresh", authHandlerInstance.RefreshToken)
	}

	// Protected routes (require JWT)
	protectedGroup := router.Group("/api")
	protectedGroup.Use(authMiddleware.JWTMiddleware(authUseCase))
	{
		protectedGroup.GET("/me", func(c *gin.Context) {
			userID, _ := c.Get("user_id")
			role, _ := c.Get("role")
			c.JSON(http.StatusOK, gin.H{
				"user_id": userID,
				"role":    role,
			})
		})

		// Admin only routes
		adminGroup := protectedGroup.Group("/admin")
		adminGroup.Use(authMiddleware.RequireRole("admin"))
		{
			adminGroup.GET("/users", func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{"message": "Admin users list"})
			})
		}
	}

	return router
}

// performanceMiddleware logs HTTP request performance
func performanceMiddleware(perfLog *logging.PerformanceLogger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		c.Next()

		duration := time.Since(start)
		perfLog.LogHTTPRequest(
			c.Request.Context(),
			c.Request.Method,
			c.Request.URL.Path,
			c.Writer.Status(),
			duration,
		)
	}
}

// rateLimitLogger logs rate limit violations
func rateLimitLogger(securityLog *logging.SecurityLogger) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		// Check if rate limited
		if c.Writer.Status() == http.StatusTooManyRequests {
			securityLog.LogRateLimitExceeded(c.Request.Context(), c.Request.URL.Path)
		}
	}
}

// healthCheckHandler returns a health check endpoint with dependency checks
func healthCheckHandler(db *sql.DB, redisCache *cache.RedisCache) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Second)
		defer cancel()

		status := "OK"
		checks := make(map[string]interface{})

		// Check database
		if err := db.PingContext(ctx); err != nil {
			status = "DEGRADED"
			checks["database"] = map[string]interface{}{
				"status": "DOWN",
				"error":  err.Error(),
			}
		} else {
			checks["database"] = map[string]interface{}{
				"status": "UP",
			}
		}

		// Check Redis if available
		if redisCache != nil {
			if err := redisCache.Ping(ctx); err != nil {
				status = "DEGRADED"
				checks["redis"] = map[string]interface{}{
					"status": "DOWN",
					"error":  err.Error(),
				}
			} else {
				checks["redis"] = map[string]interface{}{
					"status": "UP",
				}
			}
		} else {
			checks["redis"] = map[string]interface{}{
				"status": "DISABLED",
			}
		}

		httpStatus := http.StatusOK
		if status == "DEGRADED" {
			httpStatus = http.StatusServiceUnavailable
		}

		c.JSON(httpStatus, gin.H{
			"status":    status,
			"timestamp": time.Now().UTC(),
			"checks":    checks,
		})
	}
}
