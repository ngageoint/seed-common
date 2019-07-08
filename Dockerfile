ARG IMAGE=golang:alpine
FROM $IMAGE

LABEL VERSION="1.0.0" \
    RUN="docker run" \
    SOURCE="https://github.com/ngageoint/seed-common" \
    DESCRIPTION="seed-common library" \
    CLASSIFICATION="UNCLASSIFIED"

COPY . $GOPATH/src/github.com/ngageoint/seed-common

# requires a running registry with seed-common example images built
CMD cd $GOPATH/src/github.com/ngageoint/seed-common && go test ./...