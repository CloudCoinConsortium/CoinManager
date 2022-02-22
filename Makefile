PROJECT_NAME=cloudcoin_manager

GOBASE=$(shell pwd)
GOPATH=$(GOBASE)/vendor:$(GOBASE):/home/alexander/go
GOFILES=$(wildcard *.go)

.PHONY: tester

all: build buildwin

build:
	GOPATH=$(GOPATH) go build -o $(PROJECT_NAME) -v $(GOFILES)  
#working 95012be384b6ca69b9a6daa6091b1cd2 webview.dll
#working 95012be384b6ca69b9a6daa6091b1cd2 WebView2Loader.dll
prepare:
#	cp webview/libwebview2/build/native/x64/WebView2Loader* /usr/x86_64-w64-mingw32/lib/
#	ln -s /usr/x86_64-w64-mingw32/lib/libshlwapi.a /usr/x86_64-w64-mingw32/lib/libShlwapi.a
#	ln -s /usr/x86_64-w64-mingw32/include/shlwapi.h /usr/x86_64-w64-mingw32/include/Shlwapi.h
#	sudo ln -s /usr/x86_64-w64-mingw32/lib/libshlwapi.a /usr/x86_64-w64-mingw32/lib/libShlwapi.a
#	cp ./vendor/github.com/polevpn/webview/WebView2Loader.dll /usr/x86_64-w64-mingw32/lib/
	cp webview/libwebview2/build/native/include/EventToken.h vendor/github.com/polevpn/webview/
	cp webview/libwebview2/build/native/include/webview2.h vendor/github.com/polevpn/webview/




buildwin:
	rsrc -ico 256x256.ico
	#CFLAGS=-I/home/alexander/dev/superraida/webview/libwebview2/build/native/include/ LD_LIBRARY_PATH=/home/alexander/dev/superraida/rdll CGO_ENABLED=1 CXX=x86_64-w64-mingw32-g++ CC=x86_64-w64-mingw32-gcc GOPATH=$(GOPATH) GOOS=windows GOARCH=amd64 go build -ldflags="-H windowsgui -extldflags \"-L /home/alexander/dev/superraida/rdll\"" -o $(PROJECT_NAME).exe -v $(GOFILES) 
	CGO_CXXFLAGS=-I/home/alexander/dev/superraida/webview/libwebview2/build/native/include/ LD_LIBRARY_PATH=/home/alexander/dev/superraida/rdll CGO_ENABLED=1 CXX=x86_64-w64-mingw32-g++ CC=x86_64-w64-mingw32-gcc GOPATH=$(GOPATH) GOOS=windows GOARCH=amd64 go build -ldflags="-H windowsgui -extldflags \"-L /home/alexander/dev/superraida/rdll\"" -o $(PROJECT_NAME).exe 

buildmac:
	#GOPATH=$(GOPATH) GOOS=darwin GOARCH=amd64 go build  -o $(PROJECT_NAME)_darwin -v $(GOFILES) 
	CGO_ENABLED=1 GOPATH=$(GOPATH) GOOS=darwin GOARCH=arm64 go build -x -o $(PROJECT_NAME)_darwin -v $(GOFILES) 

clean:
	go clean
	rm -f $(PROJECT_NAME)
	rm -f $(PROJECT_NAME).exe


c:
	gcc goaes.c aes.c -o aes

icon:
	rsrc -ico 256x256.ico

tester:
	GOPATH=$(GOPATH) go build -o raida_tester -v tester/tester.go

deps:
	
