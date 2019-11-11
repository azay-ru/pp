set now=%date%
go build -ldflags "-X main.version=%now%"
