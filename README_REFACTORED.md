# NATS Client - Broker UI

Uma aplicação GUI para gerenciar conexões NATS, publicadores e subscritores.

## Nova Estrutura do Projeto

O projeto foi refatorado para separar claramente as responsabilidades:

```
/
├── main.go                           # Ponto de entrada da aplicação
├── internal/                         # Código interno da aplicação
│   ├── models/                       # Estruturas de dados
│   │   └── models.go                 # Server, Topic, Subscription, Message
│   ├── database/                     # Camada de acesso a dados
│   │   ├── database.go               # Configuração do banco
│   │   ├── server_repository.go     # Operações CRUD para servidores
│   │   ├── topic_repository.go      # Operações CRUD para tópicos
│   │   └── subscription_repository.go # Operações CRUD para subscrições
│   ├── services/                     # Lógica de negócio
│   │   ├── server_service.go        # Gerenciamento de servidores NATS
│   │   └── message_service.go       # Gerenciamento de mensagens
│   └── ui/                          # Interface de usuário
│       ├── components/              # Componentes reutilizáveis
│       │   ├── dialogs.go           # Diálogos padrão
│       │   └── menus.go             # Menus e botões
│       └── views/                   # Views principais
│           ├── main_window.go       # Janela principal
│           └── tab_manager.go       # Gerenciador de abas
├── natscli/                         # Cliente NATS (existente)
├── themes/                          # Temas da aplicação (existente)
└── icons/                           # Ícones da aplicação (existente)
```

## Benefícios da Refatoração

### 1. Separação de Responsabilidades
- **Models**: Estruturas de dados centralizadas
- **Database**: Operações de banco isoladas em repositories
- **Services**: Lógica de negócio separada da UI
- **UI**: Interface organizada em componentes e views

### 2. Manutenibilidade
- Código mais organizado e fácil de entender
- Funções menores e mais focadas
- Redução significativa no tamanho do `main.go` (de ~779 para ~32 linhas)

### 3. Testabilidade
- Services podem ser testados independentemente
- Repositories são testáveis com mocks
- UI separada da lógica de negócio

### 4. Reutilização
- Componentes de UI reutilizáveis
- Services podem ser compartilhados entre diferentes interfaces
- Repositories seguem padrões consistentes

## Principais Mudanças

### Antes (main.go monolítico)
- ~779 linhas de código
- UI, lógica e banco misturados
- Variáveis globais
- Difícil manutenção

### Depois (arquitetura em camadas)
- main.go com apenas ~32 linhas
- Separação clara de responsabilidades
- Injeção de dependências
- Código organizado e modular

## Como Usar

```bash
# Compilar
go build

# Executar
./broker-ui
```

## Funcionalidades

- ✅ Gerenciamento de servidores NATS
- ✅ Criação de publicadores (topics)
- ✅ Criação de subscritores
- ✅ Envio e recebimento de mensagens
- ✅ Dashboard de monitoramento
- ✅ Temas claro e escuro
- ✅ Interface gráfica intuitiva

## Tecnologias

- **Go**: Linguagem de programação
- **Fyne**: Framework para GUI
- **SQLite**: Banco de dados local
- **NATS**: Sistema de mensageria

---

Desenvolvido por [Alexandre E Souza](https://www.linkedin.com/in/devevantelista)