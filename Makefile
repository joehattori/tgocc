GO_OBJS=main.go

tgocc: $(GO_OBJS)
	go build -o tgocc $(GO_OBJS)

.PHONY: clean test
clean:
	rm -f tgocc *.o tmp.*

test: tgocc
	./test.sh
