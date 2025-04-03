tidy:
	go mod tidy

clean:
	go clean -testcache

test: tidy clean
	go test --count 5 -failfast -cover ./...

test-ptr: tidy clean
	go test -run $(ptr) ./...

lint: tidy
	golangci-lint run

docs:
	godoc -http :8080
