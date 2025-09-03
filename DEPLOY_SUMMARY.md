# 🚀 Terra Allwert - Resumo do Deploy em Produção

## ✅ Melhorias Implementadas

### 1. **Subdomínios Dedicados**
- **API Principal**: `terra-allwert.online`
- **MinIO S3**: `minio.terra-allwert.online`
- **MinIO Console**: `minio.terra-allwert.online/console`
- **PostgreSQL**: `db.terra-allwert.online:5432` (apenas via túnel SSH)
- **Traefik Dashboard**: `traefik.terra-allwert.online`

### 2. **Configuração Traefik Avançada**
- ✅ SSL/TLS automático com Let's Encrypt
- ✅ Redirect HTTP → HTTPS automático
- ✅ Load balancing para múltiplas réplicas
- ✅ Health checks configurados
- ✅ Dashboard protegido com autenticação

### 3. **Segurança Aprimorada**
- ✅ PostgreSQL isolado (não exposto diretamente)
- ✅ Senhas fortes em todas as variáveis
- ✅ JWT Secret de 64+ caracteres
- ✅ CORS restrito aos domínios corretos
- ✅ Logs estruturados em JSON

### 4. **Arquivos Criados/Modificados**

#### Configurações Docker
- `docker/docker-compose.swarm.prd.yml` ✅ **Melhorado**
  - Subdomínios dedicados para MinIO
  - Labels Traefik para db.terra-allwert.online
  - Redes internas e externas configuradas
  
- `docker/docker-compose.traefik.yml` ✅ **Novo**
  - Traefik com SSL automático
  - Suporte a múltiplos entry points
  - Dashboard protegido

#### Configurações de Ambiente
- `.env.prd` ✅ **Melhorado**
  - Endpoint externo do MinIO
  - CORS com todos os subdomínios
  - Variáveis PostgreSQL consistentes

#### Scripts e Documentação
- `scripts/deploy-production.sh` ✅ **Melhorado**
  - Deploy automático do Traefik
  - Verificações de pré-requisitos
  - Output com novos endpoints
  
- `DNS_SETUP.md` ✅ **Novo**
  - Guia completo de configuração DNS
  - Exemplos para diferentes provedores
  - Comandos de verificação

## 🚀 Como Fazer o Deploy

### Passo 1: Configurar DNS
```bash
# Configure estes registros DNS:
terra-allwert.online         → IP_DA_VPS
minio.terra-allwert.online   → IP_DA_VPS  
db.terra-allwert.online      → IP_DA_VPS
traefik.terra-allwert.online → IP_DA_VPS
```

### Passo 2: Preparar Ambiente
```bash
# Na VPS, clonar projeto
git clone seu-repositorio
cd terra-allwert/api

# Configurar variáveis de produção
cp .env.prd.example .env.prd
nano .env.prd  # Configure com senhas reais
```

### Passo 3: Deploy Automatizado
```bash
# Executar deploy completo
./scripts/deploy-production.sh
```

## 🌐 Endpoints Disponíveis

Após o deploy, estes serviços estarão disponíveis:

| Serviço | URL | Descrição |
|---------|-----|-----------|
| **API Principal** | https://terra-allwert.online | API REST da aplicação |
| **MinIO S3 API** | https://minio.terra-allwert.online | Storage S3-compatible |
| **MinIO Console** | https://minio.terra-allwert.online/console | Interface web do MinIO |
| **Traefik Dashboard** | https://traefik.terra-allwert.online | Monitoramento do proxy |
| **PostgreSQL** | db.terra-allwert.online:5432 | Banco (via túnel SSH) |

## 🔒 Segurança

### Acesso ao Banco de Dados
```bash
# ✅ SEGURO: Via túnel SSH
ssh -L 5432:localhost:5432 user@SEU_IP_VPS
psql -h localhost -p 5432 -U terraallwert_prd -d terraallwert_production

# ❌ INSEGURO: Exposição direta (desabilitada por padrão)
```

### Firewall Recomendado
```bash
ufw allow 22    # SSH
ufw allow 80    # HTTP
ufw allow 443   # HTTPS
ufw deny 5432   # PostgreSQL bloqueado
ufw deny 9000   # MinIO bloqueado (apenas via Traefik)
ufw enable
```

## 🔧 Comandos de Operação

### Verificar Status
```bash
docker service ls
docker service logs -f terra-allwert-prd_prd-api
```

### Escalar Serviços
```bash
docker service scale terra-allwert-prd_prd-api=3
```

### Backup do Banco
```bash
docker exec $(docker ps -q -f name=terra-allwert-prd_prd-db) \
  pg_dump -U terraallwert_prd terraallwert_production > backup.sql
```

### Atualizar Aplicação
```bash
# Rebuild da imagem
docker build -f docker/Dockerfile -t terra-allwert-api:prd .

# Update do serviço
docker service update --image terra-allwert-api:prd terra-allwert-prd_prd-api
```

## ✅ Checklist Final

- [ ] DNS configurado e propagado
- [ ] Arquivo `.env.prd` com senhas reais
- [ ] Docker Swarm inicializado
- [ ] Firewall configurado
- [ ] Deploy executado com sucesso
- [ ] Todos os serviços rodando (`docker service ls`)
- [ ] SSL funcionando (certificados Let's Encrypt)
- [ ] API respondendo: `curl https://terra-allwert.online/api/health`
- [ ] MinIO acessível: `curl https://minio.terra-allwert.online/minio/health/live`
- [ ] Backup automático configurado

---

**🎉 Pronto! Sua aplicação Terra Allwert está rodando em produção com subdomínios dedicados e SSL automático!**