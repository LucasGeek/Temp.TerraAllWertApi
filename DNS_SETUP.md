# Configuração DNS - Terra Allwert

## 📋 Registros DNS Necessários

Configure estes registros DNS no seu provedor (Cloudflare, Route53, etc.):

### Registros A (Aponte todos para o IP da sua VPS)

```
# Domínio principal
terra-allwert.online             A    SEU_IP_VPS

# Subdomínios dos serviços
minio.terra-allwert.online       A    SEU_IP_VPS
db.terra-allwert.online          A    SEU_IP_VPS
traefik.terra-allwert.online     A    SEU_IP_VPS

# Opcional - www
www.terra-allwert.online         A    SEU_IP_VPS
```

### Registro CNAME (Alternativo)

Se preferir usar CNAME para os subdomínios:

```
minio.terra-allwert.online       CNAME    terra-allwert.online
db.terra-allwert.online          CNAME    terra-allwert.online  
traefik.terra-allwert.online     CNAME    terra-allwert.online
www.terra-allwert.online         CNAME    terra-allwert.online
```

## 🔧 Exemplo de Configuração por Provedor

### Cloudflare
1. Faça login no Cloudflare
2. Selecione o domínio `terra-allwert.online`
3. Vá em DNS → Records
4. Adicione os registros A conforme tabela acima
5. Certifique-se que o proxy está **desabilitado** (nuvem cinza) para teste inicial

### AWS Route 53
```json
{
    "Changes": [
        {
            "Action": "CREATE",
            "ResourceRecordSet": {
                "Name": "terra-allwert.online",
                "Type": "A",
                "TTL": 300,
                "ResourceRecords": [{"Value": "SEU_IP_VPS"}]
            }
        },
        {
            "Action": "CREATE", 
            "ResourceRecordSet": {
                "Name": "minio.terra-allwert.online",
                "Type": "A",
                "TTL": 300,
                "ResourceRecords": [{"Value": "SEU_IP_VPS"}]
            }
        }
    ]
}
```

### Google Cloud DNS
```bash
# Zona DNS
gcloud dns record-sets transaction start --zone=terra-allwert

# Registros A
gcloud dns record-sets transaction add --zone=terra-allwert \
  --name=terra-allwert.online --type=A --ttl=300 SEU_IP_VPS

gcloud dns record-sets transaction add --zone=terra-allwert \
  --name=minio.terra-allwert.online --type=A --ttl=300 SEU_IP_VPS

# Executar
gcloud dns record-sets transaction execute --zone=terra-allwert
```

## ✅ Verificação DNS

Após configurar, teste a resolução DNS:

```bash
# Testar resolução
nslookup terra-allwert.online
nslookup minio.terra-allwert.online  
nslookup db.terra-allwert.online
nslookup traefik.terra-allwert.online

# Ou usando dig
dig terra-allwert.online
dig minio.terra-allwert.online

# Teste de conectividade
ping terra-allwert.online
```

## 🚀 Depois da Configuração DNS

1. **Aguarde a propagação** (pode levar até 48h, geralmente 5-15 minutos)

2. **Execute o deploy**:
```bash
./scripts/deploy-production.sh
```

3. **Teste os serviços**:
```bash
# API Principal
curl -f https://terra-allwert.online/api/health

# MinIO Health Check  
curl -f https://minio.terra-allwert.online/minio/health/live
```

4. **Acesse os painéis**:
   - API: https://terra-allwert.online
   - MinIO Console: https://minio.terra-allwert.online/console
   - Traefik Dashboard: https://traefik.terra-allwert.online

## 🔒 Configurações de Segurança Recomendadas

### Firewall VPS
```bash
# Permitir apenas portas necessárias
ufw allow 22    # SSH
ufw allow 80    # HTTP  
ufw allow 443   # HTTPS
ufw deny 5432   # PostgreSQL (apenas interno)
ufw deny 9000   # MinIO (apenas via Traefik)
ufw deny 9001   # MinIO Console (apenas via Traefik)
ufw enable
```

### SSL/TLS Automático
O Traefik está configurado com Let's Encrypt para SSL automático.
Os certificados serão gerados automaticamente para todos os domínios.

### Acesso ao Banco de Dados
⚠️ **IMPORTANTE**: O PostgreSQL não deve ser exposto diretamente na internet.
Use túnel SSH para acessar:

```bash
# Túnel SSH para PostgreSQL
ssh -L 5432:localhost:5432 usuario@SEU_IP_VPS

# Agora conecte em localhost:5432
psql -h localhost -p 5432 -U terraallwert_prd -d terraallwert_production
```

## 📞 Troubleshooting DNS

### DNS não resolve
- Verifique se os registros foram salvos corretamente
- Aguarde a propagação (use https://dnschecker.org)
- Teste com diferentes servidores DNS: `nslookup terra-allwert.online 8.8.8.8`

### SSL não funciona
- Aguarde alguns minutos para Let's Encrypt gerar os certificados
- Verifique logs do Traefik: `docker service logs traefik_traefik`
- Confirme que as portas 80 e 443 estão abertas

### Serviços não respondem
- Verifique se o Docker Swarm está rodando: `docker service ls`
- Teste conectividade direta: `curl http://SEU_IP_VPS`
- Verifique logs: `docker service logs -f terra-allwert-prd_prd-api`