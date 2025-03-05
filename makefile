# Check to see if we can use ash, in Alpine images, or default to BASH.
SHELL_PATH = /bin/ash
SHELL = $(if $(wildcard $(SHELL_PATH)),/bin/ash,/bin/bash)


chat-run:
	go run chat/api/services/cap/main.go

chat-test:
	curl -i -X GET http://localhost:9000/test

chat-test-error:
	curl -i -X GET http://localhost:9000/testerror

chat-test-panic:
	curl -i -X GET http://localhost:9000/testpanic

chat-hack-0:
	go run chat/api/tooling/client/main.go 0

chat-hack-1:
	go run chat/api/tooling/client/main.go 1


# ==============================================================================
# Modules support

tidy:
	go mod tidy
	go mod vendor

deps-upgrade:
	go get -u -v ./...
	go mod tidy
	go mod vendor