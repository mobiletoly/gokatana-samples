## About

GoKatana based application providing authentication and user management REST API service
as well as web interface to manage users and roles.

Tech stack:

- GoKatana libraries to simplify building web service
- PostgreSQL database access
- Swagger for API documentation and to generate REST models
- Web UI based on HTMX with Tailwind CSS for styling
- Docker for local development
- TestContainers for integration tests
- Hexagonal architecture pattern for better code organization

## Hot reload while developing and changing template files

1. install `air` (https://github.com/air-verse/air)
2. launch:
   ```shell
   IAMSERVICE_DATABASE_PASSWORD=postgres IAMSERVICE_DATABASE_USER=postgres air run --deployment=local
    ```
