package aoiproxy

import (
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"time"
)

const (
	BLUE  = "blue"
	GREEN = "green"
)

type BackEnd struct {
	*httputil.ReverseProxy
	Target string
	Name   string
}

func NewBackEnd(backendUrl *url.URL, name string) BackEnd {
	return BackEnd{httputil.NewSingleHostReverseProxy(backendUrl), backendUrl.String(), name}
}

// Handler for blue green switcher
type BlueGreenHandler struct {
	Blue        BackEnd
	Green       BackEnd
	Environment string
}

func NewBlueGreen(blueString string, greenString string) *BlueGreenHandler {
	bgh := BlueGreenHandler{Environment: BLUE}

	blueUrl, _ := url.Parse(blueString)
	greenUrl, _ := url.Parse(greenString)

	bgh.Blue = NewBackEnd(blueUrl, BLUE)
	bgh.Green = NewBackEnd(greenUrl, GREEN)

	return &bgh
}

func (bgh *BlueGreenHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	if bgh.Environment == BLUE {
		bgh.Blue.ServeHTTP(w, req)
	} else {
		bgh.Green.ServeHTTP(w, req)
	}
}

func logTime(tag string, start time.Time) {
	log.Printf("[Timer Log] %s : %s", tag, time.Since(start))
}

func (be *BackEnd) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	defer logTime("[be]  serving from "+be.Name+" ("+be.Target+"): "+req.URL.Path, time.Now())
	be.ReverseProxy.ServeHTTP(w, req)
}
