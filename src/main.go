package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	"api/api/handlers"
	"api/data/repositories"
	"api/infra/auth"
	"api/infra/cache"
	"api/infra/config"
	"api/infra/database"
	"api/infra/middleware"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

// AppInfo contém informações básicas da aplicação
type AppInfo struct {
	Nome        string `json:"nome"`
	Versao      string `json:"versao"`
	Ambiente    string `json:"ambiente"`
	TempoInicio string `json:"tempo_inicio"`
	VersaoGo    string `json:"versao_go"`
}

// App representa a aplicação principal com todas suas dependências
type App struct {
	config *config.Config
	fiber  *fiber.App
	db     *gorm.DB
	redis  *redis.Client
	info   *AppInfo
}

func main() {
	log.Println("🚀 Iniciando Terra Allwert API...")

	// Carregar configurações do ambiente
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("❌ Falha ao carregar configurações: %v", err)
	}

	// Criar instância da aplicação
	app := &App{
		config: cfg,
		info: &AppInfo{
			Nome:        cfg.App.Name,
			Versao:      cfg.App.Version,
			Ambiente:    cfg.App.Environment,
			TempoInicio: time.Now().Format(time.RFC3339),
			VersaoGo:    runtime.Version(),
		},
	}

	// Inicializar todos os componentes
	if err := app.Initialize(); err != nil {
		log.Fatalf("❌ Falha ao inicializar aplicação: %v", err)
	}

	// Configurar parada elegante da aplicação
	app.setupGracefulShutdown()

	// Iniciar servidor HTTP
	app.Start()
}

// Initialize configura e inicializa todos os componentes da aplicação
func (a *App) Initialize() error {
	log.Println("📋 Inicializando componentes da aplicação...")

	// Conectar ao banco de dados PostgreSQL
	if err := a.initDatabase(); err != nil {
		return fmt.Errorf("falha ao conectar ao banco de dados: %w", err)
	}

	// Conectar ao Redis
	if err := a.initRedis(); err != nil {
		return fmt.Errorf("falha ao conectar ao Redis: %w", err)
	}

	// Executar migrações e seeds
	if err := a.runMigrations(); err != nil {
		return fmt.Errorf("falha ao executar migrações: %w", err)
	}

	// Verificar e carregar dados iniciais
	if err := a.loadInitialData(); err != nil {
		return fmt.Errorf("falha ao carregar dados iniciais: %w", err)
	}

	// Configurar aplicação Fiber
	if err := a.initFiberApp(); err != nil {
		return fmt.Errorf("falha ao configurar aplicação web: %w", err)
	}

	// Iniciar tarefas em background
	if err := a.startBackgroundTasks(); err != nil {
		return fmt.Errorf("falha ao iniciar tarefas em background: %w", err)
	}

	log.Println("✅ Todos os componentes inicializados com sucesso")
	return nil
}

// initDatabase estabelece conexão com PostgreSQL
func (a *App) initDatabase() error {
	log.Println("🔌 Conectando ao banco de dados PostgreSQL...")

	db, err := database.ConnectPostgres(*a.config)
	if err != nil {
		return err
	}

	a.db = db
	log.Println("✅ Conexão com banco de dados estabelecida")
	return nil
}

// initRedis estabelece conexão com Redis
func (a *App) initRedis() error {
	log.Println("🔌 Conectando ao Redis...")

	redisClient, err := cache.ConnectRedis(*a.config)
	if err != nil {
		return err
	}

	a.redis = redisClient
	log.Println("✅ Conexão com Redis estabelecida")
	return nil
}

// runMigrations executa migrações do banco e seeds iniciais
func (a *App) runMigrations() error {
	log.Println("🗄️ Executando migrações do banco de dados...")

	if err := database.AutoMigrate(a.db); err != nil {
		return fmt.Errorf("falha ao executar migrações: %w", err)
	}

	log.Println("✅ Migrações executadas com sucesso")
	return nil
}

// loadInitialData verifica e carrega dados iniciais
func (a *App) loadInitialData() error {
	log.Println("🔧 Verificando dados iniciais...")
	
	// Criar repositórios
	userRepo := repositories.NewUserRepository(a.db)
	
	// Criar serviços de autenticação
	authService := auth.NewJWTService(userRepo, a.config)
	
	// Criar usuários iniciais
	if err := database.CreateInitialUser(a.db, authService); err != nil {
		log.Printf("⚠️ Aviso: Falha ao criar usuários iniciais: %v", err)
	}
	
	log.Println("✅ Banco de dados pronto para uso")
	return nil
}

// initFiberApp configura o servidor HTTP Fiber
func (a *App) initFiberApp() error {
	log.Println("🌐 Configurando aplicação web...")

	// Configurar Fiber com otimizações básicas
	app := fiber.New(fiber.Config{
		AppName:               a.config.App.Name,
		ServerHeader:          "Terra Allwert",
		DisableStartupMessage: false,
		ReadTimeout:           30 * time.Second,
		WriteTimeout:          30 * time.Second,
		IdleTimeout:           60 * time.Second,
		BodyLimit:             4 * 1024 * 1024, // 4MB
		EnablePrintRoutes:     a.config.App.Debug,
		Prefork:               false, // Não usar prefork em containers
	})

	// Middleware de recuperação de pânico
	app.Use(recover.New(recover.Config{
		EnableStackTrace: a.config.App.Debug,
	}))

	// Configurar rotas
	a.setupRoutes(app)

	a.fiber = app
	log.Println("✅ Aplicação web configurada")
	return nil
}

// setupRoutes configura todas as rotas da aplicação
func (a *App) setupRoutes(app *fiber.App) {
	log.Println("🛣️ Configurando rotas da aplicação...")

	// Criar repositórios
	userRepo := repositories.NewUserRepository(a.db)
	
	// Criar serviços
	authService := auth.NewJWTService(userRepo, a.config)
	
	// Criar handlers
	authHandler := handlers.NewAuthHandler(authService)
	
	// Criar middleware
	authMiddleware := middleware.NewAuthMiddleware(authService)

	// Health check endpoint (público)
	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status":  "ok",
			"service": a.info.Nome,
			"version": a.info.Versao,
			"env":     a.info.Ambiente,
			"uptime":  time.Since(time.Now()).String(),
		})
	})

	// API info endpoint (público)
	app.Get("/", func(c *fiber.Ctx) error {
		return c.JSON(a.info)
	})

	// Grupo de rotas de autenticação (públicas)
	auth := app.Group("/api/auth")
	auth.Post("/login", authHandler.Login)
	auth.Post("/refresh", authHandler.RefreshToken)

	// Grupo de rotas protegidas
	api := app.Group("/api")
	api.Use(authMiddleware.RequireAuth())
	
	// Profile endpoint (autenticado)
	api.Get("/profile", authHandler.GetProfile)
	api.Post("/logout", authHandler.Logout)

	log.Println("✅ Rotas configuradas com autenticação")
}

// startBackgroundTasks inicia o scheduler para tarefas em background
func (a *App) startBackgroundTasks() error {
	log.Println("⏰ Tarefas em background prontas")
	return nil
}

// Start inicia o servidor HTTP
func (a *App) Start() {
	port := ":" + a.config.App.Port

	log.Printf("🌟 Servidor iniciando na porta %s", a.config.App.Port)
	log.Printf("📊 Ambiente: %s", a.config.App.Environment)
	log.Printf("🔗 URL: http://localhost%s", port)
	log.Printf("📚 Documentação: http://localhost%s/api/v1/docs", port)
	log.Printf("💚 Health Check: http://localhost%s/health", port)

	if a.config.IsPrd() {
		log.Println("🔒 Modo de produção ativado")
	} else if a.config.IsDev() {
		log.Println("🔧 Modo de desenvolvimento ativado")
	}

	if err := a.fiber.Listen(port); err != nil {
		log.Fatalf("❌ Falha ao iniciar servidor: %v", err)
	}
}

// setupGracefulShutdown configura parada elegante da aplicação
func (a *App) setupGracefulShutdown() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-c
		log.Println("⏳ Iniciando parada elegante da aplicação...")

		// Definir timeout para shutdown
		ctx, cancel := context.WithTimeout(context.Background(), a.config.App.GracefulTimeout)
		defer cancel()

		// Tarefas em background (futuro)

		// Parar servidor HTTP
		if a.fiber != nil {
			log.Println("⏹️ Parando servidor HTTP...")
			if err := a.fiber.ShutdownWithContext(ctx); err != nil {
				log.Printf("❌ Erro ao parar servidor: %v", err)
			}
		}

		// Fechar conexão com banco de dados
		if a.db != nil {
			log.Println("⏹️ Fechando conexão com banco de dados...")
			if err := database.CloseConnection(a.db); err != nil {
				log.Printf("❌ Erro ao fechar banco: %v", err)
			}
		}

		// Fechar conexão com Redis
		if a.redis != nil {
			log.Println("⏹️ Fechando conexão com Redis...")
			if err := cache.CloseConnection(a.redis); err != nil {
				log.Printf("❌ Erro ao fechar Redis: %v", err)
			}
		}

		log.Println("✅ Aplicação parada com sucesso")
		os.Exit(0)
	}()
}