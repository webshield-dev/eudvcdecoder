REPO_NAME   := eudvcdecoder

test:
	@make lint
	@make gosec
	@echo "** Testing **"
	go get -t -d ./...
	go clean -testcache
	go test -short -covermode=atomic  ./...

gosec:
	@echo "** gosec **"
	go install github.com/securego/gosec/cmd/gosec
	gosec -quiet -fmt=json ./...

lint:
	@echo "** Linting **"
	go install golang.org/x/lint/golint
	golint -set_exit_status ./...
