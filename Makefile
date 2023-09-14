compile:
	go mod tidy
	GOOS=linux GOARCH=amd64 go build -o cnexporter-linux-amd64 main.go