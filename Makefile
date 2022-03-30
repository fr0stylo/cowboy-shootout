build:
	CGO_ENABLED=0 GOOS=linux go build  -ldflags '-w -s' -installsuffix cgo -o bin/starter ./starter/starter.go
	CGO_ENABLED=0 GOOS=linux go build  -ldflags '-w -s' -installsuffix cgo -o bin/shooter ./shooter/shooter.go

run: build
	./bin/starter  -shooters ./shooters.json -executable ./bin/shooter

.PHONY: build
