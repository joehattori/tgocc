SRCS=main.go token.go parse.go gen.go node.go type.go var.go

tgocc: $(SRCS)
	go build -o tgocc $(SRCS)

clean:
	rm -f tgocc *.o tmp*

test: tgocc
	./test.sh

.PHONY: clean test tgocc
