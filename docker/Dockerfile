# =============================================================================
# DOCKERFILE MULTI-STAGE PARA TERRA ALLWERT API
# =============================================================================
# Arquivo: docker/Dockerfile
# Propósito: Build multi-stage para diferentes ambientes (dev, stg, prd)
# 
# TARGETS DISPONÍVEIS:
# - dev: Ambiente de desenvolvimento com ferramentas de debug
# - stg: Ambiente de staging com configurações intermediárias
# - prd: Ambiente de produção otimizado e seguro
# - default: Aponta para produção (prd)
# 
# USO:
# docker build --target dev -t terra_allwert/api:dev .
# docker build --target stg -t terra_allwert/api:stg .
# docker build --target prd -t terra_allwert/api:latest .
# =============================================================================

# =============================================================================
# BASE STAGE - DEPENDÊNCIAS E CONFIGURAÇÕES COMUNS
# =============================================================================
FROM golang:1.24-alpine AS base

# -----------------------------------------------------------------------------
# METADADOS E LABELS
# -----------------------------------------------------------------------------
LABEL maintainer="Terra Allwert Team <dev@terra_allwert.com>"
LABEL description="Terra Allwert API - Sistema de cotações de moedas"
LABEL version="1.0.0"

# -----------------------------------------------------------------------------
# INSTALAÇÃO DE DEPENDÊNCIAS DO SISTEMA
# -----------------------------------------------------------------------------
# Instalar pacotes essenciais para build e runtime
RUN apk add --no-cache \
    git \
    ca-certificates \
    tzdata \
    curl \
    make \
    bash

# -----------------------------------------------------------------------------
# CONFIGURAÇÃO DO AMBIENTE GO
# -----------------------------------------------------------------------------
# Definir diretório de trabalho dentro do container
WORKDIR /app

# Configurar variáveis de ambiente Go para otimização
ENV CGO_ENABLED=0
ENV GOOS=linux
ENV GOARCH=amd64

# -----------------------------------------------------------------------------
# GERENCIAMENTO DE DEPENDÊNCIAS GO
# -----------------------------------------------------------------------------
# Copiar arquivos de dependências primeiro (camada de cache otimizada)
# ✅ CORRIGIDO: Os arquivos estão em src/, não na raiz
COPY src/go.mod src/go.sum ./

# Download e verificação de dependências (apenas se go.mod existir)
# Esta camada será cachada se go.mod/go.sum não mudarem
RUN if [ -f go.mod ]; then \
        echo "📦 Baixando dependências Go..." && \
        go mod download && \
        go mod verify; \
    else \
        echo "⚠️  go.mod não encontrado, pulando download de dependências"; \
    fi

# -----------------------------------------------------------------------------
# CÓPIA DO CÓDIGO FONTE
# -----------------------------------------------------------------------------
# Copiar todo o código fonte (depois das dependências para cache otimizado)
# ✅ CORRIGIDO: Copiar apenas o diretório src para o container
COPY src/ ./

# Verificar estrutura do projeto e arquivos copiados
RUN echo "📁 Estrutura do projeto após cópia:" && \
    ls -la && \
    echo "📁 Verificando arquivos Go:" && \
    find . -name "*.go" -type f | head -10 && \
    echo "📁 Verificando conteúdo completo:" && \
    find . -type f | head -20 && \
    if [ -f "main.go" ]; then \
        echo "✅ Arquivo main.go encontrado"; \
    else \
        echo "⚠️  Arquivo main.go não encontrado"; \
        echo "📁 Listando todos os arquivos:"; \
        find . -type f; \
    fi

# =============================================================================
# DEVELOPMENT TARGET - AMBIENTE DE DESENVOLVIMENTO
# =============================================================================
FROM base AS dev

# -----------------------------------------------------------------------------
# LABELS ESPECÍFICOS DE DESENVOLVIMENTO
# -----------------------------------------------------------------------------
LABEL environment="development"
LABEL debug="enabled"
LABEL hot_reload="enabled"

# -----------------------------------------------------------------------------
# FERRAMENTAS DE DESENVOLVIMENTO
# -----------------------------------------------------------------------------
# Instalar Delve (debugger Go)
RUN echo "🔧 Instalando ferramentas de desenvolvimento..." && \
    go install github.com/go-delve/delve/cmd/dlv@latest

# Instalar Air para hot reload (opcional)
RUN go install github.com/air-verse/air@latest

# Instalar ferramentas de lint e formatação
RUN go install golang.org/x/tools/cmd/goimports@latest && \
    go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# -----------------------------------------------------------------------------
# CONFIGURAÇÃO DE USUÁRIO DE DESENVOLVIMENTO
# -----------------------------------------------------------------------------
# Criar grupo de desenvolvimento com ID específico
RUN addgroup -g 1001 -S devgroup

# Criar usuário de desenvolvimento não-root
RUN adduser -u 1001 -S devuser -G devgroup -s /bin/bash

# -----------------------------------------------------------------------------
# CONFIGURAÇÃO DE PERMISSÕES PARA DESENVOLVIMENTO
# -----------------------------------------------------------------------------
# Configurar permissões do diretório de trabalho
RUN chown -R devuser:devgroup /app && \
    chmod -R 755 /app

# Criar diretório para cache do Go
RUN mkdir -p /home/devuser/.cache/go-build && \
    chown -R devuser:devgroup /home/devuser/.cache

# -----------------------------------------------------------------------------
# CONFIGURAÇÃO DE USUÁRIO
# -----------------------------------------------------------------------------
# Mudar para usuário não-root (segurança)
USER devuser

# -----------------------------------------------------------------------------
# CONFIGURAÇÃO DE PORTAS PARA DESENVOLVIMENTO
# -----------------------------------------------------------------------------
# Expor porta da aplicação
EXPOSE 8080

# Expor porta do debugger Delve
EXPOSE 2345

# Expor porta do pprof (profiling)
EXPOSE 6060

# -----------------------------------------------------------------------------
# PONTO DE ENTRADA PARA DESENVOLVIMENTO
# -----------------------------------------------------------------------------
# Comando padrão para desenvolvimento (pode ser sobrescrito)
CMD ["sh", "-c", "echo '🚀 Modo desenvolvimento iniciado' && go run ."]

# =============================================================================
# BUILD STAGE - COMPILAÇÃO OTIMIZADA PARA PRODUÇÃO
# =============================================================================
FROM base AS builder

# -----------------------------------------------------------------------------
# LABELS ESPECÍFICOS DE BUILD
# -----------------------------------------------------------------------------
LABEL stage="builder"
LABEL purpose="compile_optimized_binary"

# -----------------------------------------------------------------------------
# CONFIGURAÇÃO DE BUILD OTIMIZADO
# -----------------------------------------------------------------------------
# Definir variáveis de build
ARG BUILD_TIME
ARG GIT_COMMIT
ARG VERSION=1.0.0

# Adiciona utilitário `file` temporariamente para diagnóstico do binário
RUN apk add --no-cache file

# Exibir informações de build
RUN echo "🔨 Iniciando build otimizado..." && \
    echo "📅 Build time: ${BUILD_TIME}" && \
    echo "🔗 Git commit: ${GIT_COMMIT}" && \
    echo "🏷️  Version: ${VERSION}"

# -----------------------------------------------------------------------------
# COMPILAÇÃO OTIMIZADA
# -----------------------------------------------------------------------------
# Compilar aplicação diretamente (código já está em /app)
RUN echo "⚙️  Compilando aplicação Terra Allwert..." && \
    CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-w -s -extldflags '-static' \
              -X main.BuildTime=${BUILD_TIME} \
              -X main.GitCommit=${GIT_COMMIT} \
              -X main.Version=${VERSION}" \
    -a -installsuffix cgo \
    -o /app/terra_allwert-api \
    . && \
    echo "✅ Compilação concluída com sucesso"

# Verificar se o binário foi criado corretamente
RUN ls -la /app/terra_allwert-api && \
    file /app/terra_allwert-api && \
    echo "📦 Tamanho do binário: $(du -h /app/terra_allwert-api | cut -f1)"

# =============================================================================
# STAGING TARGET - AMBIENTE DE HOMOLOGAÇÃO
# =============================================================================
FROM alpine:3.18 AS stg

# -----------------------------------------------------------------------------
# LABELS ESPECÍFICOS DE STAGING
# -----------------------------------------------------------------------------
LABEL environment="staging"
LABEL version="1.0.0"
LABEL purpose="testing_and_validation"

# -----------------------------------------------------------------------------
# INSTALAÇÃO DE DEPENDÊNCIAS PARA STAGING
# -----------------------------------------------------------------------------
# Instalar dependências necessárias para staging (inclui ferramentas de debug)
RUN apk --no-cache add \
    ca-certificates \
    tzdata \
    curl \
    bash \
    jq \
    netcat-openbsd \
    htop \
    && rm -rf /var/cache/apk/

# -----------------------------------------------------------------------------
# CONFIGURAÇÃO DE USUÁRIO NÃO-ROOT
# -----------------------------------------------------------------------------
# Criar grupo da aplicação
RUN addgroup -g 1001 -S appgroup

# Criar usuário da aplicação não-root
RUN adduser -u 1001 -S appuser -G appgroup -s /bin/bash

# -----------------------------------------------------------------------------
# CONFIGURAÇÃO DO AMBIENTE DE STAGING
# -----------------------------------------------------------------------------
# Configurar diretório de trabalho
WORKDIR /app

# Copiar binário compilado do stage builder
COPY --from=builder /app/terra_allwert-api .

# Verificar se o binário foi copiado corretamente
RUN ls -la /app/terra_allwert-api && \
    echo "✅ Binário copiado para staging"

# -----------------------------------------------------------------------------
# CONFIGURAÇÃO DE DIRETÓRIOS E PERMISSÕES
# -----------------------------------------------------------------------------
# Criar diretórios necessários para aplicação
RUN mkdir -p /app/logs /app/config /app/tmp && \
    chown -R appuser:appgroup /app && \
    chmod 755 /app && \
    chmod +x /app/terra_allwert-api

# -----------------------------------------------------------------------------
# CONFIGURAÇÃO DE TIMEZONE
# -----------------------------------------------------------------------------
# Configurar timezone para Brasil
RUN ln -sf /usr/share/zoneinfo/America/Campo_Grande /etc/localtime && \
    echo "America/Campo_Grande" > /etc/timezone

# -----------------------------------------------------------------------------
# CONFIGURAÇÃO DE USUÁRIO
# -----------------------------------------------------------------------------
# Mudar para usuário não-root
USER appuser

# -----------------------------------------------------------------------------
# CONFIGURAÇÃO DE PORTA
# -----------------------------------------------------------------------------
# Expor porta da aplicação
EXPOSE 8080

# -----------------------------------------------------------------------------
# HEALTH CHECK PARA STAGING
# -----------------------------------------------------------------------------
# Health check otimizado para staging (mais tolerante)
HEALTHCHECK --interval=30s --timeout=10s --start-period=30s --retries=3 \
    CMD curl -f http://localhost:8080/health || exit 1

# -----------------------------------------------------------------------------
# COMANDO PADRÃO PARA STAGING
# -----------------------------------------------------------------------------
CMD ["./terra_allwert-api"]

# =============================================================================
# PRODUCTION TARGET - AMBIENTE DE PRODUÇÃO
# =============================================================================
FROM alpine:3.18 AS prd

# -----------------------------------------------------------------------------
# LABELS ESPECÍFICOS DE PRODUÇÃO
# -----------------------------------------------------------------------------
LABEL environment="production"
LABEL version="1.0.0"
LABEL maintainer="Terra Allwert Team <prod@terra_allwert.com>"
LABEL security="hardened"
LABEL optimization="maximum"

# -----------------------------------------------------------------------------
# INSTALAÇÃO MÍNIMA PARA PRODUÇÃO
# -----------------------------------------------------------------------------
# Instalar apenas dependências essenciais (mínimas para segurança)
RUN apk --no-cache add \
    ca-certificates \
    tzdata
    # ✅ NOTA: Removidas ferramentas de debug para produção

# Limpar cache completamente
RUN rm -rf /var/cache/apk/* /tmp/* /var/tmp/*

# -----------------------------------------------------------------------------
# CONFIGURAÇÃO DE USUÁRIO NÃO-ROOT (PRODUÇÃO)
# -----------------------------------------------------------------------------
# Criar grupo da aplicação com ID específico
RUN addgroup -g 1001 -S appgroup

# Criar usuário da aplicação não-root (sem shell para segurança)
RUN adduser -u 1001 -S appuser -G appgroup -D -s /sbin/nologin

# -----------------------------------------------------------------------------
# CONFIGURAÇÃO DO AMBIENTE DE PRODUÇÃO
# -----------------------------------------------------------------------------
# Configurar diretório de trabalho
WORKDIR /app

# Copiar binário compilado do stage builder
COPY --from=builder /app/terra_allwert-api .

# -----------------------------------------------------------------------------
# CONFIGURAÇÃO DE SEGURANÇA MÁXIMA
# -----------------------------------------------------------------------------
# Criar apenas diretório de logs (mínimo necessário)
RUN mkdir -p /app/logs && \
    chown -R appuser:appgroup /app && \
    chmod 755 /app && \
    chmod 555 /app/terra_allwert-api  # ✅ Permissão restritiva (apenas execução)

# Remover permissões desnecessárias
RUN chmod -R o-rwx /app

# -----------------------------------------------------------------------------
# CONFIGURAÇÃO DE TIMEZONE PARA PRODUÇÃO
# -----------------------------------------------------------------------------
# Configurar timezone para Brasil
RUN ln -sf /usr/share/zoneinfo/America/Campo_Grande /etc/localtime && \
    echo "America/Campo_Grande" > /etc/timezone

# -----------------------------------------------------------------------------
# CONFIGURAÇÃO DE USUÁRIO
# -----------------------------------------------------------------------------
# Mudar para usuário não-root (segurança crítica)
USER appuser

# -----------------------------------------------------------------------------
# CONFIGURAÇÃO DE PORTA (PRODUÇÃO)
# -----------------------------------------------------------------------------
# Expor apenas porta necessária
EXPOSE 8080

# -----------------------------------------------------------------------------
# HEALTH CHECK OTIMIZADO PARA PRODUÇÃO
# -----------------------------------------------------------------------------
# Health check com wget (mais leve que curl) e configurações otimizadas
HEALTHCHECK --interval=60s --timeout=5s --start-period=60s --retries=2 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

# -----------------------------------------------------------------------------
# COMANDO FINAL PARA PRODUÇÃO
# -----------------------------------------------------------------------------
CMD ["./terra_allwert-api"]

# =============================================================================
# DEFAULT TARGET - PADRÃO APONTA PARA PRODUÇÃO
# =============================================================================
FROM prd AS default

# Este target é usado quando nenhum target específico é especificado
# Por segurança, o padrão é sempre produção