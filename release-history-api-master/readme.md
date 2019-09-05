# Release History API

## Purpose

This project allows to track deployments and releases of micro services:

- Deployment
    - Project: Project of the deployed service (e.g. alt)
    - Service: Service deployed (e.g. tracker-service)
    - Environment: Environment where the service was deployed (e.g. prod, staging)
    - Tag: Tag to identify the version deployed (e.g. git commit hash)
- Release
    - Project: Project of the deployed service (e.g. alt)
    - Number: Number to identify the release
    
When a release is created, all deployed services to the prod environment for the given project will be automatically associated with the new release.

## Development

### Prerequisites

- Golang 1.11+
- docker-compose

### Build

- `go build`

### Run

- Run app in command line:
    - `docker-compose up -d postgres`
    - `go run main.go` or `go build && ./release-history-api`
- Run app in docker: `docker-compose up --build -d`

## Configuration

Configuration is made using environment variables:
- `POSTGRES_CONNECTION_STRING`: full postgres connection string (default: `postgres://releasehistory:releasehistorylocal@localhost:5432/releasehistory?sslmode=disable`)
- `USERNAME`: username for basic auth (default: `local`)
- `PASSWORD`: password for basic auth (default: `local123`)

## API examples

### Deployment

#### Create new deployment

```
curl -u local:local123 -X POST \
  http://localhost:3000/deployment \
  -d '{
	"project": "alt",
	"service": "tracker-service",
	"tag": "d5d6ec591a9c7bb2fa8a6d3e033d05bdd7c1f8cc172",
	"environment": "staging"
    }'
```

#### List deployments

- Get current deployments for a given project and environment

```
     curl -u local:local123 -X GET \
     'http://localhost:3000/deployment?environment=staging&project=alt'
```

- Get all deployments for a given project and environment

```
     curl -u local:local123 -X GET \
     'http://localhost:3000/deployment?environment=staging&project=alt&showAll=true'
```

- Get deployments at a particular date for a given project and environment

```
     curl -u local:local123 -X GET \
     'http://localhost:3000/deployment?date=2019-04-10T13:07:35.905554Z&environment=staging&project=alt'
```

- Get deployment by ID

```
     curl -u local:local123 -X GET \
     'http://localhost:3000/deployment/123'
```

### Release

#### Create new release

```
curl -u local:local123 -X POST \
  http://localhost:3000/release \
  -d '{
	"project": "alt",
	"number": "v1.2.3"
    }'
```

#### List releases

- Get current release for a given project

```
     curl -u local:local123 -X GET \
     'http://localhost:3000/release?eproject=alt'
```

- Get all releases for a given project

```
     curl -u local:local123 -X GET \
     'http://localhost:3000/release?project=alt&showAll=true'
```

- Get release at a particular date for a given project

```
     curl -u local:local123 -X GET \
     'http://localhost:3000/release?date=2019-04-10T13:07:35.905554Z&project=alt'
```

- Get release by number and project

```
     curl -u local:local123 -X GET \
     'http://localhost:3000/release?number=v1.2.3&project=alt'
```

- Get release by ID

```
     curl -u local:local123 -X GET \
     'http://localhost:3000/release/123'
```
