# Testes Go - Terra Allwert API

Este módulo contém todos os testes da API Terra Allwert implementados em Go, seguindo as melhores práticas de testing e organizados conforme a estrutura definida no projeto.

## 📁 Estrutura dos Testes

```
test/
├── go.mod                    # Módulo Go de testes (separado da aplicação)
├── fixtures/                 # Fixtures e dados de teste
│   └── testutils/
│       └── fixtures.go       # Utilitários para criação de dados de teste
├── unit/                     # Testes unitários
│   ├── auth_service_test.go       # Testes do serviço de autenticação
│   ├── file_upload_service_test.go # Testes do serviço de upload
│   └── tower_repository_test.go   # Testes do repositório de torres
├── integration/              # Testes de integração
│   ├── graphql_api_test.go        # Testes da API GraphQL
│   └── minio_upload_test.go       # Testes de integração MinIO
└── e2e/                     # Testes end-to-end
    ├── auth_test.go              # Testes E2E de autenticação
    └── full_workflow_test.go     # Workflow completo E2E
```

## 🧪 Tipos de Testes

### 1. **Testes Unitários** (`unit/`)

Testam componentes isolados usando mocks e stubs.

#### `auth_service_test.go`
- ✅ **AuthService.Login** - Casos de sucesso e falha
- ✅ **JWT Token Generation/Validation** - Criação e validação de tokens
- ✅ **Password Hashing** - Verificação de senhas
- ✅ **Mock repositories** - UserRepository, PasswordHasher, JWTService
- ✅ **Benchmarks** - Performance dos serviços de autenticação

#### `file_upload_service_test.go`
- ✅ **FileUploadService.GetSignedUploadURL** - Geração de URLs assinadas
- ✅ **FileUploadService.ConfirmFileUpload** - Confirmação de uploads
- ✅ **File validation** - Validação de tipos e extensões
- ✅ **Mock MinIO client** - Simulação de operações de storage
- ✅ **Error scenarios** - Tratamento de erros
- ✅ **Benchmarks** - Performance de operações de arquivo

### 2. **Testes de Integração** (`integration/`)

Testam a integração entre componentes, incluindo APIs e serviços externos.

#### `graphql_api_test.go`
- ✅ **GraphQL Test Suite** - Suite completa de testes GraphQL
- ✅ **Authentication flow** - Login e obtenção de tokens
- ✅ **File upload mutations** - getSignedUploadUrl, confirmFileUpload
- ✅ **Presentation mutations** - createMenu, createImageCarousel, etc.
- ✅ **Business data queries** - getRouteBusinessData, getCacheConfiguration
- ✅ **Error handling** - Queries inválidas e não autorizadas
- ✅ **Mock GraphQL server** - Simulação completa da API

#### `minio_upload_test.go`
- ✅ **MinIO Test Suite** - Suite de testes para MinIO
- ✅ **Presigned URL operations** - Upload e download com URLs assinadas
- ✅ **Multiple file uploads** - Upload de diferentes tipos de arquivo
- ✅ **Large file handling** - Teste com arquivos grandes (1MB+)
- ✅ **Concurrent operations** - Uploads simultâneos
- ✅ **Error scenarios** - URLs inválidas, arquivos inexistentes
- ✅ **Performance tests** - Benchmarks de velocidade de upload

### 3. **Testes End-to-End** (`e2e/`)

Testam fluxos completos da aplicação simulando usuários reais.

#### `full_workflow_test.go`
- ✅ **Complete Workflow Test** - Fluxo completo do usuário
- ✅ **Health check** - Verificação de saúde da API
- ✅ **Authentication workflow** - Login completo
- ✅ **File upload workflow** - Upload completo (URL → Upload → Confirmação)
- ✅ **Presentation creation** - Criação de todos os tipos de presentation
- ✅ **Business data operations** - Operações de dados de negócio
- ✅ **Sync operations** - Sincronização offline
- ✅ **Data consistency** - Verificação de integridade dos dados
- ✅ **Performance tests** - Tempo de execução do workflow
- ✅ **Concurrent operations** - Operações simultâneas

## 🚀 Como Executar os Testes

### Pré-requisitos

1. **Go 1.21+** instalado
2. **Serviços de infraestrutura** rodando:
   ```bash
   docker compose -f docker/docker-compose.infra.yml up -d
   ```
3. **API rodando** (para testes de integração):
   ```bash
   cd src && go run main.go
   ```

### Comandos de Teste

```bash
# Navegar para o módulo de testes
cd test

# Executar todos os testes
go test ./...

# Executar testes com verbose
go test -v ./...

# Executar apenas testes unitários
go test ./unit/...

# Executar apenas testes de integração
go test ./integration/...

# Executar apenas testes E2E
go test ./e2e/...

# Executar com coverage
go test -cover ./...

# Executar benchmarks
go test -bench=. ./...

# Executar testes específicos
go test -run TestAuthService_Login_Success ./unit/
go test -run TestGraphQLTestSuite ./integration/
go test -run TestCompleteWorkflow ./e2e/

# Executar com timeout (útil para E2E)
go test -timeout=5m ./e2e/...
```

### Executar Testes em Paralelo

```bash
# Executar testes em paralelo (mais rápido)
go test -parallel 4 ./...

# Para testes que precisam de mais recursos
go test -parallel 2 ./integration/...
```

## 📋 Fixtures e Dados de Teste

### `testutils/fixtures.go`

Fornece funções utilitárias para criar dados de teste consistentes:

```go
// Users
testUser := testutils.CreateTestUser(entities.RoleAdmin)
adminUser := testutils.CreateAdminUser()
viewerUser := testutils.CreateViewerUser()

// Entities
tower := testutils.CreateTestTower()
floor := testutils.CreateTestFloor("tower-id")
apartment := testutils.CreateTestApartment("floor-id")

// Auth
claims := testutils.CreateTestJWTClaims(entities.RoleAdmin)
loginReq := testutils.CreateTestLoginRequest()
loginResp := testutils.CreateTestLoginResponse()

// Gallery
galleryImage := testutils.CreateTestGalleryImage()
imagePin := testutils.CreateTestImagePin("gallery-id")
```

## 🔧 Configuração dos Testes

### Variáveis de Ambiente de Teste

```bash
# API
API_URL=http://localhost:3000
GRAPHQL_ENDPOINT=http://localhost:3000/graphql

# MinIO para testes
MINIO_ENDPOINT=localhost:9000
MINIO_ACCESS_KEY=minio
MINIO_SECRET_KEY=minio123
MINIO_BUCKET=terraallwert

# Database para testes
DB_HOST=localhost
DB_PORT=5432
DB_USER=apiuser
DB_PASSWORD=apipass
DB_NAME=terraallwert_test

# Redis para testes
REDIS_HOST=localhost
REDIS_PORT=6379
```

### Configuração de Mock

Os testes usam mocks extensivos para isolar componentes:

- **MockUserRepository** - Simulação do repositório de usuários
- **MockPasswordHasher** - Simulação do hash de senhas  
- **MockJWTService** - Simulação do serviço JWT
- **MockMinIOClient** - Simulação do cliente MinIO
- **MockFileRepository** - Simulação do repositório de arquivos

## 📊 Cobertura de Testes

### Objetivos de Cobertura

- **Unit tests**: 90%+ de cobertura
- **Integration tests**: 80%+ dos endpoints
- **E2E tests**: 70%+ dos workflows principais

### Relatório de Cobertura

```bash
# Gerar relatório de cobertura
go test -coverprofile=coverage.out ./...

# Ver cobertura no browser
go tool cover -html=coverage.out

# Ver cobertura por função
go tool cover -func=coverage.out
```

## 🎯 Funcionalidades Testadas

### ✅ Autenticação e Autorização
- [x] Login com email/senha
- [x] Geração de tokens JWT
- [x] Validação de tokens
- [x] Refresh tokens
- [x] Roles e permissões

### ✅ Upload e Gerenciamento de Arquivos
- [x] Geração de URLs assinadas
- [x] Upload para MinIO
- [x] Confirmação de uploads
- [x] Validação de tipos de arquivo
- [x] Múltiplos uploads simultâneos
- [x] Handling de arquivos grandes

### ✅ API GraphQL
- [x] Todas as mutations definidas no API_INTEGRATION.md
- [x] Todas as queries definidas no API_INTEGRATION.md
- [x] Error handling e validação
- [x] Autenticação em endpoints protegidos
- [x] Introspection do schema

### ✅ Presentations
- [x] Criação de menus
- [x] Image carousels
- [x] Floor plans  
- [x] Pin maps
- [x] CRUD completo

### ✅ Business Data
- [x] Obtenção de dados por rota
- [x] Atualização de dados de negócio
- [x] Configurações de cache
- [x] Resolução de conflitos

### ✅ Sincronização Offline
- [x] Request full sync (ZIP)
- [x] Status de sincronização
- [x] Metadata de sync
- [x] URLs de download

## 🐛 Debugging de Testes

### Logs Detalhados

```bash
# Executar com logs verbose
go test -v -args -test.v=true ./...

# Debug de testes específicos
go test -v -run TestSpecificTest ./unit/ -args -test.v=true
```

### Identificar Testes Lentos

```bash
# Executar com timeout e identificar gargalos
go test -v -timeout=30s ./... | grep -E "(PASS|FAIL|panic|timeout)"
```

### Memory Profiling

```bash
# Executar com profiling de memória
go test -memprofile=mem.prof -run TestSpecificTest ./...

# Analisar profile
go tool pprof mem.prof
```

## ⚡ Performance e Benchmarks

### Benchmarks Disponíveis

```bash
# Benchmarks de autenticação
go test -bench=BenchmarkJWTService ./unit/

# Benchmarks de upload
go test -bench=BenchmarkFileUpload ./unit/
go test -bench=BenchmarkMinIOUpload ./integration/

# Benchmarks de GraphQL
go test -bench=BenchmarkGraphQLQuery ./integration/

# Benchmarks E2E
go test -bench=BenchmarkCompleteWorkflow ./e2e/
```

### Resultados de Performance Esperados

- **JWT Generation**: < 1ms
- **File Upload (1KB)**: < 10ms  
- **File Upload (1MB)**: < 100ms
- **GraphQL Query**: < 5ms
- **Complete Workflow**: < 5s

## 🔄 CI/CD Integration

### GitHub Actions

```yaml
name: Go Tests
on: [push, pull_request]

jobs:
  test:
    runs-on: ubuntu-latest
    services:
      postgres:
        image: postgres:16
        env:
          POSTGRES_PASSWORD: apipass
          POSTGRES_USER: apiuser
          POSTGRES_DB: terraallwert_test
      redis:
        image: redis:7-alpine
      minio:
        image: quay.io/minio/minio:latest
        
    steps:
      - uses: actions/checkout@v3
      - uses: actions/setup-go@v3
        with:
          go-version: '1.21'
          
      - name: Run Unit Tests
        working-directory: ./test
        run: go test ./unit/...
        
      - name: Run Integration Tests  
        working-directory: ./test
        run: go test ./integration/...
        
      - name: Run E2E Tests
        working-directory: ./test  
        run: go test -timeout=10m ./e2e/...
        
      - name: Upload Coverage
        run: |
          go test -coverprofile=coverage.out ./...
          curl -s https://codecov.io/bash | bash
```

### Makefile Commands

```makefile
.PHONY: test test-unit test-integration test-e2e test-coverage

test:
	cd test && go test ./...

test-unit:
	cd test && go test ./unit/...

test-integration:
	cd test && go test ./integration/...

test-e2e:
	cd test && go test -timeout=10m ./e2e/...

test-coverage:
	cd test && go test -coverprofile=coverage.out ./...
	cd test && go tool cover -html=coverage.out -o coverage.html

benchmark:
	cd test && go test -bench=. ./...
```

## 🎛️ Configurações Avançadas

### Test Tags

```bash
# Executar apenas testes rápidos
go test -tags=fast ./...

# Executar apenas testes de integração
go test -tags=integration ./...

# Pular testes que precisam de serviços externos
go test -tags=unit ./...
```

### Parallel Execution

```go
// Em testes que podem rodar em paralelo
func TestParallelSafe(t *testing.T) {
    t.Parallel()
    // teste aqui
}

// Em benchmarks
func BenchmarkParallel(b *testing.B) {
    b.SetParallelism(4)
    b.RunParallel(func(pb *testing.PB) {
        // benchmark aqui
    })
}
```

## 📈 Métricas e Monitoramento

### Coleta de Métricas

```bash
# Executar testes com métricas
go test -json ./... > test-results.json

# Analisar tempo de execução
go test -json ./... | jq '.Action == "pass" | .Elapsed'
```

### Relatórios de Teste

```bash
# Gerar relatório em formato JUnit
go test -json ./... | go-junit-report > junit.xml

# Relatório de cobertura em formato Cobertura
gocov convert coverage.out | gocov-xml > coverage.xml
```

---

## 🎯 Próximos Passos

1. **Adicionar mais cenários de erro** nos testes existentes
2. **Implementar testes de stress** para alta concorrência  
3. **Adicionar testes de segurança** para validação de inputs
4. **Criar testes de regressão** para bugs conhecidos
5. **Implementar mutation testing** para verificar qualidade dos testes
6. **Adicionar testes de contract** para APIs externas

---

**💡 Dica**: Execute `make test` para executar todos os testes ou use os scripts bash em `/scripts/` para testes específicos da infraestrutura.