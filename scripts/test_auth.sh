#!/bin/bash

# Script de teste para o sistema de autenticação GraphQL
# Uso: ./test_auth.sh

BASE_URL="http://localhost:3000/graphql"
HEADERS="Content-Type: application/json"

echo "🔧 Testando Sistema de Autenticação GraphQL - Terra Allwert API"
echo "=================================================="

# Função para fazer requisição GraphQL
graphql_request() {
    local query="$1"
    local variables="$2" 
    local auth_header="$3"
    
    local headers="$HEADERS"
    if [ -n "$auth_header" ]; then
        headers="$HEADERS; Authorization: Bearer $auth_header"
    fi
    
    # Escapar as aspas e quebras de linha na query
    local escaped_query=$(echo "$query" | sed 's/"/\\"/g' | tr '\n' ' ')
    
    curl -s -X POST "$BASE_URL" \
        -H "$headers" \
        -d "{\"query\": \"$escaped_query\", \"variables\": $variables}" | jq '.'
}

echo "1. 🔑 Testando Login com credenciais padrão..."
LOGIN_QUERY='mutation Login($input: LoginInput!) {
  login(input: $input) {
    token
    refreshToken
    expiresAt
    user {
      id
      username
      email
      role
      active
    }
  }
}'

LOGIN_VARIABLES='{
  "input": {
    "email": "admin@terraallwert.com",
    "password": "admin123"
  }
}'

LOGIN_RESPONSE=$(graphql_request "$LOGIN_QUERY" "$LOGIN_VARIABLES")
echo "$LOGIN_RESPONSE"

# Extrair token do response
TOKEN=$(echo "$LOGIN_RESPONSE" | jq -r '.data.login.token // empty')

if [ -z "$TOKEN" ] || [ "$TOKEN" == "null" ]; then
    echo "❌ ERRO: Login falhou ou servidor não está rodando"
    echo "   Certifique-se que:"
    echo "   1. O servidor está rodando em $BASE_URL"
    echo "   2. As migrations foram executadas"  
    echo "   3. Os seeds foram executados"
    exit 1
fi

echo "✅ Login realizado com sucesso!"
echo "Token: ${TOKEN:0:50}..."
echo

echo "2. 👤 Testando query 'me' com token..."
ME_QUERY='query Me {
  me {
    id
    username
    email
    role
    active
    lastLogin
    createdAt
  }
}'

ME_RESPONSE=$(graphql_request "$ME_QUERY" "{}" "$TOKEN")
echo "$ME_RESPONSE"
echo

echo "3. 📋 Testando query pública 'towers' sem token..."
TOWERS_QUERY='query Towers {
  towers {
    id
    name
    description
    totalApartments
    createdAt
  }
}'

TOWERS_RESPONSE=$(graphql_request "$TOWERS_QUERY" "{}")
echo "$TOWERS_RESPONSE"
echo

echo "4. 👥 Testando query admin 'users' com token..."
USERS_QUERY='query Users {
  users {
    id
    username
    email
    role
    active
    createdAt
  }
}'

USERS_RESPONSE=$(graphql_request "$USERS_QUERY" "{}" "$TOKEN")
echo "$USERS_RESPONSE"
echo

echo "5. 🚫 Testando query 'users' sem token (deve falhar)..."
NO_AUTH_RESPONSE=$(graphql_request "$USERS_QUERY" "{}")
echo "$NO_AUTH_RESPONSE"
echo

echo "6. 🔄 Testando refresh token..."
REFRESH_TOKEN=$(echo "$LOGIN_RESPONSE" | jq -r '.data.login.refreshToken // empty')

if [ -n "$REFRESH_TOKEN" ] && [ "$REFRESH_TOKEN" != "null" ]; then
    REFRESH_QUERY='mutation RefreshToken($refreshToken: String!) {
      refreshToken(refreshToken: $refreshToken) {
        token
        refreshToken
        expiresAt
        user {
          username
          role
        }
      }
    }'
    
    REFRESH_VARIABLES="{\"refreshToken\": \"$REFRESH_TOKEN\"}"
    REFRESH_RESPONSE=$(graphql_request "$REFRESH_QUERY" "$REFRESH_VARIABLES")
    echo "$REFRESH_RESPONSE"
else
    echo "❌ Refresh token não encontrado no response de login"
fi

echo
echo "7. 🏠 Testando query de apartamentos (pública)..."
APARTMENTS_QUERY='query Apartments {
  apartments {
    id
    number
    area
    bedrooms
    suites
    status
    available
    price
  }
}'

APARTMENTS_RESPONSE=$(graphql_request "$APARTMENTS_QUERY" "{}")
echo "$APARTMENTS_RESPONSE"

echo
echo "=================================================="
echo "🎯 Testes concluídos!"
echo
echo "Verificações realizadas:"
echo "✅ Login com credenciais padrão (ADMIN)"
echo "✅ Extração de dados do usuário logado (me)"
echo "✅ Query pública sem autenticação (towers)"  
echo "✅ Query admin com autenticação (users)"
echo "✅ Proteção de endpoint admin sem auth"
echo "✅ Refresh de token"
echo "✅ Query pública de apartamentos"
echo
echo "Credenciais disponíveis:"
echo "  ADMIN:  admin@terraallwert.com / admin123"
echo "  VIEWER: viewer@terraallwert.com / viewer123"
echo "  ADMIN:  admin2@terraallwert.com / admin123" 
echo "  VIEWER: demo@terraallwert.com / demo123"
echo
echo "Se todos os testes passaram, o sistema de autenticação está funcionando! 🚀"
echo "Para usar GraphQL Playground: http://localhost:3000/graphql"