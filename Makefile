compile:
	go mod tidy
	GOOS=linux GOARCH=amd64 go build -v -o cnexporter-linux-amd64 main.go