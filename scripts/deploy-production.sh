#!/bin/bash

# =============================================================================
# Terra Allwert API - Script de Deploy em Produção
# =============================================================================
# Este script realiza o deploy da API Terra Allwert em ambiente de produção
# usando Docker Swarm com as configurações do arquivo .env.prd
# =============================================================================

set -e  # Sair em caso de erro

# Cores para output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Diretório base do projeto
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_DIR="$(dirname "$SCRIPT_DIR")"
ENV_FILE="$PROJECT_DIR/.env.prd"

echo -e "${BLUE}🚀 Terra Allwert - Deploy de Produção${NC}"
echo -e "${BLUE}=====================================${NC}"

# Verificar se o arquivo .env.prd existe
if [ ! -f "$ENV_FILE" ]; then
    echo -e "${RED}❌ Erro: Arquivo .env.prd não encontrado!${NC}"
    echo -e "${YELLOW}💡 Instruções:${NC}"
    echo "1. Copie o arquivo .env.prd.example para .env.prd"
    echo "2. Configure todas as variáveis de ambiente"
    echo "3. Execute este script novamente"
    exit 1
fi

# Carregar variáveis do .env.prd
echo -e "${BLUE}📋 Carregando configurações de produção...${NC}"
export $(grep -v '^#' "$ENV_FILE" | xargs)

# Verificar se as variáveis obrigatórias estão definidas
required_vars=(
    "TERRA_ALLWERT_PRD_DB_USER"
    "TERRA_ALLWERT_PRD_DB_PASSWORD" 
    "TERRA_ALLWERT_PRD_DB_NAME"
    "TERRA_ALLWERT_PRD_MINIO_ACCESS_KEY"
    "TERRA_ALLWERT_PRD_MINIO_SECRET_KEY"
    "TERRA_ALLWERT_PRD_REDIS_PASSWORD"
    "TERRA_ALLWERT_PRD_JWT_SECRET"
)

echo -e "${BLUE}🔍 Verificando variáveis obrigatórias...${NC}"
for var in "${required_vars[@]}"; do
    if [ -z "${!var}" ]; then
        echo -e "${RED}❌ Erro: Variável $var não está definida no .env.prd${NC}"
        exit 1
    fi
done

echo -e "${GREEN}✅ Todas as variáveis obrigatórias estão configuradas${NC}"

# Verificar se o Docker Swarm está inicializado
if ! docker info --format '{{.Swarm.LocalNodeState}}' | grep -q active; then
    echo -e "${YELLOW}⚠️  Docker Swarm não está ativo. Inicializando...${NC}"
    docker swarm init
fi

# Verificar se a rede TerraAllWertNet existe
if ! docker network ls --format '{{.Name}}' | grep -q "^TerraAllWertNet$"; then
    echo -e "${BLUE}🌐 Criando rede TerraAllWertNet...${NC}"
    docker network create -d overlay --attachable TerraAllWertNet
fi

# Build da imagem da API
echo -e "${BLUE}🏗️  Construindo imagem da API...${NC}"
cd "$PROJECT_DIR"
docker build -f docker/Dockerfile -t terra-allwert-api:prd .

# Deploy do stack
echo -e "${BLUE}🚢 Realizando deploy do stack...${NC}"
docker stack deploy \
    --compose-file docker/docker-compose.swarm.prd.yml \
    --env-file "$ENV_FILE" \
    terra-allwert-prd

echo -e "${GREEN}✅ Deploy realizado com sucesso!${NC}"

# Aguardar serviços ficarem prontos
echo -e "${BLUE}⏳ Aguardando serviços ficarem prontos...${NC}"
sleep 30

# Verificar status dos serviços
echo -e "${BLUE}📊 Status dos serviços:${NC}"
docker service ls --filter name=terra-allwert-prd

# Logs dos serviços
echo -e "${BLUE}📋 Logs recentes da API:${NC}"
docker service logs --tail 20 terra-allwert-prd_prd-api

echo ""
echo -e "${GREEN}🎉 Deploy de produção concluído!${NC}"
echo -e "${BLUE}📡 Serviços disponíveis:${NC}"
echo "   • API: http://app.terra-allwert.online"
echo "   • MinIO Console: http://app.terra-allwert.online/minio-console"
echo ""
echo -e "${YELLOW}📋 Comandos úteis:${NC}"
echo "   • Ver serviços: docker service ls"
echo "   • Ver logs: docker service logs -f terra-allwert-prd_prd-api"
echo "   • Escalar serviços: docker service scale terra-allwert-prd_prd-api=3"
echo "   • Parar stack: docker stack rm terra-allwert-prd"
echo ""
echo -e "${BLUE}🔒 Lembre-se de configurar:${NC}"
echo "   • SSL/TLS com certificados válidos"
echo "   • Firewall e regras de segurança"
echo "   • Monitoramento e alertas"
echo "   • Backups automáticos"