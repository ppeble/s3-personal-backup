default: test

test:
	go test -tags test -coverprofile ./backup/test.cov ./backup
	go tool cover -func "./backup/test.cov" | awk '$$3 !~ /^100/ { print; gaps++} END { exit gaps }'

vet:
	go vet ./...

run: build
	cp build/dreamhost-personal-backup dreamhost-personal-backup
	./dreamhost-personal-backup

build: clean
	go build -o build/dreamhost-personal-backup ./

clean:
	rm -rf build
	rm -rf dreamhost-personal-backup

.PHONY: default test vet build clean
