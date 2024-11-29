# TempFile.Link Backend

## Commands

### Development

1. Install the dependecise:

```shell
go mod download
```

2. Run the server:

```shell
go run main.go
```

or

```shell
go install github.com/air-verse/air@latest
air
```

### Production

#### Linux

1. Build the application

```shell
go build -o server main.go
```

2. Run the application

```shell
./server
```

#### Windows

1. Build the application

```shell
go build -o server.exe main.go
```

2. Run the application

```shell
.\server.exe
```

### Docker

1. Build the image

```shell
docker build -t tempfile . 
```

2. Run the image

```shell
docker run --name tempfile -p 3001:3001 --env-file .env -e DB_HOST=host.docker.internal tempfile
```
