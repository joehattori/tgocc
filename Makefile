SRCS=*.go

tgocc: $(SRCS)
	go build -o tgocc $(SRCS)

clean:
	rm -f tgocc *.o tmp*

test: tgocc
	./tgocc test/util.c > tmp_util.s
	./tgocc test/test1.c > tmp1.s
	./tgocc test/test2.c > tmp2.s
	cc -c tmp_util.s -o tmp_util.o
	cc -c tmp1.s -o tmp1.o
	cc -xc -c test/test2.c -o tmp2.o
	cc -static -o tmp tmp_util.o tmp1.o tmp2.o
	./tmp

.PHONY: clean test tgocc
