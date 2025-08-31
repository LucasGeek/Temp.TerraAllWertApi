# 📋 API Documentation - Go + GraphQL Migration
## Projeto Terra Allwert - Nova Arquitetura

**Data:** 31 de Agosto de 2025  
**Versão:** 2.0  
**Tecnologias:** Golang + GraphQL + MinIO + PostgreSQL  

---

## 🎯 Objetivos da Nova API

### 1.1 Equivalência com Sistema Legado
Manter **100% das funcionalidades** do sistema Flask original:
- ✅ Gestão completa de Torres, Pavimentos e Apartamentos
- ✅ Sistema de upload e galeria de imagens
- ✅ Sistema de busca avançada de apartamentos
- ✅ Marcadores interativos (pins) em imagens e plantas
- ✅ Sistema de configurações dinâmicas

### 1.2 Melhorias Arquiteturais Críticas

#### 🚫 **NUNCA PROCESSAR QUALQUER ARQUIVO PELA API**
- **URLs Assinadas**: Frontend conecta diretamente ao MinIO via signed URLs
- **Zero File Processing**: API apenas gerencia metadados, nunca processa arquivos
- **Gestão de Metadados**: Go controla informações dos arquivos sem manipulá-los
- **Bulk Operations**: Criação automática de arquivos .zip para downloads em lote

#### 📈 **Melhorias de Performance e Escalabilidade**
- **GraphQL**: Uma única query para múltiplos recursos relacionados
- **PostgreSQL**: Maior robustez e performance que MySQL
- **Concurrent Processing**: Aproveitamento de concorrência nativa do Go
- **Type Safety**: Sistema de tipos forte com GraphQL schemas

---

## 🏗️ Arquitetura da Nova API

### 2.1 Stack Tecnológica

```
┌─────────────────────────────────────────────────────────────┐
│                    Flutter Frontend                         │
│          (Conecta diretamente ao MinIO via signed URLs)     │
└─────────────────────┬───────────────────────────────────────┘
                      │ GraphQL Queries/Mutations
                      ▼
┌─────────────────────────────────────────────────────────────┐
│                 Go + GraphQL API                            │
│  • gqlgen (GraphQL Server)                                  │
│  • Fiber (HTTP Framework)                                   │
│  • GORM (Database ORM)                                      │
│  • Validator (Input validation)                             │
└─────────────────────┬───────────────────────────────────────┘
                      │ Database Queries
                      ▼
┌─────────────────────────────────────────────────────────────┐
│                   PostgreSQL                                │
│  • Dados estruturados (torres, apartamentos, metadados)     │
│  • Índices otimizados para busca                            │
│  • JSONB para dados flexíveis                               │
└─────────────────────┬───────────────────────────────────────┘
                      │ Metadata Management
                      ▼
┌─────────────────────────────────────────────────────────────┐
│                      MinIO                                   │
│  • Armazenamento de todos os arquivos                       │
│  • URLs assinadas para acesso direto                        │
│  • Organização hierárquica de buckets                       │
│  • Bulk .zip creation para downloads                        │
└─────────────────────────────────────────────────────────────┘
```

### 2.2 Estrutura de Pastas
```
│   ├── config/
│   │   └── config.go               # Configurações da aplicação
│   ├── database/
│   │   ├── connection.go           # Setup PostgreSQL
│   │   └── migrations/             # Database migrations
│   ├── graph/
│   │   ├── schema.resolvers.go     # GraphQL resolvers
│   │   ├── schema.graphqls         # GraphQL schema definitions
│   │   └── generated/              # Generated GraphQL code
│   ├── models/
│   │   ├── torre.go               # Torre model
│   │   ├── pavimento.go           # Pavimento model
│   │   ├── apartamento.go         # Apartamento model
│   │   └── gallery.go             # Gallery model
│   ├── services/
│   │   ├── torre_service.go       # Torre business logic
│   │   ├── storage_service.go     # MinIO integration
│   │   ├── search_service.go      # Search functionality
│   │   └── bulk_service.go        # Bulk operations
│   ├── utils/
│   │   ├── validators.go          # Input validation
│   │   ├── response.go            # Response helpers
│   │   └── auth.go                # Authentication
│   └── middleware/
│       ├── auth.go                # Auth middleware
│       ├── cors.go                # CORS configuration
│       └── logging.go             # Request logging
│   └── storage/
│       └── minio.go               # MinIO client wrapper
├── scripts/
│   ├── migration.sql              # Database migration scripts
│   └── seed.sql                   # Seed data
```

---

## 📊 GraphQL Schema Definition

### 3.1 Core Types

```graphql
# Building/Tower (Torre)
type Tower {
  id: ID!
  name: String!              # nome da torre
  description: String        # descrição da torre
  floors: [Floor!]!          # lista de pavimentos
  totalApartments: Int!      # total de apartamentos
  createdAt: Time!           # data de criação
  updatedAt: Time!           # última atualização
}

# Floor (Pavimento)
type Floor {
  id: ID!
  number: String!            # número do pavimento (ex: "1", "Térreo", "Cobertura")
  tower: Tower!              # torre a qual pertence
  towerId: ID!
  bannerUrl: String          # URL de banner opcional
  bannerMetadata: FileMetadata
  apartments: [Apartment!]!  # lista de apartamentos
  totalApartments: Int!      # total de apartamentos nesse pavimento
  createdAt: Time!
  updatedAt: Time!
}

# Apartment (Apartamento)
type Apartment {
  id: ID!
  number: String!            # número do apartamento
  area: String               # área do apartamento
  suites: Int                # quantidade de suítes
  bedrooms: Int              # quantidade de dormitórios
  parkingSpots: Int          # quantidade de vagas de garagem
  status: ApartmentStatus!   # status atual
  floor: Floor!              # pavimento
  floorId: ID!
  mainImageUrl: String       # imagem principal
  floorPlanUrl: String       # planta baixa
  solarPosition: String      # posição solar
  price: Float               # preço
  available: Boolean!        # disponível para venda?
  mainImageMetadata: FileMetadata
  floorPlanMetadata: FileMetadata
  images: [ApartmentImage!]! # galeria de imagens do apartamento
  createdAt: Time!
  updatedAt: Time!
}

# Apartment Status (Status do Apartamento)
enum ApartmentStatus {
  AVAILABLE     # disponível
  RESERVED      # reservado
  SOLD          # vendido
  MAINTENANCE   # em manutenção
}

# Apartment Image (Imagem do Apartamento)
type ApartmentImage {
  id: ID!
  apartment: Apartment!
  apartmentId: ID!
  imageUrl: String!          # URL da imagem
  imageMetadata: FileMetadata!
  description: String        # descrição opcional
  order: Int!                # ordem de exibição
  createdAt: Time!
}

# Gallery Image (Imagem de Galeria)
type GalleryImage {
  id: ID!
  route: String!             # rota/slug de navegação
  imageUrl: String!
  thumbnailUrl: String
  imageMetadata: FileMetadata!
  thumbnailMetadata: FileMetadata
  title: String
  description: String
  displayOrder: Int!         # ordem de exibição
  pins: [ImagePin!]!         # marcadores interativos
  createdAt: Time!
  updatedAt: Time!
}

# Interactive Pins (Marcadores Interativos)
type ImagePin {
  id: ID!
  galleryImage: GalleryImage!
  galleryImageId: ID!
  xCoord: Float!             # coordenada X
  yCoord: Float!             # coordenada Y
  title: String
  description: String
  apartment: Apartment       # ligação com apartamento
  apartmentId: ID
  linkUrl: String
  createdAt: Time!
}

# File Metadata (Metadados de Arquivo)
type FileMetadata {
  fileName: String!
  fileSize: Int!
  contentType: String!
  uploadedAt: Time!
  checksum: String
  width: Int
  height: Int
}

# Application Config (Configurações da Aplicação)
type AppConfig {
  logoUrl: String
  apiBaseUrl: String!
  minioBaseUrl: String!
  appVersion: String!
  cacheControlMaxAge: Int!
  updatedAt: Time!
}

# Signed Upload URL (URL assinada para upload direto)
type SignedUploadUrl {
  uploadUrl: String!
  accessUrl: String!
  expiresIn: Int!
  fields: JSON
}

# Bulk Download Info (Informações de Download em Lote)
type BulkDownload {
  downloadUrl: String!
  fileName: String!
  fileSize: Int!
  expiresIn: Int!
  createdAt: Time!
}

# Custom Scalars
scalar Time   # datas e horários
scalar JSON   # campos dinâmicos
```

### 3.2 Input Types

```graphql
# Tower Input
input CreateTowerInput {
  name: String!              # nome da torre
  description: String        # descrição da torre
}

input UpdateTowerInput {
  id: ID!
  name: String               # nome da torre
  description: String        # descrição da torre
}

# Floor Input
input CreateFloorInput {
  number: String!            # número do pavimento
  towerId: ID!               # ID da torre
  bannerUpload: Upload       # upload do banner
}

input UpdateFloorInput {
  id: ID!
  number: String             # número do pavimento
  bannerUpload: Upload       # upload do banner
}

# Apartment Input
input CreateApartmentInput {
  number: String!            # número do apartamento
  floorId: ID!               # ID do pavimento
  area: String               # área do apartamento
  suites: Int                # quantidade de suítes
  bedrooms: Int              # quantidade de dormitórios
  parkingSpots: Int          # quantidade de vagas
  status: ApartmentStatus    # status do apartamento
  solarPosition: String      # posição solar
  price: Float               # preço
  available: Boolean         # disponível?
  mainImageUpload: Upload    # upload da imagem principal
  floorPlanUpload: Upload    # upload da planta baixa
}

input UpdateApartamentoInput {
  id: ID!
  numero: String
  area: String
  suites: Int
  dormitorios: Int
  vagas: Int
  status: ApartamentoStatus
  posicaoSolar: String
  preco: Float
  disponivel: Boolean
  mainImageUpload: Upload
  floorPlanUpload: Upload
}

# Busca de Apartamentos Input
input ApartamentoSearchInput {
  numero: String
  suites: Int
  dormitorios: Int
  vagas: Int
  posicaoSolar: String
  torreId: ID
  pavimentoId: ID
  precoMin: Float
  precoMax: Float
  areaMin: String
  areaMax: String
  status: ApartamentoStatus
  disponivel: Boolean
  limit: Int
  offset: Int
}

# Gallery Input
input CreateGalleryImageInput {
  route: String!
  imageUpload: Upload!
  thumbnailUpload: Upload
  title: String
  description: String
  displayOrder: Int
}

input UpdateGalleryImageInput {
  id: ID!
  title: String
  description: String
  displayOrder: Int
  imageUpload: Upload
  thumbnailUpload: Upload
}

# Image Pin Input
input CreateImagePinInput {
  galleryImageId: ID!
  xCoord: Float!
  yCoord: Float!
  title: String
  description: String
  apartamentoId: ID
  linkUrl: String
}

input UpdateImagePinInput {
  id: ID!
  xCoord: Float
  yCoord: Float
  title: String
  description: String
  apartamentoId: ID
  linkUrl: String
}

# Upload Input
input Upload {
  file: Upload!
}
```

### 3.3 Query Operations

```graphql
type Query {
  # Torres
  torres: [Torre!]!
  torre(id: ID!): Torre
  
  # Pavimentos
  pavimentos(torreId: ID): [Pavimento!]!
  pavimento(id: ID!): Pavimento
  
  # Apartamentos
  apartments(floorId: ID): [Apartment!]!               # lista apartamentos de um pavimento
  apartment(id: ID!): Apartment                        # busca apartamento por ID
  searchApartments(input: ApartmentSearchInput!): [Apartment!]! # busca apartamentos com filtros
  
  # Gallery (Galeria)
  galleryImages(route: String): [GalleryImage!]!       # lista imagens por rota
  galleryImage(id: ID!): GalleryImage                  # busca imagem por ID
  galleryRoutes: [String!]!                            # lista todas as rotas disponíveis
  
  # Image Pins (Marcadores)
  imagePins(galleryImageId: ID!): [ImagePin!]!         # lista pins de uma imagem
  imagePin(id: ID!): ImagePin                          # busca pin por ID
  
  # Configuration (Configurações)
  appConfig: AppConfig!                                # configurações da aplicação
  
  # File Management (Gestão de Arquivos)
  generateSignedUploadUrl(fileName: String!, contentType: String!, folder: String!): SignedUploadUrl! # gera URL assinada para upload
  generateBulkDownload(towerId: ID): BulkDownload!     # gera download em lote
}
```

### 3.4 Mutation Operations

```graphql
type Mutation {
  # Towers (Torres)
  createTower(input: CreateTowerInput!): Tower!        # criar nova torre
  updateTower(input: UpdateTowerInput!): Tower!        # atualizar torre
  deleteTower(id: ID!): Boolean!                       # deletar torre
  
  # Floors (Pavimentos)
  createFloor(input: CreateFloorInput!): Floor!        # criar novo pavimento
  updateFloor(input: UpdateFloorInput!): Floor!        # atualizar pavimento
  deleteFloor(id: ID!): Boolean!                       # deletar pavimento
  
  # Apartments (Apartamentos)
  createApartment(input: CreateApartmentInput!): Apartment!         # criar novo apartamento
  updateApartment(input: UpdateApartmentInput!): Apartment!         # atualizar apartamento
  deleteApartment(id: ID!): Boolean!                               # deletar apartamento
  addApartmentImage(apartmentId: ID!, imageUpload: Upload!, description: String): ApartmentImage! # adicionar imagem ao apartamento
  removeApartmentImage(imageId: ID!): Boolean!                     # remover imagem do apartamento
  reorderApartmentImages(apartmentId: ID!, imageIds: [ID!]!): [ApartmentImage!]! # reordenar imagens do apartamento
  
  # Gallery (Galeria)
  createGalleryImage(input: CreateGalleryImageInput!): GalleryImage!       # criar nova imagem na galeria
  updateGalleryImage(input: UpdateGalleryImageInput!): GalleryImage!       # atualizar imagem da galeria
  deleteGalleryImage(id: ID!): Boolean!                                    # deletar imagem da galeria
  reorderGalleryImages(route: String!, imageIds: [ID!]!): [GalleryImage!]! # reordenar imagens da galeria
  
  # Image Pins (Marcadores)
  createImagePin(input: CreateImagePinInput!): ImagePin!                   # criar novo marcador
  updateImagePin(input: UpdateImagePinInput!): ImagePin!                   # atualizar marcador
  deleteImagePin(id: ID!): Boolean!                                        # deletar marcador
  
  # Configuration (Configurações)
  updateAppConfig(logoUpload: Upload, apiBaseUrl: String, cacheControlMaxAge: Int): AppConfig! # atualizar configurações
}
```

### 3.5 Subscription Operations

```graphql
type Subscription {
  # Real-time updates for apartment availability (Atualizações em tempo real)
  apartmentStatusChanged(towerId: ID): Apartment!                          # mudanças de status de apartamento
  
  # Gallery updates (Atualizações da galeria)
  galleryImageAdded(route: String!): GalleryImage!                        # imagem adicionada à galeria
  galleryImageRemoved(route: String!): ID!                                # imagem removida da galeria
  
  # Bulk download progress (Progresso de download em lote)
  bulkDownloadProgress(downloadId: ID!): BulkDownloadProgress!             # progresso do download
}

type BulkDownloadProgress {
  downloadId: ID!
  progress: Float!           # progresso de 0.0 a 1.0
  status: BulkDownloadStatus! # status atual
  message: String            # mensagem opcional
}

enum BulkDownloadStatus {
  PREPARING     # preparando
  COMPRESSING   # comprimindo
  UPLOADING     # enviando
  COMPLETED     # completo
  FAILED        # falhou
}
```

---

## 🔄 API Endpoints & Examples

### 4.1 Query Towers with Apartments (Consultar Torres com Apartamentos)

**Query:**
```graphql
query GetTowersWithApartments {
  towers {
    id
    name                 # nome da torre
    description          # descrição
    totalApartments      # total de apartamentos
    floors {
      id
      number             # número do pavimento
      bannerUrl          # URL do banner
      totalApartments    # total de apartamentos no pavimento
      apartments {
        id
        number           # número do apartamento
        area             # área
        suites           # suítes
        status           # status
        price            # preço
        available        # disponível?
        mainImageUrl     # imagem principal
      }
    }
  }
}
```

**Response:**
```json
{
  "data": {
    "towers": [
      {
        "id": "1",
        "name": "Torre 1",
        "description": "Torre residencial com vista para o mar",
        "totalApartments": 120,
        "floors": [
          {
            "id": "101",
            "number": "1º Pavimento",
            "bannerUrl": "https://minio.example.com/terra-allwert/tower1/floor1/banner.jpg",
            "totalApartments": 4,
            "apartments": [
              {
                "id": "1001",
                "number": "101",
                "area": "85m²",
                "suites": 2,
                "status": "AVAILABLE",
                "price": 450000.0,
                "available": true,
                "mainImageUrl": "https://minio.example.com/terra-allwert/apartments/1001/main.jpg"
              }
            ]
          }
        ]
      }
    ]
  }
}
```

### 4.2 Advanced Apartment Search (Busca Avançada de Apartamentos)

**Query:**
```graphql
query SearchApartments($input: ApartmentSearchInput!) {
  searchApartments(input: $input) {
    id
    number           # número
    area             # área
    suites           # suítes
    bedrooms         # dormitórios
    parkingSpots     # vagas
    status           # status
    price            # preço
    solarPosition    # posição solar
    floor {
      id
      number         # número do pavimento
      tower {
        id
        name         # nome da torre
      }
    }
    mainImageUrl     # imagem principal
    floorPlanUrl     # planta baixa
  }
}
```

**Variables:**
```json
{
  "input": {
    "suites": 2,
    "towerId": "1",
    "priceMax": 500000.0,
    "available": true,
    "limit": 20,
    "offset": 0
  }
}
```

### 4.3 Create Apartment (Criar Apartamento)

**Mutation:**
```graphql
mutation CreateApartment($input: CreateApartmentInput!) {
  createApartment(input: $input) {
    id
    number           # número
    area             # área
    suites           # suítes
    status           # status
    price            # preço
    floor {
      id
      number         # número do pavimento
      tower {
        name         # nome da torre
      }
    }
    mainImageUrl     # imagem principal
    floorPlanUrl     # planta baixa
  }
}
```

**Variables:**
```json
{
  "input": {
    "number": "205",
    "floorId": "102",
    "area": "95m²",
    "suites": 2,
    "bedrooms": 3,
    "parkingSpots": 2,
    "status": "AVAILABLE",
    "solarPosition": "Norte/Sul",
    "price": 485000.0,
    "available": true
  }
}
```

### 4.4 Direct MinIO Upload (Upload Direto para MinIO)

**Step 1 - Get Signed URL:**
```graphql
query GetSignedUploadUrl($fileName: String!, $contentType: String!, $folder: String!) {
  generateSignedUploadUrl(
    fileName: $fileName,
    contentType: $contentType,
    folder: $folder
  ) {
    uploadUrl
    accessUrl
    expiresIn
    fields
  }
}
```

**Variables:**
```json
{
  "fileName": "planta_apartamento_205.jpg",
  "contentType": "image/jpeg",
  "folder": "apartamentos/205/floor-plans"
}
```

**Step 2 - Frontend Uploads Directly:**
```javascript
// Frontend code - Direct upload to MinIO
const formData = new FormData();
formData.append('file', file);

// Add any required fields from signed URL response
Object.entries(signedUrlResponse.fields || {}).forEach(([key, value]) => {
  formData.append(key, value);
});

const uploadResponse = await fetch(signedUrlResponse.uploadUrl, {
  method: 'POST',
  body: formData
});

// File is now available at signedUrlResponse.accessUrl
```

### 4.5 Bulk Download de Torre

**Query:**
```graphql
query GenerateBulkDownload($torreId: ID!) {
  generateBulkDownload(torreId: $torreId) {
    downloadUrl
    fileName
    fileSize
    expiresIn
  }
}
```

**Variables:**
```json
{
  "towerId": "1"
}
```

---

## 🗄️ Database Schema (PostgreSQL)

### 5.1 Tables Structure

```sql
-- Towers table (Torres)
CREATE TABLE towers (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL UNIQUE,          -- nome da torre
    description TEXT,                           -- descrição da torre
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Floors table (Pavimentos)
CREATE TABLE floors (
    id SERIAL PRIMARY KEY,
    number VARCHAR(100) NOT NULL,               -- número do pavimento
    tower_id INTEGER NOT NULL REFERENCES towers(id) ON DELETE CASCADE,
    banner_url TEXT,                            -- URL do banner
    banner_metadata JSONB,                      -- metadados do banner
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(tower_id, number)
);

-- Apartments table (Apartamentos)
CREATE TABLE apartments (
    id SERIAL PRIMARY KEY,
    number VARCHAR(50) NOT NULL,                -- número do apartamento
    floor_id INTEGER NOT NULL REFERENCES floors(id) ON DELETE CASCADE,
    area VARCHAR(50),                           -- área do apartamento
    suites INTEGER DEFAULT 0,                   -- quantidade de suítes
    bedrooms INTEGER DEFAULT 0,                 -- quantidade de dormitórios
    parking_spots INTEGER DEFAULT 0,            -- quantidade de vagas
    status apartment_status DEFAULT 'AVAILABLE', -- status do apartamento
    solar_position VARCHAR(100),                -- posição solar
    price DECIMAL(12,2),                        -- preço
    available BOOLEAN DEFAULT true,             -- disponível para venda?
    main_image_url TEXT,                        -- URL da imagem principal
    main_image_metadata JSONB,                  -- metadados da imagem principal
    floor_plan_url TEXT,                        -- URL da planta baixa
    floor_plan_metadata JSONB,                  -- metadados da planta baixa
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(floor_id, number)
);

-- Apartment status enum (Status do apartamento)
CREATE TYPE apartment_status AS ENUM ('AVAILABLE', 'RESERVED', 'SOLD', 'MAINTENANCE');

-- Apartment images table (Imagens do apartamento)
CREATE TABLE apartment_images (
    id SERIAL PRIMARY KEY,
    apartment_id INTEGER NOT NULL REFERENCES apartments(id) ON DELETE CASCADE,
    image_url TEXT NOT NULL,                    -- URL da imagem
    image_metadata JSONB NOT NULL,              -- metadados da imagem
    description TEXT,                           -- descrição da imagem
    order_index INTEGER NOT NULL DEFAULT 0,     -- ordem de exibição
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Gallery images table (Imagens da galeria)
CREATE TABLE gallery_images (
    id SERIAL PRIMARY KEY,
    route VARCHAR(255) NOT NULL,               -- rota de navegação
    image_url TEXT NOT NULL,                   -- URL da imagem
    image_metadata JSONB NOT NULL,             -- metadados da imagem
    thumbnail_url TEXT,                        -- URL da miniatura
    thumbnail_metadata JSONB,                  -- metadados da miniatura
    title VARCHAR(255),                        -- título
    description TEXT,                          -- descrição
    display_order INTEGER DEFAULT 0,           -- ordem de exibição
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Image pins table (Marcadores interativos)
CREATE TABLE image_pins (
    id SERIAL PRIMARY KEY,
    gallery_image_id INTEGER NOT NULL REFERENCES gallery_images(id) ON DELETE CASCADE,
    x_coord DECIMAL(8,6) NOT NULL,             -- coordenada X
    y_coord DECIMAL(8,6) NOT NULL,             -- coordenada Y
    title VARCHAR(255),                        -- título
    description TEXT,                          -- descrição
    apartment_id INTEGER REFERENCES apartments(id) ON DELETE SET NULL, -- referência ao apartamento
    link_url TEXT,                             -- URL de link
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- App configuration table (Configurações da aplicação)
CREATE TABLE app_configs (
    id SERIAL PRIMARY KEY,
    logo_url TEXT,                             -- URL do logo
    api_base_url TEXT NOT NULL,                -- URL base da API
    minio_base_url TEXT NOT NULL,              -- URL base do MinIO
    app_version VARCHAR(50) NOT NULL DEFAULT '2.0.0', -- versão da aplicação
    cache_control_max_age INTEGER DEFAULT 3600, -- cache control
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Bulk downloads tracking (Rastreamento de downloads em lote)
CREATE TABLE bulk_downloads (
    id SERIAL PRIMARY KEY,
    tower_id INTEGER REFERENCES towers(id) ON DELETE CASCADE,
    download_url TEXT NOT NULL,                -- URL de download
    file_name VARCHAR(255) NOT NULL,           -- nome do arquivo
    file_size BIGINT NOT NULL,                 -- tamanho do arquivo
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL, -- expira em
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Indexes for performance (Índices para performance)
CREATE INDEX idx_floors_tower_id ON floors(tower_id);
CREATE INDEX idx_apartments_floor_id ON apartments(floor_id);
CREATE INDEX idx_apartments_status ON apartments(status);
CREATE INDEX idx_apartments_available ON apartments(available);
CREATE INDEX idx_apartments_search ON apartments(suites, bedrooms, parking_spots, price);
CREATE INDEX idx_gallery_images_route ON gallery_images(route);
CREATE INDEX idx_gallery_images_display_order ON gallery_images(route, display_order);
CREATE INDEX idx_image_pins_gallery_image_id ON image_pins(gallery_image_id);
CREATE INDEX idx_apartment_images_apartment_id ON apartment_images(apartment_id);
CREATE INDEX idx_apartment_images_order ON apartment_images(apartment_id, order_index);

-- Full-text search index for apartments (busca textual)
CREATE INDEX idx_apartments_search_text ON apartments 
USING GIN (to_tsvector('portuguese', number || ' ' || COALESCE(area, '') || ' ' || COALESCE(solar_position, '')));
```

---

## 🚀 MinIO Integration & File Management

### 6.1 Bucket Structure

```
terra-allwert/
├── towers/                    # torres
│   ├── tower-1/               # torre 1
│   │   ├── banner.jpg
│   │   ├── floors/             # pavimentos
│   │   │   ├── floor-1/        # pavimento 1
│   │   │   │   ├── banner.jpg
│   │   │   │   └── plans/      # plantas
│   │   │   └── floor-2/        # pavimento 2
│   │   └── bulk-downloads/     # downloads em lote
│   │       └── tower-1-complete-20250831.zip
│   └── tower-2/               # torre 2
├── apartments/               # apartamentos
│   ├── 101/
│   │   ├── main-image.jpg     # imagem principal
│   │   ├── floor-plan.jpg     # planta baixa
│   │   └── gallery/           # galeria
│   │       ├── image-1.jpg
│   │       └── image-2.jpg
│   └── 102/
├── gallery/                  # galeria geral
│   ├── home/                 # página inicial
│   │   ├── image-1.jpg
│   │   ├── thumbnail-1.jpg    # miniaturas
│   │   └── image-2.jpg
│   ├── manifesto-terra/
│   ├── common-areas/         # áreas comuns
│   └── lagoon/
├── configurations/          # configurações
│   └── logo.png
└── temp/                     # temporários
    └── bulk-processing/      # processamento em lote
```

### 6.2 Signed URL Generation

**Go Service Example:**
```go
type StorageService struct {
    minioClient *minio.Client
    bucketName  string
}

func (s *StorageService) GenerateSignedUploadURL(fileName, contentType, folder string) (*SignedUploadURL, error) {
    objectName := fmt.Sprintf("%s/%s", folder, fileName)
    
    // Create presigned URL for PUT operation
    presignedURL, err := s.minioClient.PresignedPutObject(
        context.Background(),
        s.bucketName,
        objectName,
        time.Hour*1, // 1 hour expiry
    )
    if err != nil {
        return nil, err
    }
    
    // Access URL (without query parameters)
    accessURL := fmt.Sprintf("https://%s/%s/%s", 
        s.minioClient.EndpointURL().Host, 
        s.bucketName, 
        objectName)
    
    return &SignedUploadURL{
        UploadURL: presignedURL.String(),
        AccessURL: accessURL,
        ExpiresIn: 3600,
    }, nil
}
```

### 6.3 Bulk Download Service

**Go Implementation:**
```go
func (s *BulkService) GenerateTorreBulkDownload(torreId string) (*BulkDownload, error) {
    // 1. Query all torre data
    torre, err := s.torreService.GetTorreComplete(torreId)
    if err != nil {
        return nil, err
    }
    
    // 2. Create temporary directory
    tempDir := fmt.Sprintf("temp/bulk-processing/%s-%d", torreId, time.Now().Unix())
    
    // 3. Download all files to temp directory (concurrent)
    var wg sync.WaitGroup
    sem := make(chan struct{}, 10) // Limit concurrent downloads
    
    for _, pavimento := range torre.Pavimentos {
        for _, apartamento := range pavimento.Apartamentos {
            wg.Add(1)
            go func(apt *Apartamento) {
                defer wg.Done()
                sem <- struct{}{}
                defer func() { <-sem }()
                
                s.downloadApartamentoFiles(apt, tempDir)
            }(apartamento)
        }
    }
    wg.Wait()
    
    // 4. Create ZIP file
    zipFileName := fmt.Sprintf("torre-%s-complete-%s.zip", torreId, time.Now().Format("20060102"))
    zipPath := fmt.Sprintf("torres/torre-%s/bulk-downloads/%s", torreId, zipFileName)
    
    err = s.createZipFile(tempDir, zipPath)
    if err != nil {
        return nil, err
    }
    
    // 5. Generate download URL
    downloadURL, err := s.storageService.GenerateSignedDownloadURL(zipPath, time.Hour*24)
    if err != nil {
        return nil, err
    }
    
    // 6. Clean up temp directory
    os.RemoveAll(tempDir)
    
    return &BulkDownload{
        DownloadURL: downloadURL,
        FileName:    zipFileName,
        FileSize:    s.getFileSize(zipPath),
        ExpiresIn:   86400, // 24 hours
    }, nil
}
```

---

## 🔐 Authentication & Authorization

### 7.1 JWT Authentication

**JWT Claims Structure:**
```go
type Claims struct {
    UserID   string   `json:"user_id"`
    Username string   `json:"username"`
    Roles    []string `json:"roles"`
    jwt.StandardClaims
}
```

**Supported Roles:**
- `admin`: Full access to all operations
- `viewer`: Read-only access
- `editor`: Can modify content but not delete
- `uploader`: Can upload files and modify galleries

### 7.2 GraphQL Authentication Directive

**Schema Directive:**
```graphql
directive @auth(requires: Role = USER) on OBJECT | FIELD_DEFINITION

enum Role {
  USER
  EDITOR
  ADMIN
}

# Usage examples
type Mutation @auth(requires: ADMIN) {
  deleteTorre(id: ID!): Boolean!
  updateAppConfig(...): AppConfig!
}

type Query {
  torres: [Torre!]! # Public access
  
  apartamento(id: ID!): Apartamento @auth(requires: USER)
  
  generateSignedUploadUrl(...): SignedUploadUrl! @auth(requires: EDITOR)
}
```

---

## 📈 Performance & Caching

### 8.1 Query Optimization Strategies

**DataLoader Pattern:**
```go
type Loaders struct {
    TorreLoader      *dataloader.Loader
    PavimentoLoader  *dataloader.Loader
    ApartamentoLoader *dataloader.Loader
    GalleryLoader    *dataloader.Loader
}

// Batch loading implementation
func (l *Loaders) LoadApartamentosByPavimento(pavimentoIDs []string) ([]*Apartamento, []error) {
    // Single query to load all apartamentos for multiple pavimentos
    // Returns results in same order as input IDs
}
```

**Database Query Optimization:**
```sql
-- Optimized apartment search with full-text search
SELECT a.*, p.numero as pavimento_numero, t.nome as torre_nome
FROM apartamentos a
JOIN pavimentos p ON p.id = a.pavimento_id
JOIN torres t ON t.id = p.torre_id
WHERE ($1::int IS NULL OR a.suites = $1)
  AND ($2::int IS NULL OR t.id = $2)
  AND ($3::decimal IS NULL OR a.preco <= $3)
  AND ($4::boolean IS NULL OR a.disponivel = $4)
  AND ($5::text IS NULL OR to_tsvector('portuguese', a.numero || ' ' || COALESCE(a.posicao_solar, '')) @@ plainto_tsquery('portuguese', $5))
ORDER BY t.nome, p.numero, a.numero
LIMIT $6 OFFSET $7;
```

### 8.2 Redis Caching Strategy

**Cache Keys Structure:**
```
terra-allwert:torres:all                          # List of all torres
terra-allwert:torre:{id}                          # Individual torre data  
terra-allwert:torre:{id}:pavimentos               # Torre's pavimentos
terra-allwert:pavimento:{id}:apartamentos         # Pavimento's apartamentos
terra-allwert:search:apartamentos:{hash}          # Search results
terra-allwert:gallery:{route}                     # Gallery images by route
terra-allwert:config:app                          # App configuration
```

**Go Implementation:**
```go
type CacheService struct {
    redis  *redis.Client
    defaultTTL time.Duration
}

func (c *CacheService) GetOrSetTorres(fetcher func() ([]*Torre, error)) ([]*Torre, error) {
    key := "terra-allwert:torres:all"
    
    // Try cache first
    cached, err := c.redis.Get(context.Background(), key).Result()
    if err == nil {
        var torres []*Torre
        if err := json.Unmarshal([]byte(cached), &torres); err == nil {
            return torres, nil
        }
    }
    
    // Cache miss - fetch from database
    torres, err := fetcher()
    if err != nil {
        return nil, err
    }
    
    // Store in cache
    data, _ := json.Marshal(torres)
    c.redis.SetEX(context.Background(), key, data, c.defaultTTL)
    
    return torres, nil
}
```

---

## 🚨 Error Handling & Monitoring

### 9.1 GraphQL Error Extensions

**Structured Error Response:**
```go
type ErrorExtension struct {
    Code    string                 `json:"code"`
    Details map[string]interface{} `json:"details,omitempty"`
}

// Error codes
const (
    ErrorCodeValidation     = "VALIDATION_ERROR"
    ErrorCodeNotFound      = "NOT_FOUND"
    ErrorCodeUnauthorized  = "UNAUTHORIZED"
    ErrorCodeFileUpload    = "FILE_UPLOAD_ERROR"
    ErrorCodeStorageError  = "STORAGE_ERROR"
    ErrorCodeDatabaseError = "DATABASE_ERROR"
)

// Usage in resolver
func (r *mutationResolver) CreateApartamento(ctx context.Context, input CreateApartamentoInput) (*Apartamento, error) {
    if err := validateApartamentoInput(input); err != nil {
        return nil, &gqlerror.Error{
            Message: "Validation failed",
            Extensions: map[string]interface{}{
                "code": ErrorCodeValidation,
                "details": map[string]interface{}{
                    "field": err.Field,
                    "reason": err.Reason,
                },
            },
        }
    }
    
    // ... rest of implementation
}
```

### 9.2 Structured Logging

**Log Structure:**
```go
type LogEntry struct {
    Level        string                 `json:"level"`
    Timestamp    time.Time              `json:"timestamp"`
    Message      string                 `json:"message"`
    Operation    string                 `json:"operation"`
    UserID       string                 `json:"user_id,omitempty"`
    RequestID    string                 `json:"request_id"`
    Duration     time.Duration          `json:"duration,omitempty"`
    Error        string                 `json:"error,omitempty"`
    GraphQLQuery string                 `json:"graphql_query,omitempty"`
    Metadata     map[string]interface{} `json:"metadata,omitempty"`
}

// Usage
logger.Info("Apartamento created successfully").
    WithField("operation", "CreateApartamento").
    WithField("apartamento_id", apartamento.ID).
    WithField("user_id", userID).
    WithField("duration", time.Since(start)).
    Log()
```

### 9.3 Health Checks & Metrics

**Health Check Endpoints:**
```go
type HealthCheck struct {
    Status    string                 `json:"status"`
    Version   string                 `json:"version"`
    Timestamp time.Time              `json:"timestamp"`
    Services  map[string]ServiceHealth `json:"services"`
}

type ServiceHealth struct {
    Status      string        `json:"status"`
    ResponseTime time.Duration `json:"response_time"`
    LastError   string        `json:"last_error,omitempty"`
}

// GET /health
func (s *Server) HealthHandler(c *fiber.Ctx) error {
    health := &HealthCheck{
        Status:    "healthy",
        Version:   s.config.Version,
        Timestamp: time.Now(),
        Services: map[string]ServiceHealth{
            "database": s.checkDatabase(),
            "redis":    s.checkRedis(),
            "minio":    s.checkMinIO(),
        },
    }
    
    return c.JSON(health)
}
```

---

## 📋 Migration Plan from Legacy System

### 10.1 Data Migration Strategy

**Phase 1: Database Migration**
```sql
-- Migration script from MySQL to PostgreSQL
INSERT INTO torres (nome, descricao, created_at)
SELECT nome, NULL, created_at FROM legacy_mysql.torres;

INSERT INTO pavimentos (numero, torre_id, banner_url, created_at)
SELECT p.numero, t.id, p.banner_url, p.created_at
FROM legacy_mysql.pavimentos p
JOIN torres t ON t.nome = (SELECT nome FROM legacy_mysql.torres lt WHERE lt.id = p.torre_id);

-- Continue for all tables with proper ID mapping
```

**Phase 2: File Migration**
```go
func MigrateFilesToMinIO() error {
    // 1. List all files in legacy Flask uploads directory
    // 2. Upload each file to MinIO with proper organization
    // 3. Update database URLs to point to MinIO
    // 4. Verify file accessibility
    
    files, err := filepath.Glob("/legacy/uploads/**/*")
    if err != nil {
        return err
    }
    
    for _, file := range files {
        // Determine new MinIO path based on file structure
        minioPath := mapLegacyPathToMinIOPath(file)
        
        // Upload to MinIO
        err := s.storageService.UploadFile(file, minioPath)
        if err != nil {
            log.Printf("Failed to migrate file %s: %v", file, err)
            continue
        }
        
        // Update database references
        err = s.updateDatabaseFileURL(file, minioPath)
        if err != nil {
            log.Printf("Failed to update database for file %s: %v", file, err)
        }
    }
    
    return nil
}
```

### 10.2 API Compatibility Layer

**Legacy REST Endpoints Support:**
```go
// Maintain compatibility with existing Flutter app during migration
func (s *Server) setupLegacyRoutes() {
    // GET /api/torres -> GraphQL torres query
    s.app.Get("/api/torres", s.legacyGetTorres)
    
    // POST /api/apartamentos/search -> GraphQL searchApartamentos
    s.app.Post("/api/apartamentos/search", s.legacySearchApartamentos)
    
    // POST /api/upload -> GraphQL generateSignedUploadUrl + direct upload
    s.app.Post("/api/upload", s.legacyUpload)
}

func (s *Server) legacyGetTorres(c *fiber.Ctx) error {
    // Execute GraphQL query internally
    query := `query { torres { id nome pavimentos { id numero } } }`
    result := s.graphqlHandler.ExecuteQuery(query, nil, c.UserContext())
    
    // Convert GraphQL response to legacy REST format
    legacyResponse := convertTorresToLegacyFormat(result)
    return c.JSON(legacyResponse)
}
```

---

## 🎯 Success Metrics & KPIs

### 11.1 Performance Benchmarks

| Metric | Legacy Flask | Target Go+GraphQL | Improvement |
|--------|--------------|-------------------|-------------|
| **Average Response Time** | 200ms | <50ms | 75% faster |
| **Concurrent Users** | 50 | 500+ | 10x increase |
| **Database Queries per Request** | 5-15 | 1-3 | 70% reduction |
| **Memory Usage** | 512MB | 128MB | 75% reduction |
| **File Upload Speed** | 2MB/s | Direct to MinIO | 5-10x faster |

### 11.2 Reliability Targets

- **Uptime**: 99.9% (43.8 minutes downtime/month max)
- **Error Rate**: <0.1% of requests
- **File Upload Success**: 99.95%
- **Search Response Time**: <100ms for 95% of queries
- **Bulk Download Generation**: <30 seconds for complete torre

### 11.3 Monitoring Dashboards

**Key Metrics to Track:**
```go
// Prometheus metrics
var (
    httpRequestsTotal = prometheus.NewCounterVec(
        prometheus.CounterOpts{
            Name: "http_requests_total",
            Help: "Total number of HTTP requests",
        },
        []string{"method", "endpoint", "status"},
    )
    
    graphqlQueryDuration = prometheus.NewHistogramVec(
        prometheus.HistogramOpts{
            Name: "graphql_query_duration_seconds",
            Help: "GraphQL query execution time",
        },
        []string{"operation", "query"},
    )
    
    fileUploadSize = prometheus.NewHistogramVec(
        prometheus.HistogramOpts{
            Name: "file_upload_size_bytes",
            Help: "Size of uploaded files",
        },
        []string{"file_type"},
    )
)
```

---

## 🚀 Deployment & DevOps

### 12.1 Docker Configuration

**Dockerfile:**
```dockerfile
# Build stage
FROM golang:1.21-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main cmd/server/main.go

# Production stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates tzdata
WORKDIR /root/

COPY --from=builder /app/main .
COPY --from=builder /app/internal/graph/schema.graphqls ./schema.graphqls

CMD ["./main"]
```

**docker-compose.yml:**
```yaml
version: '3.8'
services:
  api:
    build: .
    ports:
      - "8080:8080"
    environment:
      - DATABASE_URL=postgres://user:pass@postgres:5432/terra_allwert
      - REDIS_URL=redis://redis:6379
      - MINIO_ENDPOINT=minio:9000
      - MINIO_ACCESS_KEY=minioadmin
      - MINIO_SECRET_KEY=minioadmin
      - JWT_SECRET=your-jwt-secret-here
    depends_on:
      - postgres
      - redis
      - minio
    restart: unless-stopped

  postgres:
    image: postgres:15-alpine
    environment:
      - POSTGRES_DB=terra_allwert
      - POSTGRES_USER=user
      - POSTGRES_PASSWORD=pass
    volumes:
      - postgres_data:/var/lib/postgresql/data
      - ./scripts/migration.sql:/docker-entrypoint-initdb.d/01-migration.sql
    restart: unless-stopped

  redis:
    image: redis:7-alpine
    volumes:
      - redis_data:/data
    restart: unless-stopped

  minio:
    image: minio/minio:latest
    command: server /data --console-address ":9001"
    ports:
      - "9000:9000"
      - "9001:9001"
    environment:
      - MINIO_ACCESS_KEY=minioadmin
      - MINIO_SECRET_KEY=minioadmin
    volumes:
      - minio_data:/data
    restart: unless-stopped

volumes:
  postgres_data:
  redis_data:
  minio_data:
```

### 12.2 Environment Configuration

```go
type Config struct {
    Port         string `env:"PORT" envDefault:"8080"`
    DatabaseURL  string `env:"DATABASE_URL,required"`
    RedisURL     string `env:"REDIS_URL,required"`
    
    MinIOEndpoint   string `env:"MINIO_ENDPOINT,required"`
    MinIOAccessKey  string `env:"MINIO_ACCESS_KEY,required"`
    MinIOSecretKey  string `env:"MINIO_SECRET_KEY,required"`
    MinIOBucketName string `env:"MINIO_BUCKET_NAME" envDefault:"terra-allwert"`
    
    JWTSecret       string `env:"JWT_SECRET,required"`
    JWTExpirationHours int `env:"JWT_EXPIRATION_HOURS" envDefault:"24"`
    
    CORSAllowedOrigins []string `env:"CORS_ALLOWED_ORIGINS" envSeparator:","`
    
    LogLevel        string `env:"LOG_LEVEL" envDefault:"info"`
    EnablePlayground bool  `env:"ENABLE_PLAYGROUND" envDefault:"false"`
}
```

---

## 📚 Conclusion & Next Steps

### 13.1 Summary of Improvements

**✅ Major Architectural Improvements:**
1. **Zero File Processing**: API nunca processa arquivos, apenas metadados
2. **Direct MinIO Integration**: Frontend conecta diretamente via signed URLs
3. **GraphQL Efficiency**: Single queries para múltiplos recursos relacionados
4. **Type Safety**: Schema forte e validação automática
5. **Bulk Operations**: Geração automática de .zip files para downloads

**✅ Performance Gains:**
- 75% faster response times
- 10x concurrent user capacity
- 70% reduction in database queries
- Direct file uploads (no API bottleneck)

**✅ Scalability & Reliability:**
- Horizontal scaling ready
- Proper error handling and monitoring
- Health checks and metrics
- Redis caching strategy

### 13.2 Implementation Timeline

**Phase 1 (Weeks 1-2): Foundation**
- Setup Go project structure
- Implement core GraphQL schema
- PostgreSQL migration
- Basic CRUD operations

**Phase 2 (Weeks 3-4): Storage Integration**
- MinIO direct upload implementation
- Signed URL generation
- File metadata management
- Bulk download service

**Phase 3 (Weeks 5-6): Advanced Features**
- Search functionality
- Image pins and galleries
- Authentication and authorization
- Caching implementation

**Phase 4 (Weeks 7-8): Production Ready**
- Error handling and monitoring
- Performance optimization
- Load testing
- Documentation and deployment

### 13.3 Migration Strategy

**Parallel Development Approach:**
1. Develop new Go API alongside legacy Flask system
2. Implement compatibility layer for gradual migration
3. Migrate data in phases (torres → pavimentos → apartamentos)
4. Update Flutter app endpoints progressively
5. Sunset legacy system after full validation

This documentation provides a comprehensive roadmap for migrating the Terra Allwert project to a modern, scalable, and maintainable Go+GraphQL architecture while maintaining all existing functionality and adding significant improvements in performance and developer experience.