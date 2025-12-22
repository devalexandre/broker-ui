# Universal Messaging Client - Broker UI

![Broker UI](assets/screenshot.png)

A modern, universal desktop application for managing multiple messaging systems, built with Go and Fyne. This tool provides a unified interface to connect to various message brokers including NATS, RabbitMQ, Google Cloud Pub/Sub, and more.

[Download](https://geoffrey-artefacts.fynelabs.com/github/devalexandre/devalexandre/broker-ui/825/index.html)

## 🌟 New Architecture - Multi-Provider Support

The application has been completely refactored with a **pluggable messaging architecture** that supports multiple messaging systems through a unified interface.

### 🔌 Supported Messaging Providers

| Provider | Status | Features |
|----------|--------|----------|
| **NATS** | ✅ **Fully Implemented** | Wildcards (`*`, `>`), Real-time Pub/Sub, Authentication |
| **RabbitMQ** | ✅ **Fully Implemented** | Exchanges, Routing Keys, Queues, AMQP 0.9.1 |
| **Google Cloud Pub/Sub** | ✅ **Fully Implemented** | Topics, Subscriptions, Emulator Support, GCP Production |
| **Kafka** | ✅ **Implemented (basic)** | Publish/Subscribe via consumer group |
| **Redis** | ✅ **Implemented (Pub/Sub)** | Pub/Sub (PUBLISH/SUBSCRIBE/PSUBSCRIBE) |
| **MQTT** | ✅ **Implemented (QoS 0)** | Publish/Subscribe, mqtt:// and mqtts:// |

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
- **Audit (Received Messages)**: History view with filter/pagination and cleanup

### 🔄 Automatic Updates
- **Self-Update**: Integrated automatic update system
- **Periodic Check**: Search for updates every 12 hours
- **Digital Signature**: Cryptographically verified updates

### 🗑️ Enhanced UI Features
- **Server Management**: Add, edit, and delete server connections with confirmation dialogs
- **Tab Auto-Refresh**: Publisher/Subscriber tabs load automatically when created
- **Custom Icons**: Trash bin icons for delete operations, provider-specific icons
- **Responsive Layout**: Adjustable panel widths and responsive design
- **Error Handling**: Comprehensive error dialogs and validation

## 🆕 Recent Updates

### Version 3.0 - Multi-Provider Architecture
- ✅ **Google Cloud Pub/Sub Support**: Full implementation with emulator and production support
- ✅ **Provider Auto-Detection**: Intelligent protocol detection from URLs
- ✅ **Enhanced UI**: Automatic tab refresh, delete confirmations, improved layouts
- ✅ **Bug Fixes**: Fixed tab loading issues, improved error handling
- ✅ **GitHub Actions**: Updated CI/CD pipeline with latest artifact actions

### Key Improvements
- **Real-time Tab Updates**: No more manual refresh needed when adding publishers/subscribers
- **Smart URL Parsing**: Automatic provider detection for seamless connections
- **Enhanced Server Management**: Full CRUD operations with user-friendly confirmations
- **Production-Ready**: Support for both local development and cloud production environments

## 🛠️ Technologies Used

- **Go**: Main programming language with latest Go 1.24 support
- **Fyne**: Cross-platform GUI framework for modern interfaces
- **NATS**: High-performance messaging system with wildcards and real-time pub/sub
- **RabbitMQ**: Robust AMQP 0.9.1 implementation with exchanges and routing
- **Google Cloud Pub/Sub**: Enterprise-grade messaging with emulator and production support
- **SQLite**: Local database with automatic migrations and schema updates
- **Crypto/Ed25519**: Digital signature verification for secure updates
- **Clean Architecture**: Modular design with separation of concerns and provider abstraction

## 📦 Installation

### Prerequisites
- Go 1.19 or higher
- Git
- Docker and Docker Compose (optional, for testing with local messaging servers)

### Quick Start with Docker
```bash
git clone https://github.com/devalexandre/broker-ui.git
cd broker-ui

# Start messaging servers (NATS, RabbitMQ, Pub/Sub Emulator, Redis, MQTT, Kafka)
docker-compose up -d

# Build and run the application
go mod tidy
go build -o broker-ui
./broker-ui
```

### Build Only
```bash
git clone https://github.com/devalexandre/broker-ui.git
cd broker-ui
go mod tidy
go build -o broker-ui
```

### Connection Examples
**Simple URLs (protocol added automatically):**
- **NATS**: `localhost:4222`
- **RabbitMQ**: `admin:admin123@localhost:5672/`
- **Google Cloud Pub/Sub (Emulator)**: `localhost:8085`
- **Google Cloud Pub/Sub (Production)**: `my-project-id` or `gcp://my-project-id`
- **Redis**: `localhost:6379`
- **MQTT**: `localhost:1883`
- **Kafka**: `localhost:19092`

**Full URLs (also supported):**
- **NATS**: `nats://localhost:4222`
- **RabbitMQ**: `amqp://admin:admin123@localhost:5672/`
- **Pub/Sub**: Auto-detected based on URL pattern
- **Redis**: `redis://localhost:6379/0`
- **MQTT**: `mqtt://localhost:1883?client_id=broker-ui&keepalive=30`
- **Kafka**: `kafka://localhost:19092?group=broker-ui&start=latest`

### Management Interfaces
- **NATS Monitoring**: http://localhost:8222
- **RabbitMQ Management**: http://localhost:15672 (admin/admin123)
- **Pub/Sub Emulator**: Runs on localhost:8085 (no web interface)
- **Redpanda (Kafka) Admin API**: http://localhost:9644

### Google Cloud Pub/Sub Setup

#### Local Development (Emulator)
The docker-compose includes a Pub/Sub emulator that runs locally:
```bash
# Start the emulator
docker-compose up pubsub-emulator -d

# Connect using: localhost:8085
```

#### Production (Google Cloud)
For production GCP connections:
1. **Service Account**: Set up a service account with Pub/Sub permissions
2. **Authentication**: Use one of these methods:
   ```bash
   # Option 1: Service Account Key
   export GOOGLE_APPLICATION_CREDENTIALS="/path/to/service-account.json"
   
   # Option 2: gcloud CLI
   gcloud auth application-default login
   ```
3. **Connection**: Use your project ID as the URL: `my-production-project`

## 🎯 How to Use

### 1. Connect to a Messaging Server
1. Click "Add Server"
2. Select the messaging provider (NATS, RabbitMQ, PUBSUB, Redis, MQTT, Kafka)
3. Enter server name and connection URL:
   - **NATS**: `localhost:4222` or `nats://localhost:4222`
   - **RabbitMQ**: `admin:admin123@localhost:5672/` or `amqp://admin:admin123@localhost:5672/`
   - **Pub/Sub Emulator**: `localhost:8085`
   - **Pub/Sub Production**: `my-project-id`
4. Click "Confirm" - the provider is auto-detected from the URL
5. Select the server from the side list to connect

### 2. Create a Publisher (Topic)
1. With a connected server, click "Add Topic" from the server menu
2. Enter the topic name
3. **The tab appears automatically** with the publisher interface
4. Use the tab to send messages with custom subjects/routing keys

### 3. Create a Subscriber
1. Click "Add Subscription" from the server menu
2. Enter the subscription name and subject pattern:
   - **NATS**: `user.*`, `orders.>`, `specific.subject`
   - **RabbitMQ**: Queue names like `user_events`, `order_processing`
   - **Pub/Sub**: Topic names like `user-events`, `order-processing`
3. **The tab appears automatically** with the subscriber interface
4. Messages appear in real-time as they arrive

### 4. Server Management
- **Edit Servers**: Click the edit button in any server tab to modify connection details
- **Delete Servers**: Click the trash icon next to any server in the list (with confirmation)
- **Multiple Providers**: Connect to different messaging systems simultaneously
- **Auto-Detection**: Provider type is automatically detected from URL patterns

### 5. Monitor Activity
- Use the "Dashboard" tab to see unified statistics across all providers
- Each subscriber tab shows messages in real-time with provider identification
- Publishers maintain a history of sent messages
- Cross-provider monitoring shows activity from all connected systems

## 🎨 Visual Resources

### Custom Icons
- **Add Server**: Server icon for adding connections
- **Theme Toggle**: Sun/moon icons that change according to theme
- **Exit**: Elegant exit icon
- **Publisher**: Specific icon for publishers
- **Subscriber**: Specific icon for subscribers
- **Delete Operations**: Custom trash bin icons for server deletion
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
│   │       ├── rabbitmq.go
│   │       └── pubsub.go
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

- [NATS.io](https://nats.io/) - High-performance messaging system implementation
- [RabbitMQ](https://www.rabbitmq.com/) - Robust AMQP messaging platform
- [Google Cloud Pub/Sub](https://cloud.google.com/pubsub) - Enterprise messaging service
- [Fyne](https://fyne.io/) - Excellent Go GUI framework
- [Dracula Theme](https://draculatheme.com/) - Inspiration for the dark theme
- [SQLite](https://www.sqlite.org/) - Reliable embedded database

---

⭐ If this project was useful to you, consider giving it a star on GitHub!
