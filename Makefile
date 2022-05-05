build:
	go build -v ./...

test:
	go test -v ./... -run 'Unit'

integration:
	go test -v ./... -run 'Integration'
