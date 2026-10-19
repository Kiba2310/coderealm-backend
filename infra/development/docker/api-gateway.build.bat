set CGO_ENABLED=0
set GOOS=windows
set GOARCH=amd64
go build -o build/api-gateway ./services/api-gateway
