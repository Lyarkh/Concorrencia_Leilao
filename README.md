# Concorrencia Leilao - Fechamento Automatico

Sistema de leiloes com fechamento automatico via Goroutines. Ao criar um leilao, uma goroutine e disparada para monitorar o tempo e fechar o leilao automaticamente apos a duracao configurada.

## Funcionalidade Implementada

- **Fechamento automatico**: Goroutine iniciada na criacao do leilao que aguarda `AUCTION_DURATION` e atualiza o status para `Completed` no MongoDB
- **Configuravel via env**: Duracao do leilao definida pela variavel `AUCTION_DURATION`
- **Nao-bloqueante**: A goroutine roda em background sem bloquear a thread principal

## Variaveis de Ambiente

| Variavel | Descricao | Default | Exemplo |
|----------|-----------|---------|---------|
| `AUCTION_DURATION` | Tempo ate o fechamento automatico do leilao | `5m` | `20s`, `2m`, `1h` |
| `AUCTION_INTERVAL` | Intervalo de verificacao de leilao no bid | `5m` | `20s` |
| `BATCH_INSERT_INTERVAL` | Intervalo de batch insert de bids | `20s` | `10s` |
| `MAX_BATCH_SIZE` | Tamanho maximo do batch de bids | `4` | `10` |
| `MONGODB_URL` | URL de conexao do MongoDB | - | `mongodb://admin:admin@mongodb:27017/auctions?authSource=admin` |
| `MONGODB_DB` | Nome do banco de dados | - | `auctions` |

## Como Executar

### Subir a aplicacao

```bash
docker compose up -d --build
```

A aplicacao estara disponivel em `http://localhost:8080`.

### Parar a aplicacao

```bash
docker compose down
```

### Rodar os testes

```bash
docker compose --profile test up --build --abort-on-container-exit
```

## Endpoints e Exemplos de Uso

### Criar um leilao

```bash
curl -X POST http://localhost:8080/auction \
  -H "Content-Type: application/json" \
  -d '{
    "product_name": "iPhone 15",
    "category": "Tecnologia",
    "description": "iPhone 15 Pro Max 256GB novo na caixa",
    "condition": 1
  }'
```

### Listar leiloes ativos

```bash
curl http://localhost:8080/auction?status=0
```

### Buscar leilao por ID

```bash
curl http://localhost:8080/auction/{auctionId}
```

### Criar um lance (bid)

```bash
curl -X POST http://localhost:8080/bid \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": "550e8400-e29b-41d4-a716-446655440000",
    "auction_id": "{auctionId}",
    "amount": 5000.00
  }'
```

### Listar lances de um leilao

```bash
curl http://localhost:8080/bid/{auctionId}
```

### Buscar lance vencedor

```bash
curl http://localhost:8080/auction/winner/{auctionId}
```

### Buscar usuario por ID

```bash
curl http://localhost:8080/user/{userId}
```

## Verificando o Fechamento Automatico

1. Crie um leilao (com `AUCTION_DURATION=20s` no `.env`)
2. Aguarde 20 segundos
3. Consulte o leilao pelo ID — o status deve ter mudado para `1` (Completed)

```bash
# 1. Criar leilao
curl -s -X POST http://localhost:8080/auction \
  -H "Content-Type: application/json" \
  -d '{
    "product_name": "Test Auto Close",
    "category": "Test",
    "description": "Teste de fechamento automatico do leilao",
    "condition": 1
  }'

# 2. Listar leiloes ativos e pegar o ID
curl -s http://localhost:8080/auction?status=0

# 3. Aguardar AUCTION_DURATION (20s por padrao)
sleep 21

# 4. Verificar que o leilao foi fechado (status = 1)
curl -s http://localhost:8080/auction/{auctionId}
```

## Arquitetura

```
cmd/auction/main.go          -> Entry point + DI
internal/
  entity/                    -> Entidades de dominio
  infra/
    api/web/controller/      -> Controllers HTTP (Gin)
    database/
      auction/
        create_auction.go    -> Criacao + goroutine de fechamento automatico
        find_auction.go      -> Consultas de leilao
      bid/                   -> Criacao e consulta de bids
  usecase/                   -> Casos de uso
configuration/               -> Logger, DB connection, REST errors
```
