.SILENT :
.PHONY : tcping clean fmt

TAG:=`git describe --abbrev=0 --tags`
GITCOMMIT:=`git rev-parse HEAD`
LDFLAGS:=-X main.version=$(TAG) -X main.gitCommit=$(GITCOMMIT)

all: tcping

tcping:
	echo "Building tcping"
	go install -ldflags "$(LDFLAGS)"

dist-clean:
	rm -rf dist
	rm -f tcping-*.tar.gz

dist: dist-clean
	mkdir -p dist/alpine-linux/amd64 && GOOS=linux GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -a -tags netgo -installsuffix netgo -o dist/alpine-linux/amd64/tcping
	mkdir -p dist/linux/amd64 && GOOS=linux GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o dist/linux/amd64/tcping
	mkdir -p dist/linux/armel && GOOS=linux GOARCH=arm GOARM=5 go build -ldflags "$(LDFLAGS)" -o dist/linux/armel/tcping
	mkdir -p dist/linux/armhf && GOOS=linux GOARCH=arm GOARM=6 go build -ldflags "$(LDFLAGS)" -o dist/linux/armhf/tcping
	mkdir -p dist/darwin/amd64 && GOOS=darwin GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o dist/darwin/amd64/tcping
	mkdir -p dist/windows/amd64 && GOOS=windows GOARCH=amd64 GOARM=6 go build -ldflags "$(LDFLAGS)" -o dist/windows/amd64/tcping.exe

release: dist
	tar -cvzf tcping-alpine-linux-amd64-$(TAG).tar.gz -C dist/alpine-linux/amd64 tcping
	tar -cvzf tcping-linux-amd64-$(TAG).tar.gz -C dist/linux/amd64 tcping
	tar -cvzf tcping-linux-armel-$(TAG).tar.gz -C dist/linux/armel tcping
	tar -cvzf tcping-linux-armhf-$(TAG).tar.gz -C dist/linux/armhf tcping
	tar -cvzf tcping-darwin-amd64-$(TAG).tar.gz -C dist/darwin/amd64 tcping
	tar -cvzf tcping-windows-amd64-$(TAG).tar.gz -C dist/windows/amd64 tcping.exe
