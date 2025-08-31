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

	"api/api/handler"
	"api/api/middleware"
	"api/api/router"
	"api/data/repositories"
	"api/data/services"
	"api/infra/cache"
	"api/infra/client"
	"api/infra/config"
	"api/infra/database"

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
	config    *config.Config
	fiber     *fiber.App
	db        *gorm.DB
	redis     *redis.Client
	scheduler *services.SchedulerService
	info      *AppInfo
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

	if err := database.RunMigrations(a.db); err != nil {
		return fmt.Errorf("falha ao executar migrações: %w", err)
	}

	log.Println("🌱 Executando seeds iniciais...")
	if err := database.RunSeeds(a.db); err != nil {
		return fmt.Errorf("falha ao executar seeds: %w", err)
	}

	log.Println("✅ Migrações e seeds executados com sucesso")
	return nil
}

// loadInitialData verifica e carrega dados iniciais
func (a *App) loadInitialData() error {
	log.Println("🔧 Verificando dados iniciais...")

	// Criar repositórios necessários para verificação
	userRepo := repositories.NewUserRepository(a.db)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Verificar se há dados no banco
	hasData, err := userRepo.HasUsers(ctx)
	if err != nil {
		log.Printf("⚠️ Aviso: Falha ao verificar dados existentes: %v", err)
		return nil
	}

	// Se não há dados, carregar dados iniciais
	if !hasData {
		log.Println("🔄 Primeira inicialização detectada. Carregando dados iniciais...")

		// Criar serviços necessários
		userService := services.NewUserService(userRepo)

		// Carregar dados iniciais
		if err := userService.LoadInitialData(ctx); err != nil {
			log.Printf("⚠️ Aviso: Falha ao carregar dados iniciais: %v", err)
		} else {
			log.Println("✅ Dados iniciais carregados com sucesso")
		}
	} else {
		log.Println("✅ Dados já existem no banco")
	}

	return nil
}

// initFiberApp configura o servidor HTTP Fiber
func (a *App) initFiberApp() error {
	log.Println("🌐 Configurando aplicação web...")

	// Configurar Fiber com otimizações para produção
	app := fiber.New(fiber.Config{
		AppName:               a.config.App.Name,
		ServerHeader:          "Terra Allwert",
		DisableStartupMessage: false,
		ErrorHandler:          middleware.ErrorHandler(a.config),
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

	// Middleware de ID de requisição
	app.Use(middleware.RequestID())

	// Middleware de logging
	app.Use(middleware.Logger(a.config))

	// Middleware de CORS
	app.Use(middleware.CORS(a.config))

	// Middleware de segurança
	app.Use(middleware.Security(a.config))

	// Middleware de rate limiting
	app.Use(middleware.RateLimit(a.config))

	// Middleware de monitoramento
	app.Use(middleware.Monitoring(a.config))

	// Configurar rotas
	a.setupRoutes(app)

	a.fiber = app
	log.Println("✅ Aplicação web configurada")
	return nil
}

// setupRoutes configura todas as rotas da aplicação
func (a *App) setupRoutes(app *fiber.App) {
	log.Println("🛣️ Configurando rotas da aplicação...")

	// Inicializar repositórios
	userRepo := repositories.NewUserRepository(a.db)
	productRepo := repositories.NewProductRepository(a.db)
	orderRepo := repositories.NewOrderRepository(a.db)

	// Inicializar clientes externos
	externalClient := client.NewExternalClient(a.config)

	// Inicializar serviços
	cacheService := services.NewCacheService(a.redis)
	userService := services.NewUserService(userRepo)
	productService := services.NewProductService(productRepo)
	orderService := services.NewOrderService(orderRepo, productRepo)

	// Inicializar handlers com todas as dependências
	userHandler := handler.NewUserHandler(
		userService,
		cacheService,
		a.config,
	)

	productHandler := handler.NewProductHandler(
		productService,
		cacheService,
		a.config,
	)

	orderHandler := handler.NewOrderHandler(
		orderService,
		userService,
		productService,
		cacheService,
		a.config,
	)

	healthHandler := handler.NewHealthHandler(
		a.db,
		a.redis,
		a.info,
		a.config,
	)

	adminHandler := handler.NewAdminHandler(
		userRepo,
		productRepo,
		orderRepo,
		userService,
		productService,
		orderService,
		cacheService,
		externalClient,
		a.config,
		a.db,
	)

	// Configurar rotas
	router.SetupRoutes(app, userHandler, productHandler, orderHandler, healthHandler, adminHandler, a.config)

	log.Println("✅ Rotas configuradas com sucesso")
}

// startBackgroundTasks inicia o scheduler para tarefas em background
func (a *App) startBackgroundTasks() error {
	log.Println("⏰ Iniciando tarefas em background...")

	// Inicializar dependências do scheduler
	userRepo := repositories.NewUserRepository(a.db)
	productRepo := repositories.NewProductRepository(a.db)
	orderRepo := repositories.NewOrderRepository(a.db)
	cacheService := services.NewCacheService(a.redis)

	// Criar serviços necessários
	userService := services.NewUserService(userRepo)
	productService := services.NewProductService(productRepo)
	orderService := services.NewOrderService(orderRepo, productRepo)

	// Criar e iniciar scheduler
	scheduler := services.NewScheduler(
		userService,
		productService,
		orderService,
		cacheService,
		a.config,
	)
	scheduler.Start()

	a.scheduler = scheduler
	log.Println("✅ Scheduler iniciado com sucesso")
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

		// Parar scheduler
		if a.scheduler != nil {
			log.Println("⏹️ Parando scheduler...")
			a.scheduler.Stop()
		}

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