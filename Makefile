api:
	air -c .air.toml --build.cmd "go build -o bin/api main.go" --build.bin "./bin/api"

data:
	go run ./cmd/data/main.go

css:
	npx @tailwindcss/cli -i ./assets/css/input.css -o ./assets/css/output.css --minify

css-watch:
	npx @tailwindcss/cli -i ./assets/css/input.css -o ./assets/css/output.css --watch

