set now=%date%
go build -ldflags "-X github.com/azay-ru/pp/app.version=%now%"
