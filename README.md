# Nats-UI

## Environment Variables for Secure Configuration

This project uses environment variables to manage sensitive information securely. Before running the application, ensure you have set up the following environment variables:

### NATS Configuration
- `NATS_AUTH_TOKEN`: Authentication token for NATS server

### Kafka Configuration
- `KAFKA_ADMIN_USERNAME`: Username for Kafka admin
- `KAFKA_ADMIN_PASSWORD`: Password for Kafka admin
- `KAFKA_CLIENT_USERNAME`: Username for Kafka client
- `KAFKA_CLIENT_PASSWORD`: Password for Kafka client

## Setting Up Environment Variables

### For Local Development
1. Create a `.env` file in the root directory of the project.
2. Add the environment variables to the `.env` file:

```
NATS_AUTH_TOKEN=your_nats_auth_token
KAFKA_ADMIN_USERNAME=your_kafka_admin_username
KAFKA_ADMIN_PASSWORD=your_kafka_admin_password
KAFKA_CLIENT_USERNAME=your_kafka_client_username
KAFKA_CLIENT_PASSWORD=your_kafka_client_password
```

3. Use a tool like `dotenv` to load these variables into your development environment.

### For Production Deployment
- Set these environment variables securely in your production environment.
- Avoid committing the `.env` file or any file containing these secrets to version control.

## Security Best Practices
- Regularly rotate passwords and tokens.
- Use strong, unique passwords for each service.
- Limit access to these environment variables to only those who need them.
- Consider using a secrets management service for production environments.

## Running the Application
Ensure all environment variables are set before starting the application. The application will use these variables to securely connect to NATS and Kafka services.

For more information on running the application, refer to the installation and usage sections below.

[Add your existing README content here, such as installation instructions, usage guide, etc.]
