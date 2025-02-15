api:
	air -c .air.toml --build.cmd "go build -o bin/api cmd/api/main.go" --build.bin "./bin/api"

data:
	go run ./cmd/data/main.go

