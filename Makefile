.PHONY=build_csp

build_origin:
	@go build -o bin/main origin/main.go

build_concurrent:
	@go build -o bin/main concurrent/main.go

build_parallel:
	@go build -o bin/main parallel_walk_tree/main.go

build_buffer:
	@go build -o bin/main buffer/main.go

run_origin: build_origin
	@./bin/main .

run_concurrent: build_concurrent
	@./bin/main .

run_parallel: build_parallel
	@./bin/main .

run_buffer: build_buffer
	@./bin/main .