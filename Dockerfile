FROM golang:1.17.5-buster

COPY . /home/src
RUN (cd /home/src && go build -o /bin/action .)

ENTRYPOINT ["/bin/action"]