SRCS=main.go token.go parse.go gen.go

tgocc: $(SRCS)
	go build -o tgocc $(SRCS)

.PHONY: clean test
clean:
	rm -f tgocc *.o tmp*

test: tgocc
	./test.sh
