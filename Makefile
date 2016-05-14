default: test

test:
	go test -cover ./

vet:
	@go vet ./...

run: build
	cp build/dreamhost-personal-backup dreamhost-personal-backup
	./dreamhost-personal-backup

build: clean
	go build -o build/dreamhost-personal-backup ./bin

clean:
	rm -rf build
	rm -rf dreamhost-personal-backup

.PHONY: default test vet build clean
