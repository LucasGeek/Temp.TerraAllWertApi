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

```
api/
├── cmd/              # Comandos da aplicação
├── internal/         # Código privado da aplicação
│   ├── config/       # Configurações
│   ├── database/     # Conexão e migrações
│   ├── graph/        # GraphQL resolvers
│   ├── handlers/     # HTTP handlers
│   ├── middleware/   # Middlewares
│   ├── models/       # Modelos de dados
│   ├── repositories/ # Camada de acesso a dados
│   ├── services/     # Lógica de negócio
│   └── storage/      # Integração com MinIO
├── migrations/       # Migrações do banco
├── pkg/              # Código público/reutilizável
├── scripts/          # Scripts auxiliares
└── docker/           # Configurações Docker
```

## Funcionalidades Principais

### Autenticação e Autorização
- Login/Logout com JWT
- Refresh tokens
- Roles: Admin, Consultant, Client
- Middleware de autenticação

### Gestão de Torres
- CRUD completo de torres
- Upload de imagens e documentos
- Informações de localização
- Estatísticas de ocupação

### Gestão de Pavimentos
- Pavimentos por torre
- Plantas baixas
- Configuração de apartamentos

### Gestão de Apartamentos
- Tipos: Studio, 1-3 quartos, Cobertura
- Status: Disponível, Reservado, Vendido
- Características e amenidades
- Galeria de imagens
- Vídeos promocionais
- Tour virtual 360°

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
go run main.go

# Build
go build -o bin/api main.go

# Testes
go test ./...

# Migrações
go run cmd/migrate/main.go up

# Gerar código GraphQL
go run github.com/99designs/gqlgen generate
```

## Variáveis de Ambiente

```env
# Server
PORT=8080
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

## Melhorias Planejadas

1. **Segurança**: Implementar rate limiting e proteção DDoS
2. **Performance**: Otimização de queries N+1
3. **Testes**: Aumentar cobertura para 80%+
4. **Documentação**: Swagger/OpenAPI completo
5. **Monitoramento**: Integração com Prometheus/Grafana
6. **CI/CD**: Pipeline automatizado com GitHub Actions

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

## RED FLAGS - Pare e Reconsidere
- Criar interfaces quando uma implementação concreta resolve
- Usar design patterns complexos sem necessidade clara
- Otimizar performance antes de medir gargalos reais
- Adicionar dependências para funcionalidades simples
- Implementar configurações complexas "para o futuro"


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