# Introduction

                ,_---~~~~~----._
           _,,_,*^____      _____``*g*\"*,
          / __/ /'     ^.  /      \ ^@q   f
         [  @f | @))    |  | @))   l  0 _/
          \`/   \~____ / __ \_____/    \
           |           _l__l_           I
           }          [______]           I
           ]            | | |            |
           ]             ~ ~             |
           |                            |
            |                           |

# GoLang GraphQL API Boilerplate

This repository serves as a boilerplate codebase for API development at Beyul Labs. It is built using Go and follows Clean Code Architecture principles.

## Features

### **GraphQL API**

- Built using `gqlgen`.
- Supports Relay-based pagination.
- Sample Todo API for demonstration.

### **Authentication**

- JWT-based authentication.
- Refresh token support.

### **Database**

- PostgreSQL for data storage.
- Ent as the ORM for database operations.

### **Environment Management**

- Uses Viper to load environment variables from `.env`.(Todo)

### **Development Tools**

- Air for hot-reloading during development.
- Docker Compose for local development.

### **Testing**

- Unit tests, repository tests, and end-to-end (E2E) tests.

### **Code Generation**

- Ent schema generation.
- Repository mock generation.
- GraphQL code generation using gqlgen.

## Prerequisites

Before you begin, ensure you have the following installed:

- Docker and Docker Compose.
- Go (version 1.20 or higher).
- Make (for running Makefile commands).

## Project Structure

```
.
├── cmd/                  # Application entry points
│   ├── app/              # Main application
│   └── migration/        # Database migrations
├── ent/                  # Ent schema and generated code
├── pkg/                  # Shared packages
│   ├── adapter/          # Adapters (e.g., repository implementations)
│   ├── usecase/          # Use cases
│   └── utils/            # Utility functions
├── test/                 # Tests
│   ├── e2e/              # End-to-end tests
│   └── unit/             # Unit tests
├── docker/               # Docker-related files
│   └── Dockerfile
├── scripts/              # Helper scripts
├── .env                  # Environment variables
├── .air.toml             # Air configuration for hot-reloading
├── Makefile              # Makefile for common tasks
├── docker-compose.yml    # Docker Compose configuration
├── go.mod                # Go module file
└── README.md             # This file
```

## Getting Started

### 1. Clone the Repository

```sh
git clone https://github.com/beyul/golang-boilerplate.git
cd golang-graphql-boilerplate
```

### 2. Set Up Environment Variables

Currently it loads from confi.yml but need to implement .env support to load the config.yml using envsubst

### 3. Start the Services

Use Docker Compose to start the MySQL database and the Go application:

```sh
docker-compose up --build
```

This will:

- Start the PostgreSQL database.
- Build and run the Go application with hot-reloading using Air.

### 4. Set Up the Database

```sh
make setup_db
```

### 5. Run Database Migrations

```sh
make migrate_schema
```

### 6. Access the Application

- **GraphQL Playground**: Open [http://localhost:8080](http://localhost:8080) in your browser.
- **PostgreSQL Database**: Accessible at `localhost:5433`.

## Development

### Hot-Reloading with Air

The application uses Air for hot-reloading during development. Any changes to the code will automatically restart the server.

### Running Tests

- **Unit Tests:**
  ```sh
  make test_unit
  ```

### Set Up the Database

```sh
make setup_db_test
```

- **Repository Tests:**
  ```sh
  make test_repository
  ```
- **End-to-End (E2E) Tests:**
  ```sh
  make test_e2e
  ```

### Code Generation

- **Generate Ent Code:**
  ```sh
  make generate_ent
  ```
- **Generate Repository Mocks:**
  ```sh
  make generate_repo_mocks
  ```
- **Generate GraphQL Code:**
  ```sh
  make gqlgen
  ```

## Docker Setup

### Docker Compose Services

The `docker-compose.yml` file defines two services:

- **PostgreSQL**: The database service.
- **App**: The Go application service.

### Build and Run

```sh
docker-compose up --build
```

### Stop Services

```sh
docker-compose down
```

## Clean Code Architecture

The project follows Clean Code Architecture principles:

- **Core**: Contains the business logic.
- **Handlers**: Manages GraphQL requests.
- **Repositories**: Handles database operations.
- **Services**: Implements business logic services.
- **Entities**: Defines the database schema using Ent.

## Technologies Used

- **Go**: The primary programming language.
- **GraphQL**: API layer using gqlgen.
- **Ent**: ORM for database operations.
- **PostgreSQL**: Relational database.
- **Viper**: Environment variable management.
- **Air**: Hot-reloading for development.
- **Docker**: Containerization for local development.

## Contributing

Contributions are welcome! Please follow these steps:

1. Fork the repository.
2. Create a new branch:
   ```sh
   git checkout -b feature/your-feature
   ```
3. Commit your changes:
   ```sh
   git commit -m 'Add some feature'
   ```
4. Push to the branch:
   ```sh
   git push origin feature/your-feature
   ```
5. Open a pull request.

## Support

If you encounter any issues or have questions, feel free to open an issue on GitHub.

**Happy coding! 🚀**
