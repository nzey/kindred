package main

import (
	"bytes"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"

	_ "github.com/jinzhu/gorm/dialects/postgres"
)

func twilioProxy(w http.ResponseWriter, r *http.Request) {
	log.Println("receiving twilio request from client")
	r.Host = "localhost:3000"

	j := strings.Join(r.URL.Query()["q"], "")
	u, _ := url.Parse("http://localhost:300/api/twilio" + j)
	proxy := httputil.NewSingleHostReverseProxy(u)

	proxy.Transport = &transport{CapturedTransport: http.DefaultTransport}
	proxy.ServeHTTP(w, r)
}

type transport struct {
	CapturedTransport http.RoundTripper
}

func (t *transport) RoundTrip(request *http.Request) (*http.Response, error) {
	// response, err := http.DefaultTransport.RoundTrip(request)
	response, err := t.CapturedTransport.RoundTrip(request)
	bodyBytes, err := ioutil.ReadAll(response.Body)

	// body, err := httputil.DumpResponse(response, true)
	if err != nil {
		return nil, err
	}

	defer response.Body.Close()
	response.Body = ioutil.NopCloser(bytes.NewBuffer(bodyBytes))

	log.Println("proxy reponse is", response)

	return response, err
}
