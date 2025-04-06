tidy:
	go mod tidy

clean:
	go clean -testcache

test: tidy clean
	go test --count 5 -failfast -cover ./...

test-ptr: tidy clean
	go test -v -run $(ptr) ./...

lint: tidy
	golangci-lint run

docs:
	sleep 2s && xdg-open http://localhost:8080 &
	godoc -http :8080
