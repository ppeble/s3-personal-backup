default: test

PACKAGES:="./pkg/backup ./pkg/worker ./pkg/logger ./pkg/reporter"

test: vet
	@go list -f '{{.Dir}}/test.cov {{.ImportPath}}' "$(PACKAGES)"  \
		| while read coverage package ; do go test -tags test -coverprofile "$$coverage" "$$package" ; done \
		| awk -W interactive '{ print } /^FAIL/ { failures++ } END { exit failures }' ;
	@go list -f '{{.Dir}}/test.cov' "$(PACKAGES)" \
		| while read coverage ; do go tool cover -func "$$coverage" ; done \
		| awk '$$3 !~ /^100/ { print; gaps++ } END { exit gaps }' ;

vet:
	go vet ./...

run: build
	./s3-personal-backup

build: clean
	go build -o build/s3-personal-backup ./cmd/s3-personal-backup
	cp build/s3-personal-backup s3-personal-backup

clean:
	rm -rf build
	rm -rf s3-personal-backup

.PHONY: default test vet build clean run
