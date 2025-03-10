all: florilegium

again: clean all

florilegium: florilegium.go cmd/florilegium/main.go
	go build -C cmd/florilegium -o ../../florilegium

clean:
	rm -f florilegium

test:
	go test -cover

push:
	got send
	git push github

fmt:
	gofmt -s -w *.go cmd/*/main.go

cover:
	go test -coverprofile=cover.out
	go tool cover -html cover.out

README.md: README.gmi
	sisyphus -f markdown <README.gmi >README.md

doc: README.md

release: push
	git push github --tags
