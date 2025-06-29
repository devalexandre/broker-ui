# NATS Client - Broker UI

![NATS Client](assets/screenshot.png)

A modern and intuitive desktop application for managing NATS connections, built with Go and Fyne. This tool allows you to connect to NATS servers, create publishers/subscribers, and monitor messages in real-time.

[Download](https://geoffrey-artefacts.fynelabs.com/github/devalexandre/devalexandre/broker-ui/825/index.html)

## 🚀 Features

### 📡 NATS Server Management
- **Add Servers**: Easily connect to multiple NATS servers
- **Edit Connections**: Modify existing server configurations
- **Persistence**: All configurations are saved in local SQLite database
- **Connection Validation**: Automatic connectivity testing when adding servers

### 📤 Publishers (Message Producers)
- **Dynamic Creation**: Add publishers for different topics
- **Customizable Subject**: Define specific subjects for each message
- **Intuitive Interface**: Multi-line text field for complex payloads
- **Message History**: View all sent messages
- **Real-time Sending**: Publish messages instantly

### 📥 Subscribers (Message Consumers)
- **Pattern Matching**: Support for wildcards (`*`, `>`) in subjects
- **Real-time Reception**: Messages appear instantly in the interface
- **Subject Tracking**: See which subject each message came from
- **Multiple Subscriptions**: Manage several subscriptions simultaneously
- **Flexible Patterns**: Configure patterns like `user.*`, `orders.>`, etc.

### 📊 Monitoring Dashboard
- **Real-time Metrics**: Message counter received per subscription
- **Automatic Updates**: Statistics updated every second
- **Overview**: Monitor all subscriptions in a single screen

### 🎨 Interface and Themes
- **Dracula Theme**: Modern and elegant dark theme
- **Light Theme**: Clean light theme for well-lit environments
- **Dynamic Toggle**: Switch between themes with one click
- **PNG Icons**: Enhanced visual interface with custom icons
- **Responsive Layout**: Adaptive and intuitive interface

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
- **NATS**: Distributed messaging system
- **SQLite**: Local database
- **Crypto/Ed25519**: Digital signature verification

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

### 1. Connect to a NATS Server
1. Click "Add Server"
2. Enter server name and URL (e.g., `nats://localhost:4222`)
3. Click "Confirm"
4. Select the server from the side list

### 2. Create a Publisher
1. With a connected server, click "Add Publisher"
2. Enter the topic name
3. Use the created tab to send messages
4. Customize the subject if needed

### 3. Create a Subscriber
1. Click "Add Subscriber"
2. Enter the subscription name
3. Configure the subject pattern (e.g., `user.*`, `orders.>`)
4. Messages will appear automatically in the tab

### 4. Monitor Activity
- Use the "Dashboard" tab to see statistics
- Each subscriber tab shows messages in real-time
- Publishers maintain a history of sent messages

## 🎨 Visual Resources

### Custom Icons
- **Add Server**: Server icon for adding connections
- **Theme Toggle**: Sun/moon icons that change according to theme
- **Exit**: Elegant exit icon
- **Publisher**: Specific icon for publishers
- **Subscriber**: Specific icon for subscribers

### Themes
- **Dracula**: Dark theme with vibrant colors
- **Light**: Clean and bright theme
- **Smart Toggle**: Button icon changes according to active theme

## 🏗️ Architecture

```
broker-ui/
├── main.go                 # Main application
├── icons/                  # Icon resources
│   ├── theme_toggle_resource.go
│   └── png/               # PNG icons
├── themes/                # Custom themes
│   ├── dracula/
│   └── light/
├── natscli/               # NATS client
└── README.md
```

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

- [NATS.io](https://nats.io/) - Amazing messaging system
- [Fyne](https://fyne.io/) - Go GUI framework
- [Dracula Theme](https://draculatheme.com/) - Inspiration for the dark theme

---

⭐ If this project was useful to you, consider giving it a star on GitHub!
