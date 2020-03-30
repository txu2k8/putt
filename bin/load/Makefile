src	:=	load.c
bin	:=	load
lib	:=	-lpthread
warn	:=	-Wno-int-to-pointer-cast -Wno-pointer-to-int-cast

build:	$(src)
	gcc -g $(src) -o $(bin) $(lib) $(warn)
