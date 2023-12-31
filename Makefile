run:
	go run . -dict example_dict.txt

build:
	CGO_ENABLED=0 go build .