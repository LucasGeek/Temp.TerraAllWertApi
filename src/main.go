package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	"api/data/repositories"
	"api/domain/entities"
	"api/graph"
	"api/graph/generated"
	"api/infra/auth"
	"api/infra/cache"
	"api/infra/config"
	"api/infra/database"
	"api/domain/errors"
	"api/infra/middleware"
	"api/infra/storage"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/adaptor"
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

// Implementações placeholder para serviços faltantes
type storageServiceImpl struct {
	config *config.Config
}

func (s *storageServiceImpl) GenerateSignedUploadURL(ctx context.Context, fileName, contentType, folder string) (*entities.SignedUploadURL, error) {
	return &entities.SignedUploadURL{
		UploadURL: "http://localhost:9000/placeholder-upload",
		AccessURL: "http://localhost:9000/placeholder-access",
		ExpiresIn: 3600,
	}, nil
}

func (s *storageServiceImpl) GenerateSignedDownloadURL(ctx context.Context, objectPath string, expiry time.Duration) (string, error) {
	return "http://localhost:9000/placeholder-download", nil
}

func (s *storageServiceImpl) UploadFile(ctx context.Context, objectPath string, reader io.Reader, size int64, contentType string) error {
	return nil
}

func (s *storageServiceImpl) DeleteFile(ctx context.Context, objectPath string) error {
	return nil
}

func (s *storageServiceImpl) FileExists(ctx context.Context, objectPath string) (bool, error) {
	return true, nil
}

func (s *storageServiceImpl) GetFileMetadata(ctx context.Context, objectPath string) (*entities.FileMetadata, error) {
	return &entities.FileMetadata{
		FileName:    "placeholder.jpg",
		FileSize:    1024,
		ContentType: "image/jpeg",
		UploadedAt:  time.Now(),
	}, nil
}

func (s *storageServiceImpl) CreateBulkDownload(ctx context.Context, towerID string) (*entities.BulkDownload, error) {
	return &entities.BulkDownload{
		DownloadURL: "http://localhost:9000/bulk-download.zip",
		FileName:    "tower-" + towerID + ".zip",
		FileSize:    1024000,
		ExpiresIn:   3600,
		CreatedAt:   time.Now(),
	}, nil
}

func (s *storageServiceImpl) GetBulkDownloadStatus(ctx context.Context, downloadID string) (*entities.BulkDownloadStatus, error) {
	downloadURL := "http://localhost:9000/bulk-download.zip"
	return &entities.BulkDownloadStatus{
		ID:             downloadID,
		Status:         entities.BulkDownloadStateCompleted,
		Progress:       100,
		TotalFiles:     10,
		ProcessedFiles: 10,
		DownloadURL:    &downloadURL,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}, nil
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

	// Criar middleware GraphQL para autenticação
	graphqlAuthMiddleware := middleware.NewGraphQLAuthMiddleware(authService)

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

	// GraphQL Playground (desenvolvimento)
	if a.config.IsDev() {
		app.Get("/playground", adaptor.HTTPHandler(playground.Handler("GraphQL playground", "/graphql")))
		log.Println("🎮 GraphQL Playground disponível em /playground")
	}

	// Criar todos os repositórios necessários para GraphQL
	towerRepo := repositories.NewTowerRepository(a.db)
	floorRepo := repositories.NewFloorRepository(a.db)
	apartmentRepo := repositories.NewApartmentRepository(a.db)
	galleryRepo := repositories.NewGalleryRepository(a.db)
	imagePinRepo := repositories.NewImagePinRepository(a.db)
	apartmentImageRepo := repositories.NewApartmentImageRepository(a.db)
	appConfigRepo := repositories.NewAppConfigRepository(a.db)

	// Criar serviços de storage (placeholder básico)
	storageService := &storageServiceImpl{config: a.config}
	bulkDownloadService := storage.NewBulkDownloadService(
		nil, // MinIO client placeholder
		towerRepo,
		apartmentRepo,
		galleryRepo,
		"terra-allwert",
		"terra-allwert-temp",
	)

	// GraphQL endpoint com middleware de autenticação e tratamento de erros
	graphqlHandler := handler.NewDefaultServer(generated.NewExecutableSchema(generated.Config{
		Resolvers: &graph.Resolver{
			TowerRepo:           towerRepo,
			FloorRepo:           floorRepo,
			ApartmentRepo:       apartmentRepo,
			GalleryRepo:         galleryRepo,
			ImagePinRepo:        imagePinRepo,
			ApartmentImageRepo:  apartmentImageRepo,
			AppConfigRepo:       appConfigRepo,
			UserRepo:            userRepo,
			AuthService:         authService,
			StorageService:      storageService,
			BulkDownloadService: bulkDownloadService,
		},
	}))

	// Configurar tratamento de erros customizado
	graphqlHandler.SetErrorPresenter(errors.ErrorPresenter)
	graphqlHandler.SetRecoverFunc(errors.RecoverFunc)

	app.Post("/graphql", graphqlAuthMiddleware.HTTPAuthMiddleware(), func(c *fiber.Ctx) error {
		// Converter Fiber context para http.ResponseWriter e http.Request
		adaptor.HTTPHandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Adicionar usuário do Fiber ao contexto GraphQL
			r = r.WithContext(middleware.WithUser(r.Context(), c))
			graphqlHandler.ServeHTTP(w, r)
		})(c)
		return nil
	})

	log.Println("✅ Aplicação configurada apenas com GraphQL")
	log.Println("📊 GraphQL endpoint disponível em /graphql")
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
	log.Printf("📊 GraphQL: http://localhost%s/graphql", port)
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
