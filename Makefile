SRCS=main.go token.go parse.go gen.go node.go type.go var.go

tgocc: $(SRCS)
	go build -o tgocc $(SRCS)

clean:
	rm -f tgocc *.o tmp*

test: tgocc
	./tgocc test.c > tmp.s
	cc -static -o tmp tmp.s
	./tmp

test_old: tgocc
	cc -xc -c -o tmp2.o example/predefined_functions.c
	./test.sh

.PHONY: clean test tgocc
