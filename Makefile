GO_OBJS=main.go token.go parse.go gen.go

tgocc: $(GO_OBJS)
	go build -o tgocc $(GO_OBJS)

.PHONY: clean test
clean:
	rm -f tgocc *.o tmp*

test: tgocc
	./test.sh
