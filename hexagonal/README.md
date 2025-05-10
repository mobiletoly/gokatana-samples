# GoKatana REST API service based on hexagonal architecture

**hexagonal** is a template to generate a new GoKatana-based web service (REST endpoint)
It is based on Hexagonal architecture pattern and provides you with a good starting point
to build your own service.

What will you get:

- REST endpoint to add, get and list contacts
- Support of generating REST model types from swagger spec
- PostgreSQL database connection and schema migrations
- Configuration management
- Error handling
- Logging
- Integration tests based on TestContainers
- Contains sample Docker file
- Contains shell script to create your own service based on this template

First, clone the repository:

```shell
git clone https://github.com/mobiletoly/gokatana-samples.git
```

Now do `cd hexagonal` to enter the Hexagonal template directory and you have few options:

**You can run hexagonal sample directly**

or

**You can create a new service based on hexagonal template**


## (Option 1) Run hexagonal sample directly

To run hexagonal sample directly, you need to have Go 1.24 or later installed.
Before running the project first time, make sure to do

```shell
go mod tidy
```

to pick up dependencies and install few tools used by the project. By tools we mean:

- `swagger` to generate REST models from swagger spec:
  (you can read more here: https://github.com/go-swagger/go-swagger)

- `gobetter` to generate model helpers for Builder pattern
  (you can read more here: https://github.com/mobiletoly/gobetter)

we use new `tool` directive (available in go 1.24+) added to `go.mod` file to install these tools automatically.

Running `go mod tidy` will take care of installing required tools.

Then, you can run the service using:

```shell
export HEXAGONAL_DATABASE_USER=postgres
export HEXAGONAL_DATABASE_PASSWORD=postgres
go run main.go run --deployment=local
```

It will start the service, and you can access it at http://localhost:8080/api/v1/sample/version

Note that Hexagonal service provides example of using multiple API server implementations:
- `internal/adapters/apiserver_echo` package - implementation based on Echo framework
- `internal/adapters/apiserver_std` package - implementation based on standard net/http library (ServeMux)
- `internal/adapters/apiserver_chi` package - implementation based on Chi router

You can switch between them by changing opening file `internal/infra/launcher.go` and
commenting/uncommenting appropriate lines:

```go
// Option 1: Echo framework
server := apiserver_echo.Start(ctx, uc)

// Option 2: Standard net/http
//server := apiserver_std.Start(ctx, uc)

// Option 3: Chi router
//server := apiserver_chi.Start(ctx, uc)
```


## (Option 2) Create your own new service based on hexagonal template

To create a new web service based on hexagonal template, run the `create-new-service.sh` script
provided in hexagonal sample directory and pass 3 parameters (service name, package name and
output directory), e.g.:

```shell
./create-new-service.sh myservice github.com/somehandler/myservice ../../myservice
```

(replace **myservice** with the name of your new service, do not forget to replace the package name as well)

This will create a new directory **myservice** two levels up from the current directory with Hexagonal sample.
All files will be adjusted to the new service name and new package name.

If you need to change config values via environment variables, you must use `MYSERVICE_` prefix
(in case of **myservice** service) to set the values. For example, to set database username and password:

```shell
export MYSERVICE_DATABASE_USER=postgres
export MYSERVICE_DATABASE_PASSWORD=postgres
go run main.go run --deployment=local
```

The service created contains multiple similar implementations for API server adapter,
such as `apiserver_std`, `apiserver_chi`, `apiserver_echo`. You can
remove the ones you do not need.

# Swagger

REST models reside in `internal/core/model` directory, but you should not be editing it directly.
Instead, you should edit `swagger/contact.yaml` file and run `go generate` command to generate
REST models from swagger spec.


# Integration tests

Application contains set of integrations tests that can be run using `go test -v ./...` command.
It uses TestContainers framework to start PostgreSQL container and run REST API tests against it.
You need to have Docker environment installed and running to run these tests.

# Structure of the project

```
myservice/
├── configs/                  # Configuration files
│   ├── common.yaml           # Common configuration
│   ├── local.yaml            # Local development configuration
│   ├── dev.yaml              # Development environment configuration
│   └── prod.yaml             # Production environment configuration
├── dbmigrate/                # Database migration files
├── internal/
│   ├── adapters/             # Adapters (HTTP handlers, database repositories)
│   │   ├── apiserver_chi/    # (1) HTTP API handlers implemented with net/http framework
│   │   │   ├── apiserver.go  # HTTP API server launcher (and route registration)
│   │   │   └── contact.go    # HTTP API /contact route handlers
│   │   ├── apiserver_echo/   # (2) HTTP API handlers implemented with Echo framework
│   │   │   ├── apiserver.go  # ..
│   │   │   ├── contact.go    # ..
│   │   ├── apiserver_std/    # (3) HTTP API handlers implemented with net/http framework
│   │   │   ├── apiserver.go  # ...
│   │   │   └── contact.go    # ...
│   │   └── persist/          # Database repositories
│   │       ├── contact.go    # Contact repository adapter (contact table) to be called from business logic
│   │       └── internal/     # (internal implementation details)
│   │           ├── mapper/   # Mapper of database entities to business logic models
│   │           └── repo/     # Database specific functionality (SQL queries)
│   ├── core/                 # Core business logic
│   │   ├── app/              # Application configuration
│   │   ├── model/            # Domain/business models (our simple app uses REST models as business models)
│   │   └── usecase/          # Use cases (business logic)
│   └── infra/                # Infrastructure setup (DI, loads config, launches API server adapter etc)
├── intgr_test/               # Integration tests
├── main.go                   # Application entry point
└── go.mod                    # Go module definition
```
