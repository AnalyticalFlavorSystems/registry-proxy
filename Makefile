EXECUTABLE := registry-proxy

all: $(EXECUTABLE)


dep:
		go list -f '{{join .Deps "\n"}}' | xargs go list -e -f '{{if not .Standard}}{{.ImportPath}}{{end}}' | xargs go get -u

clean: 
		rm $(EXECUTABLE) || true

$(EXECUTABLE): dep build
		CGO_ENABLED=0 GOOS=linux go build -a -tags netgo -installsuffix netgo -ldflags '-w' -o "$@"
