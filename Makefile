SRCS=main.go token.go parse.go gen.go node.go type.go var.go

tgocc: $(SRCS)
	go build -o tgocc $(SRCS)

clean:
	rm -f tgocc *.o tmp*

test: tgocc
	./tgocc test.c > tmp.s
	cc -xc -c -o tmp2.o example/predefined_functions.c
	cc -static -o tmp tmp.s tmp2.o
	./tmp

test_old: tgocc
	./test.sh

.PHONY: clean test tgocc
