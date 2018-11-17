default: test

PACKAGES:="./. ./worker ./logger ./reporter"

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
	./dreamhost-personal-backup

build: clean
	go build -o build/dreamhost-personal-backup ./cmd/dreamhost-personal-backup
	cp build/dreamhost-personal-backup dreamhost-personal-backup

clean:
	rm -rf build
	rm -rf dreamhost-personal-backup

.PHONY: default test vet build clean run
