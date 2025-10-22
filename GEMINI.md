# GEMINI.md

## Project Overview

This project is a Go-based REST API that provides auto-suggestions for UK place names. It is designed for efficient prefix-based searching, leveraging a trie data structure for fast lookups.

The core of the application is a web server built with the [Gin framework](https://gin-gonic.com/). It exposes an endpoint that, given a prefix, returns a list of matching place names, sorted by relevancy.

The dataset of place names and their relevancy scores is loaded from a gzipped CSV file (`./data/placenames_with_relevancy.csv.gz`) into the in-memory trie structure upon server startup.

The application is containerized using Docker and includes a `docker-compose` setup for easy initialization of the data. It also integrates several observability and debugging features, including:
- Prometheus metrics exposed at `/metrics`.
- Health check endpoint at `/healthz`.
- Optional `pprof` profiling endpoints for debugging.

## Building and Running

### Prerequisites

- Go (version 1.25 or later)
- Docker and Docker Compose

### Data Initialization

The initial data setup requires running a Docker Compose service to download and prepare the placename data.

```bash
docker-compose up init_db
```

This command will populate the `./data` directory with the necessary database files.

### Running the Application

There are two primary ways to run the application:

**1. Using Go:**

To run the API server directly using Go:

```bash
go run main.go api-server
```

By default, the server will start on port `8080`. You can customize the port and data file path using flags:

```bash
go run main.go api-server --port 8080 --file ./data/placenames_with_relevancy.csv.gz
```

**2. Using Docker:**

The project includes a `Dockerfile` for building a container image. The `.github/workflows/build.yml` workflow demonstrates how to build and publish the image.

To build and run the Docker container locally:

```bash
docker build -t placenames-api .
docker run -p 8080:8080 placenames-api
```

### API Usage

The main endpoint for auto-suggestions is:

```
GET /v1/place-names/prefix/:query
```

- `:query`: The prefix to search for.
- `max_results` (optional query parameter): The maximum number of results to return (default: 10, max: 100).

Example requests can be found in the `test.http` file.

## Development Conventions

### Testing

The project uses `gotestsum` for running tests and generating reports. The test suite can be executed with:

```bash
go install gotest.tools/gotestsum@latest
gotestsum -- -v ./...
```

The CI pipeline in `.github/workflows/build.yml` also runs the tests and uploads coverage reports to Coveralls.

### Linting

Code quality is maintained using `golangci-lint`. To run the linter:

```bash
golangci-lint run
```

### Dependencies

Go modules are used for dependency management. The dependencies are defined in `go.mod` and `go.sum`.

### Code Structure

- `main.go`: The main entry point for the application, defining the `api-server` command.
- `cmd/api_server.go`: Contains the setup for the Gin web server, including middleware and routes.
- `internal/`: Contains the core application logic.
  - `trie.go`: Implementation of the trie data structure and data loading logic.
  - `routes/prefix.go`: The handler for the prefix search API endpoint.
- `init/`: Contains the Docker setup for data initialization.
- `Dockerfile`: Defines the Docker image for the application.
- `docker-compose.yml`: Defines the services for running the application and its dependencies.
