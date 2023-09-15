# PREFIX is environment variable, but if it is not set, then set default value
ifeq ($(PREFIX),)
    PREFIX := /usr/bin
endif

compile:
	@printf "> "
	go mod tidy
	@printf "> "
	GOOS=linux GOARCH=amd64 go build -v -o cnexporter-linux-amd64 main.go

install: compile
	@printf "> "
	sudo install -d $(PREFIX)
	@printf "> "
	sudo install -m 770 cnexporter-linux-amd64 $(PREFIX)
	@printf "> "
	sudo ln --force --verbose --symbolic $(PREFIX)/cnexporter-linux-amd64 $(PREFIX)/cnexporter
