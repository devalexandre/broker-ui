# Abstração de Messagerias - Broker UI

## Nova Arquitetura Multi-Provider

O projeto foi refatorado para suportar múltiplas messagerias através de uma interface comum, permitindo fácil adição de novos providers como RabbitMQ, Kafka, Redis, etc.

## Estrutura da Abstração

```
internal/messaging/
├── interfaces.go              # Interface comum MessagingProvider
├── providers/
│   ├── factory.go            # Factory para criar providers
│   ├── nats.go               # Implementação NATS
│   └── rabbitmq.go           # Implementação RabbitMQ (exemplo)
```

## Interface MessagingProvider

```go
type MessagingProvider interface {
    Connect(url string) error
    Publish(subject string, data []byte) error
    Subscribe(subjectPattern string, handler MessageHandler) error
    Unsubscribe(subjectPattern string) error
    Close() error
    IsConnected() bool
    GetProviderType() ProviderType
}
```

## Providers Suportados

### ✅ NATS (Implementado)
- **Conexão**: `nats://localhost:4222`
- **Padrões**: Suporta wildcards (`*`, `>`)
- **Recursos**: Pub/Sub em tempo real

### 🚧 RabbitMQ (Placeholder)
- **Conexão**: `amqp://localhost:5672`
- **Padrões**: Exchange/Routing Keys
- **Recursos**: Filas, Exchange patterns

### 📋 Futuros Providers
- **Kafka**: Topics e Partitions
- **Redis**: Pub/Sub e Streams
- **Apache Pulsar**: Topics distribuídos
- **MQTT**: IoT messaging

## Como Adicionar um Novo Provider

### 1. Implementar a Interface

```go
// internal/messaging/providers/meu_provider.go
type MeuProvider struct {
    // campos necessários
}

func (p *MeuProvider) Connect(url string) error {
    // implementar conexão
}

func (p *MeuProvider) Publish(subject string, data []byte) error {
    // implementar publicação
}

// ... outros métodos da interface
```

### 2. Registrar no Factory

```go
// internal/messaging/providers/factory.go
func (f *Factory) CreateProvider(providerType messaging.ProviderType) (messaging.MessagingProvider, error) {
    switch providerType {
    case messaging.ProviderNATS:
        return NewNATSProvider(), nil
    case messaging.ProviderMeuProvider:
        return NewMeuProvider(), nil
    // ...
    }
}
```

### 3. Adicionar à UI

```go
// Adicionar ao dropdown de seleção
providerSelect := widget.NewSelect([]string{"NATS", "MeuProvider"}, func(value string) {})
```

## Benefícios da Abstração

### 🔌 **Plugabilidade**
- Fácil adição de novos providers
- Interface consistente para todos os tipos
- Sem mudanças na lógica da aplicação

### 🔄 **Flexibilidade**
- Usuário pode escolher o provider por servidor
- Suporte simultâneo a múltiplas messagerias
- Migração transparente entre providers

### 🧪 **Testabilidade**
- Mocks implementam a mesma interface
- Testes isolados por provider
- Validação de comportamento consistente

### 📈 **Escalabilidade**
- Adição de providers sem breaking changes
- Configuração dinâmica de providers
- Suporte a diferentes casos de uso

## Configuração no Banco

A tabela `servers` foi atualizada para incluir o tipo de provider:

```sql
CREATE TABLE servers (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT,
    url TEXT,
    provider_type TEXT DEFAULT 'NATS'
);
```

## Migração de Dados

Servidores existentes são automaticamente marcados como NATS:

```go
// Migration automática
_, err = db.Exec(`ALTER TABLE servers ADD COLUMN provider_type TEXT DEFAULT 'NATS'`)
```

## Exemplo de Uso

```go
// Criar factory
factory := providers.NewFactory()

// Criar provider NATS
natsProvider, err := factory.CreateProvider(messaging.ProviderNATS)
if err != nil {
    log.Fatal(err)
}

// Conectar
err = natsProvider.Connect("nats://localhost:4222")
if err != nil {
    log.Fatal(err)
}

// Publicar mensagem
err = natsProvider.Publish("test.subject", []byte("Hello World"))
if err != nil {
    log.Fatal(err)
}

// Subscrever
err = natsProvider.Subscribe("test.*", func(subject string, data []byte) {
    fmt.Printf("Received: %s on %s\n", string(data), subject)
})
```

## Próximos Passos

1. **Implementar RabbitMQ Provider completo**
   - Adicionar dependência `github.com/streadway/amqp`
   - Implementar conexão real
   - Suporte a exchanges e filas

2. **Adicionar Kafka Provider**
   - Usar `github.com/Shopify/sarama`
   - Suporte a topics e partitions
   - Consumer groups

3. **Implementar Redis Provider**
   - Usar `github.com/go-redis/redis`
   - Pub/Sub e Streams
   - Padrões de subscribe

4. **Configurações Avançadas**
   - SSL/TLS por provider
   - Autenticação específica
   - Parâmetros customizados

---

Esta abstração torna o Broker UI uma ferramenta verdadeiramente universal para messagerias! 🚀