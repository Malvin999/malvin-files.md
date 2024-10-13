bot:
	go run ./cmd/bot

gui:
	cd ./cmd/gui && wails dev

test:
	go test ./...

install:
	go mod tidy

check:
	go fmt ./... && go vet ./... && go test ./...

init_server:
	ssh $(host) "\
		mkdir -p /app/storage && \
		mkdir -p /var/log/files.md && \
		mkdir -p /opt/files.md && \
		chown -R www-data:www-data /app && \
		chown -R www-data:www-data /var/log/files.md && \
		chown -R www-data:www-data /opt/files.md && \
		echo 'Directories created and permissions set successfully.' \
	"

deploy:
	@GREEN='\e[32m'; \
	YELLOW='\e[33m'; \
	RESET='\e[0m'; \
	printf "$${YELLOW}Building...$${RESET}\n" && \
	make check && \
	GOOS=linux GOARCH=amd64 go build -o /tmp/bot ./cmd/bot && \
	printf "$${GREEN}Build Completed$${RESET}\n" && \
	ssh $(host) "killall bot || true" && \
	scp /tmp/bot $(host):/app/bot && printf "$${GREEN}The binary is copied on the server$${RESET}\n" && \
	ssh $(host) "sudo setcap 'cap_net_bind_service=+ep' /app/bot" && \
	ssh $(host) "su -c \"cd /app && nohup ./bot >> /app/log 2>>/app/err &\" -s /bin/sh www-data" && \
	rm /tmp/bot && \
	printf "$${GREEN}Successfully deployed!$${RESET}\n"


lint:
	golangci-lint run

format:
	gofumpt -w .