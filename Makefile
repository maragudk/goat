.PHONY: benchmark
benchmark:
	go test -bench=.

.PHONY: cover
cover:
	go tool cover -html=cover.out

.PHONY: download
download:
	mkdir -p models
	cd models && curl -L -O -C - https://assets.maragu.dev/llm/Llama-3.2-1B-Instruct-Q5_K_M.llamafile
	cd models && curl -L -O -C - https://assets.maragu.dev/llm/Llama-3.2-3B-Instruct-Q5_K_M.llamafile

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
	./models/Llama-3.2-1B-Instruct-Q5_K_M.llamafile --nobrowser -c 0 --port 8090 &
	./models/Llama-3.2-3B-Instruct-Q5_K_M.llamafile --nobrowser -c 0 --port 8091 &
