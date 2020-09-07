SRCS=*.go

tgocc: $(SRCS)
	go build -o tgocc $(SRCS)

clean:
	rm -f tgocc *.o tmp*

test: tgocc
	./tgocc test/test1.c > tmp1.s
	cc -xc -c test/util.c -o tmp_util.o
	cc -xc -c test/test2.c -o tmp2.o
	cc -static -o tmp tmp1.s tmp_util.o tmp2.o
	#./tgocc test/util.c > tmp_util.s
	#./tgocc test/test2.c > tmp2.s
	#cc -no-pie -o tmp tmp1.s tmp2.s tmp_util.s
	./tmp

.PHONY: clean test tgocc
