REPO_NAME   := eudvc

test:
	@make lint
	@echo "** Testing **"
	go get -t ./...
	go clean -testcache
	go test -short -covermode=atomic  ./...

coverage:
	go get -t ./...
	go clean -testcache
	go test -short -covermode=atomic -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out

gosec:
	@echo "** gosec **"
	go get github.com/securego/gosec/cmd/gosec 
	gosec -quiet -fmt=json ./...

lint:
	@echo "** Linting **"
	go get golang.org/x/lint/golint
	golint -set_exit_status ./...

modupdate:
	@echo "updating all modules used directly"
