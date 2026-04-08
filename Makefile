.PHONY: server

server:
	go run ./cmd/server

test:
	go test ./...

check:
	go fmt ./... && go vet ./... && go test ./...

lint:
	golangci-lint run

format:
	gofumpt -w .

e2e: # make e2e test="create and move"
	killall server || true
	go run ./cmd/server & \
	cd tests && npm run test $(if $(test),-g "$(test)")

e2eh: # headed e2e tests
	killall server || true
	go run ./cmd/server & \
	cd tests && npm run test:headed $(if $(test),-g "$(test)")

e2es: # run single test
	cd tests && npm run test -- $(if $(test),-g "$(test)")

e2esh: # run single test headed
	cd tests && npm run test:headed -- $(if $(test),-g "$(test)")

sync:
	killall server || true
	go run ./cmd/server & \
	cd tests && npm run test --g "sync"

synch:
	killall server || true
	go run ./cmd/server & \
	cd tests && npm run test:headed --g "sync"

report:
	cd tests && npx playwright show-report

init_server: # create directories and configuration files on the service
	ssh $(host) "\
		mkdir -p /app/storage && \
		mkdir -p /var/log/files.md && \
		mkdir -p /opt/files.md && \
		mkdir -p /opt/files.md/tokens && \
		chown -R www-data:www-data /app && \
		chown -R www-data:www-data /var/log/files.md && \
		chown -R www-data:www-data /opt/files.md && \
		echo 'BOT_API_TOKEN=' > /app/.env && \
		echo 'STORAGE_DIR=/app/storage' >> /app/.env && \
		echo 'CERT_DIR=/opt/files.md' >> /app/.env && \
		echo 'TOKENS_DIR=/opt/files.md/tokens' >> /app/.env && \
		echo 'LOG_FILE=/var/log/files.md/server.log' >> /app/.env && \
		chown www-data:www-data /app/.env && \
		( \
			echo '[Unit]' > /etc/systemd/system/server.service && \
			echo 'Description=Files.md Server' >> /etc/systemd/system/server.service && \
			echo 'After=network.target' >> /etc/systemd/system/server.service && \
			echo '' >> /etc/systemd/system/server.service && \
			echo '[Service]' >> /etc/systemd/system/server.service && \
			echo 'User=www-data' >> /etc/systemd/system/server.service && \
			echo 'ExecStart=/app/server' >> /etc/systemd/system/server.service && \
			echo 'WorkingDirectory=/app' >> /etc/systemd/system/server.service && \
			echo 'Environment=TOKENS_SALT=your-secret-salt-here' >> /etc/systemd/system/server.service && \
			echo 'Restart=always' >> /etc/systemd/system/server.service && \
			echo 'RestartSec=5' >> /etc/systemd/system/server.service && \
			echo 'StandardOutput=append:/app/log' >> /etc/systemd/system/server.service && \
			echo 'StandardError=append:/app/err' >> /etc/systemd/system/server.service && \
			echo 'AmbientCapabilities=CAP_NET_BIND_SERVICE' >> /etc/systemd/system/server.service && \
			echo '' >> /etc/systemd/system/server.service && \
			echo '[Install]' >> /etc/systemd/system/server.service && \
			echo 'WantedBy=multi-user.target' >> /etc/systemd/system/server.service \
		) || echo 'Failed to write service file. Check permissions.'; \
		echo 'Directories created and permissions set successfully.' \
	"

deploy_systemd: # deploy as systemd service, TODO make timestamps hussle in separate dir, add js/css minify before release
	@GREEN='\e[32m'; \
	YELLOW='\e[33m'; \
	RESET='\e[0m'; \
	COMMIT_HASH=$$(git rev-parse --short HEAD); \
	printf "$${YELLOW}Building...$${RESET}\n" && \
	make check && \
	GOOS=linux GOARCH=amd64 go build -o /tmp/server ./cmd/server && \
	printf "$${GREEN}Build Completed$${RESET}\n" && \
	scp /tmp/server $(host):/app/server.new && printf "$${GREEN}The binary is copied on the server$${RESET}\n" && \
	ssh $(host) "mv /app/server.new /app/server && systemctl daemon-reload && systemctl restart server.service" && \
	rm /tmp/server && \
	printf "$${YELLOW}Versioning current files with commit: $${COMMIT_HASH}$${RESET}\n" && \
	TMPWEB=$$(mktemp -d) && \
	cp -r web "$${TMPWEB}/web" && \
	find "$${TMPWEB}/web" -name "*.html" -exec grep -l "?v=" {} \; | xargs sed -i '' 's/?v=/?v='"$${COMMIT_HASH}"'/g' && \
	tar --no-xattrs --disable-copyfile --no-fflags -czf "$${TMPWEB}/web.tar.gz" -C "$${TMPWEB}" web && \
	scp "$${TMPWEB}/web.tar.gz" files:/app/ && \
	ssh files "cd /app && tar -xzf web.tar.gz && rm web.tar.gz" && \
	rm -rf "$${TMPWEB}" && \
	printf "$${GREEN}Successfully deployed!$${RESET}\n"

deploy_binary: # deploy as regular binary, kinda deprecated, but ok for simple setup
	@GREEN='\e[32m'; \
	YELLOW='\e[33m'; \
	RESET='\e[0m'; \
	printf "$${YELLOW}Building...$${RESET}\n" && \
	make check && \
	GOOS=linux GOARCH=amd64 go build -o /tmp/server ./cmd/server && \
	printf "$${GREEN}Build Completed$${RESET}\n" && \
	ssh $(host) "killall server || true" && \
	scp /tmp/server $(host):/app/server && printf "$${GREEN}The binary is copied on the server$${RESET}\n" && \
  	ssh $(host) "sudo setcap 'cap_net_bind_service=+ep' /app/server" && \
	ssh $(host) "su -c \"cd /app && nohup ./server >> /app/log 2>>/app/err &\" -s /bin/sh www-data" && \
	rm /tmp/server && \
	printf "$${GREEN}Successfully deployed!$${RESET}\n"

