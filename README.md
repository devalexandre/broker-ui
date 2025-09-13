# Universal Messaging Client - Broker UI

![Broker UI](assets/screenshot.png)

A modern, universal desktop application for managing multiple messaging systems, built with Go and Fyne. This tool provides a unified interface to connect to various message brokers including NATS, RabbitMQ, Kafka, and more.

[Download](https://geoffrey-artefacts.fynelabs.com/github/devalexandre/devalexandre/broker-ui/825/index.html)

## 🌟 New Architecture - Multi-Provider Support

The application has been completely refactored with a **pluggable messaging architecture** that supports multiple messaging systems through a unified interface.

### 🔌 Supported Messaging Providers

| Provider | Status | Features |
|----------|--------|----------|
| **NATS** | ✅ **Fully Implemented** | Wildcards (`*`, `>`), Real-time Pub/Sub |
| **RabbitMQ** | 🚧 **Structure Ready** | Exchanges, Routing Keys, Queues |
| **Kafka** | 📋 **Planned** | Topics, Partitions, Consumer Groups |
| **Redis** | 📋 **Planned** | Pub/Sub, Streams |
| **MQTT** | 📋 **Planned** | IoT Messaging |

## 🚀 Features

### 🌐 Universal Messaging Support
- **Multi-Provider**: Connect to different messaging systems simultaneously
- **Provider Selection**: Choose the appropriate provider for each server
- **Unified Interface**: Same UI for all messaging systems
- **Easy Migration**: Switch between providers seamlessly

### 📡 Server Management
- **Multiple Providers**: Each server can use a different messaging system
- **Provider Auto-Detection**: Intelligent provider selection
- **Connection Validation**: Test connectivity before saving
- **Persistence**: All configurations saved in local SQLite database

### 📤 Universal Publishers
- **Provider-Agnostic**: Same interface for all messaging systems
- **Smart Subject Handling**: Adapts to each provider's naming conventions
- **Message History**: Track sent messages across all providers
- **Real-time Publishing**: Instant message delivery

### 📥 Universal Subscribers
- **Pattern Support**: Wildcards, routing keys, topic patterns
- **Provider-Specific Patterns**: Optimized for each messaging system
- **Real-time Reception**: Instant message display
- **Cross-Provider Monitoring**: Monitor multiple systems simultaneously

### 📊 Advanced Monitoring Dashboard
- **Multi-Provider Metrics**: Statistics from all connected systems
- **Real-time Updates**: Live counters and status
- **Provider Identification**: See which system each message came from
- **Unified View**: Single dashboard for all messaging activity

### 🎨 Modern Interface
- **Dracula Theme**: Elegant dark theme for developers
- **Light Theme**: Clean interface for bright environments
- **Dynamic Theme Toggle**: Switch themes instantly
- **Provider Icons**: Visual identification of messaging systems
- **Responsive Design**: Adaptive layout for all screen sizes

### 💾 Data Persistence
- **SQLite Database**: Local storage of all configurations
- **Auto-migration**: Automatic database schema updates
- **Automatic Backup**: Data preserved between sessions

### 🔄 Automatic Updates
- **Self-Update**: Integrated automatic update system
- **Periodic Check**: Search for updates every 12 hours
- **Digital Signature**: Cryptographically verified updates

## 🛠️ Technologies Used

- **Go**: Main programming language
- **Fyne**: Cross-platform GUI framework
- **NATS**: Primary messaging system implementation
- **RabbitMQ**: Alternative messaging system (future support)
- **SQLite**: Local database with automatic migrations
- **Crypto/Ed25519**: Digital signature verification for updates
- **Clean Architecture**: Modular design with separation of concerns

## 📦 Installation

### Prerequisites
- Go 1.19 or higher
- Git

### Build
```bash
git clone https://github.com/devalexandre/broker-ui.git
cd broker-ui
go mod tidy
go build -o broker-ui
```

### Run
```bash
./broker-ui
```

## 🎯 How to Use

### 1. Connect to a Messaging Server
1. Click "Add Server"
2. Select the messaging provider (NATS, RabbitMQ, etc.)
3. Enter server name and connection URL
   - **NATS**: `nats://localhost:4222`
   - **RabbitMQ**: `amqp://localhost:5672` (coming soon)
4. Click "Confirm"
5. Select the server from the side list

### 2. Create a Publisher
1. With a connected server, click "Add Publisher"
2. Enter the topic name
3. Use the created tab to send messages
4. Customize the subject/routing key if needed

### 3. Create a Subscriber
1. Click "Add Subscriber"
2. Enter the subscription name
3. Configure the subject pattern
   - **NATS**: `user.*`, `orders.>`, etc.
   - **RabbitMQ**: Queue patterns (coming soon)
4. Messages will appear automatically in the tab

### 4. Monitor Activity
- Use the "Dashboard" tab to see unified statistics across all providers
- Each subscriber tab shows messages in real-time
- Publishers maintain a history of sent messages
- Provider identification shows which system each message came from

## 🎨 Visual Resources

### Custom Icons
- **Add Server**: Server icon for adding connections
- **Theme Toggle**: Sun/moon icons that change according to theme
- **Exit**: Elegant exit icon
- **Publisher**: Specific icon for publishers
- **Subscriber**: Specific icon for subscribers
- **Provider Icons**: Visual identification for different messaging systems

### Themes
- **Dracula**: Dark theme with vibrant colors optimized for developers
- **Light**: Clean and bright theme for all environments
- **Smart Toggle**: Button icon changes according to active theme

## 🏗️ Architecture

The project follows clean architecture principles with clear separation of concerns:

```
broker-ui/
├── main.go                    # Application entry point
├── internal/                  # Core application logic
│   ├── models/               # Data structures
│   │   └── models.go
│   ├── database/             # Data access layer
│   │   ├── database.go
│   │   ├── server_repository.go
│   │   ├── topic_repository.go
│   │   └── subscription_repository.go
│   ├── services/             # Business logic layer
│   │   ├── server_service.go
│   │   └── message_service.go
│   ├── messaging/            # Messaging abstraction
│   │   ├── interfaces.go
│   │   └── providers/
│   │       ├── factory.go
│   │       ├── nats.go
│   │       └── rabbitmq.go
│   └── ui/                   # User interface layer
│       ├── components/
│       │   ├── dialogs.go
│       │   └── menus.go
│       └── views/
│           ├── main_window.go
│           └── tab_manager.go
├── icons/                     # Icon resources
│   ├── theme_toggle_resource.go
│   └── png/                  # PNG icons
├── themes/                    # Custom themes
│   ├── dracula/
│   └── light/
├── natscli/                   # Legacy NATS utilities
└── tests/                     # Test utilities
    ├── sender/
    └── sub/
```

### Architecture Layers
- **Models**: Define data structures used throughout the application
- **Database**: Repository pattern for data persistence with SQLite
- **Services**: Business logic that coordinates between data and UI layers
- **Messaging**: Provider abstraction allowing multiple messaging systems
- **UI**: Fyne-based user interface components and views

## 🤝 Contributing

Contributions are welcome! Please:

1. Fork the project
2. Create a feature branch (`git checkout -b feature/AmazingFeature`)
3. Commit your changes (`git commit -m 'Add some AmazingFeature'`)
4. Push to the branch (`git push origin feature/AmazingFeature`)
5. Open a Pull Request

## 📝 License

This project is under the MIT license. See the `LICENSE` file for more details.

## 👨‍💻 Author

**Alexandre E Souza**
- LinkedIn: [devevantelista](https://www.linkedin.com/in/devevantelista)
- GitHub: [@devalexandre](https://github.com/devalexandre)

## 🙏 Acknowledgments

- [NATS.io](https://nats.io/) - Primary messaging system implementation
- [RabbitMQ](https://www.rabbitmq.com/) - Alternative messaging system support
- [Fyne](https://fyne.io/) - Excellent Go GUI framework
- [Dracula Theme](https://draculatheme.com/) - Inspiration for the dark theme
- [SQLite](https://www.sqlite.org/) - Reliable embedded database

---

⭐ If this project was useful to you, consider giving it a star on GitHub!
