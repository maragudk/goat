.PHONY: benchmark
benchmark:
	go test -bench=.

.PHONY: clean
clean:
	rm -rf .goat

.PHONY: cover
cover:
	go tool cover -html=cover.out

.PHONY: down
down:
	docker compose down

.PHONY: download
download:
	mkdir -p models
	cd models && curl -L -O -C - https://assets.maragu.dev/llm/Llama-3.2-1B-Instruct-Q5_K_M.llamafile
	cd models && curl -L -O -C - https://assets.maragu.dev/llm/Llama-3.2-3B-Instruct-Q5_K_M.llamafile
	cd models && curl -L -O -C - https://assets.maragu.dev/llm/Meta-Llama-3.1-8B-Instruct-Q5_K_M.llamafile

.PHONY: install
install:
	go install .

.PHONY: lint
lint:
	golangci-lint run

.PHONY: test
test:
	go test -coverprofile=cover.out -shuffle on ./...

.PHONY: up
up:
	docker compose up -d
	./models/Meta-Llama-3.1-8B-Instruct-Q5_K_M.llamafile --nobrowser --port 8092
