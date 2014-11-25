FROM google/golang

ADD . /gopath/src/github.com/icchan/aoi-proxy

RUN go install github.com/icchan/aoi-proxy/main

# Run the outyet command by default when the container starts.
ENTRYPOINT ["/gopath/bin/main"]

# Document that the service listens on port three ports.
EXPOSE 8080
EXPOSE 8081
EXPOSE 5000
