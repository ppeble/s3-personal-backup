default: test

test:
	go test -cover ./

vet:
	@go vet ./...

run: build
	./dreamhost-personal-backup

build: clean
	go build -o dreamhost-personal-backup

clean:
	rm -rf dreamhost-personal-backup

.PHONY: default test vet build clean
