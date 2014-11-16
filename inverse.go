package aoiproxy

import (
	"log"
	"net/http"
)

// Handler for alternate requests (s-out environment)
type InverseHandler struct {
	Target *BlueGreenHandler
}

func (ih *InverseHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	log.Printf("[Alt]   serving from NOT %s: %s", ih.Target.Environment, req.URL.Path)

	if ih.Target.Environment == BLUE {
		ih.Target.Green.ServeHTTP(w, req)
	} else {
		ih.Target.Blue.ServeHTTP(w, req)
	}
}
