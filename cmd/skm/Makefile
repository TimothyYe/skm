# Binary name
BINARY=skm
# Builds the project
build:
		go build -ldflags "-X main.Version=${VERSION}" -o ${BINARY}
		go test -v
# Installs our project: copies binaries
install:
		go install
release:
		# Clean
		go clean
		rm -rf *.gz
		# Build for mac
		go build -ldflags "-X main.Version=${VERSION}"
		tar czvf ${BINARY}-mac64-${VERSION}.tar.gz ./${BINARY}
		# Build for arm
		go clean
		CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -ldflags "-X main.Version=${VERSION}"
		tar czvf ${BINARY}-arm64-${VERSION}.tar.gz ./${BINARY}
		# Build for linux
		go clean
		CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags "-X main.Version=${VERSION}"
		tar czvf ${BINARY}-linux64-${VERSION}.tar.gz ./${BINARY}
		go clean
# Cleans our projects: deletes binaries
clean:
		go clean
		rm -rf *.gz

.PHONY:  clean build
