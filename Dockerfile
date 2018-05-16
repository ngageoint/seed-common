ARG IMAGE=golang:alpine
FROM $IMAGE

LABEL VERSION="1.0.0" \
    RUN="docker run -d -p 9000:9000 -p 80:80 -v <silo db/log location>:/usr/silo silo" \
    SOURCE="https://github.com/ngageoint/seed-common" \
    DESCRIPTION="seed-common library" \
    CLASSIFICATION="UNCLASSIFIED"

COPY . $GOPATH/src/github.com/ngageoint/seed-common

CMD cd $GOPATH/src/github.com/ngageoint/seed-common && ls -l