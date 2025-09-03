# Terra Allwert API

## Objetivo do Projeto

Sistema de gerenciamento de torres residenciais e comerciais com foco em vendas e apresentação de apartamentos. A API fornece endpoints GraphQL e REST para gerenciar torres, pavimentos, apartamentos e seus recursos multimídia.

## Tecnologias Principais

- **Go 1.21+**: Linguagem principal
- **Fiber v2**: Framework web de alta performance
- **GraphQL (gqlgen)**: API GraphQL type-safe
- **GORM**: ORM para Go
- **PostgreSQL/MySQL**: Banco de dados relacional
- **MinIO**: Armazenamento de objetos S3-compatible
- **Redis**: Cache e sessões
- **JWT**: Autenticação
- **Zap**: Logger estruturado

## Estrutura do Projeto

**OBRIGATÓRIO: Separação src/test com go.work**

O projeto deve sempre ser separado em dois módulos Go usando go.work:
- `./src/`: Código da aplicação principal
- `./test/`: Todos os testes (unit, integration, e2e)

```
api/
├── go.work           # Workspace Go (src + test)
├── src/              # Módulo principal da aplicação
│   ├── go.mod        # Dependências da aplicação
│   ├── main.go       # Entry point
│   ├── api/handlers/ # HTTP handlers
│   ├── domain/       # Entidades e interfaces
│   ├── data/         # Repositórios e implementações
│   ├── infra/        # Infraestrutura (auth, middleware, storage)
│   └── docker/       # Configurações Docker
└── test/             # Módulo de testes
    ├── go.mod        # Dependências de teste
    ├── unit/         # Testes unitários
    ├── integration/  # Testes de integração
    ├── e2e/          # Testes end-to-end
    └── fixtures/     # Fixtures e mocks
```

## Funcionalidades Principais

### Autenticação e Autorização
- Login/Logout com JWT
- Refresh tokens
- Roles: Admin, Consultant, Client
- Middleware de autenticação

### Sistema de Cache
- Cache Redis para queries frequentes
- Invalidação inteligente
- TTL configurável

### Storage
- Upload de imagens com MinIO
- Redimensionamento automático
- URLs presigned para acesso seguro
- Backup automático

## Requisitos Funcionais

1. **Modo Offline**: Sincronização de dados para funcionamento sem internet
2. **Multi-tenant**: Suporte para múltiplas empresas/corretoras
3. **Relatórios**: Exportação de dados em PDF/Excel
4. **Notificações**: Sistema de notificações em tempo real
5. **Auditoria**: Log completo de ações dos usuários

## Comandos Úteis

```bash
# Desenvolvimento
go run src/main.go

# Build
go build -o bin/api src/main.go

# Testes
go test ./...

# Seeding manual (opcional - executado automaticamente no primeiro start)
go run src/cmd/seed/main.go

# Gerar código GraphQL (se usando GraphQL)
go run github.com/99designs/gqlgen generate
```

## Inicialização Automática

A API detecta automaticamente se o banco de dados está vazio no primeiro startup e executa:

1. **Auto-migrations**: Criação automática das tabelas
2. **Auto-seeding**: População inicial com dados essenciais:
   - Empresa padrão "Terra Allwert"
   - Usuários básicos (admin, manager, visitor)
   - Dados de demonstração

**Usuários criados automaticamente:**
- `admin@terra.com` / `senha123` (Super Admin)
- `admin@allwert` / `senha123` (Admin da Empresa)
- `manager@allwert` / `senha123` (Manager)
- `visitor@allwert` / `senha123` (Visitor)
- `demo@terraallwert.com` / `senha123` (Demo User)

### Controle do Seeding

O seeding automático só é executado quando:
- Não existem enterprises no banco
- Não existem users no banco

Para forçar re-seeding:
1. Limpe o banco de dados
2. Reinicie a aplicação

## Deploy em Produção

### Configuração do Ambiente de Produção

1. **Copie o arquivo de exemplo**:
```bash
cp .env.prd.example .env.prd
```

2. **Configure as variáveis de produção**:
```bash
# Edite o arquivo .env.prd com dados reais
nano .env.prd
```

3. **Execute o deploy**:
```bash
./scripts/deploy-production.sh
```

### Estrutura do Ambiente de Produção

O ambiente de produção usa Docker Swarm com os seguintes serviços:

- **API**: `terra-allwert-prd_prd-api` (2 réplicas)
- **PostgreSQL**: `terra-allwert-prd_prd-db` 
- **Redis**: `terra-allwert-prd_prd-cache`
- **MinIO**: `terra-allwert-prd_prd-minio`

### Comandos de Produção

```bash
# Ver status dos serviços
docker service ls

# Ver logs da API
docker service logs -f terra-allwert-prd_prd-api

# Escalar API para 3 réplicas
docker service scale terra-allwert-prd_prd-api=3

# Remover stack completo
docker stack rm terra-allwert-prd

# Backup do banco
docker exec $(docker ps -q -f name=terra-allwert-prd_prd-db) \
  pg_dump -U $DB_USER $DB_NAME > backup_$(date +%Y%m%d_%H%M%S).sql
```

### Segurança em Produção

- ✅ Senhas fortes e únicas para todos os serviços
- ✅ JWT Secret com 64+ caracteres
- ✅ Logs estruturados em JSON
- ⚠️  Configure SSL/TLS com certificados válidos
- ⚠️  Configure firewall e regras de rede
- ⚠️  Implemente monitoramento e alertas
- ⚠️  Configure backups automáticos

## Variáveis de Ambiente

```env
# Server
PORT=3000
ENVIRONMENT=development

# Database
DB_DRIVER=postgres
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=
DB_NAME=terra_allwert
DB_SSLMODE=disable

# Storage (MinIO)
MINIO_ENDPOINT=localhost:9000
MINIO_ACCESS_KEY=
MINIO_SECRET_KEY=
MINIO_USE_SSL=false
MINIO_BUCKET=terra-allwert

# Redis
REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_PASSWORD=
REDIS_DB=0

# JWT
JWT_SECRET=your-secret-key
JWT_EXPIRATION_HOURS=24
```

## Problemas Conhecidos do Projeto Legado

- Credenciais hardcoded no código
- Falta de validação de entrada
- Ausência de testes automatizados
- Código monolítico em app.py (2000+ linhas)
- CORS muito permissivo (*)
- Decoradores de autenticação vazios

## Contato

Para dúvidas ou sugestões sobre este projeto, consulte a documentação completa em `/docs` ou entre em contato com a equipe de desenvolvimento.

# Anti-Over-engineering - Checklist

## Antes de Implementar Qualquer Feature
- [ ] Esta funcionalidade está na lista de essenciais do MVP?
- [ ] Esta é a solução mais simples que resolve o problema?
- [ ] Esta abstração é realmente necessária agora?
- [ ] Este padrão de design agrega valor imediato?

## Processo de Commit

1. **Questionar se quer realizar o commit**
2. **Adicionar alterações**: `git add .`
3. **Commit padronizado**: `git commit -m "tipo: descrição resumida"`

## Conventional Commits (PT-BR)

### Tipos principais:
- `feat`: nova funcionalidade
- `fix`: correção de bug
- `docs`: alterações na documentação
- `style`: formatação, sem mudança de lógica
- `refactor`: refatoração sem nova funcionalidade ou fix
- `test`: adição ou correção de testes
- `chore`: tarefas de manutenção

### Formato:
```
tipo: descrição resumida em português
```

### Exemplos:
```bash
git commit -m "feat: adiciona entrada na fila via QR code"
git commit -m "fix: corrige posição na fila em tempo real"
git commit -m "docs: atualiza README com instruções de setup"
git commit -m "refactor: simplifica lógica de notificações"
git commit -m "style: aplica formatação Elixir padrão"
git commit -m "test: adiciona testes para contexto de filas"
git commit -m "chore: atualiza dependências do Phoenix"
```

## Diretrizes para Mensagens

- **Resumido**: máximo 50 caracteres
- **Imperativo**: "adiciona" ao invés de "adicionado"
- **Português brasileiro**: linguagem clara e direta
- **Foco no valor**: o que foi implementado/corrigido
- **Sem pontuação final**: não usar ponto no final

## Commits Compostos (quando necessário)

Para mudanças maiores, usar corpo do commit:
```bash
git commit -m "feat: implementa sistema de notificações

- Adiciona worker para processar mensagens WhatsApp
- Integra RabbitMQ para fila de mensagens
- Implementa templates de notificação"
```

## Evitar

- Mensagens vagas: "atualiza código", "corrige bugs"
- Misturar tipos diferentes numa mesma alteração
- Commits muito grandes (quebrar em commits menores)
- Usar português misturado com inglês
- Ao realizar não adicionar:

`
🤖 Generated with [Claude Code](https://claude.ai/code)
Co-Authored-By: default avatarClaude <noreply@anthropic.com>
`