package aoiproxy

import (
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
)

const (
	BLUE  = "blue"
	GREEN = "green"
)

type BackEnd struct {
	*httputil.ReverseProxy
	Target string
}

func NewBackEnd(backendUrl *url.URL) BackEnd {
	return BackEnd{httputil.NewSingleHostReverseProxy(backendUrl), backendUrl.String()}
}

// Handler for blue green switcher
type BlueGreenHandler struct {
	Blue        BackEnd
	Green       BackEnd
	Environment string
}

func MakeBlueGreen(blueString string, greenString string) *BlueGreenHandler {
	bgh := BlueGreenHandler{Environment: BLUE}

	blueUrl, _ := url.Parse(blueString)
	greenUrl, _ := url.Parse(greenString)

	bgh.Blue = NewBackEnd(blueUrl)
	bgh.Green = NewBackEnd(greenUrl)

	return &bgh
}

func (bgh *BlueGreenHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	log.Printf("[Main]  serving from %s: %s", bgh.Environment, req.URL.Path)

	if bgh.Environment == BLUE {
		bgh.Blue.ServeHTTP(w, req)
	} else {
		bgh.Green.ServeHTTP(w, req)
	}
}
