BENCH=.

bench:
	go test -v -benchmem -bench="$(BENCH)"
