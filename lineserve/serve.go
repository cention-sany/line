/*
This binary is to serve callback url for LINE authorization.
It also reverse proxy to Microsoft Exchange server.
*/
package main

import (
	"crypto/tls"
	"flag"
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"time"

	"golang.org/x/oauth2"

	"github.com/cention-sany/line"
	"github.com/kataras/iris"
	lb "github.com/line/line-bot-sdk-go/linebot"
)

const (
	usoftIP    = "https://127.0.0.1:444"
	callback   = "https://sanylcs.dynu.com/line/callback"
	vkCallback = "https://sanylcs.dynu.com/vk/callback"
	vkState    = "centionvksecret"
)

var (
	usoftES       *httputil.ReverseProxy
	cID           int64
	cSecret, cMID string
	mainState     int
	linec         *lb.Client
	api           *line.API
	auth          *line.Auth
)

func serve(rw http.ResponseWriter, req *http.Request) {
	usoftES.ServeHTTP(rw, req)
}

func initMicrosoftExchangeServerProxy() {
	url, _ := url.Parse(usoftIP)
	usoftES = httputil.NewSingleHostReverseProxy(url)
	usoftES.Transport = &http.Transport{
		TLSClientConfig:       &tls.Config{InsecureSkipVerify: true},
		TLSHandshakeTimeout:   time.Minute,
		ResponseHeaderTimeout: time.Minute,
		//ExpectContinueTimeout: time.Minute,
		DisableCompression: true,
	}
}

func lineOauth2Handler(c *iris.Context) {
	// Use the authorization code that is pushed to the redirect URL.
	// NewTransportWithCode will do the handshake to retrieve
	// an access token and initiate a Transport that is
	// authorized and authenticated by the retrieved token.
	var code string
	api = auth.NewAPI(oauth2.NoContext, code)
	token := api.Token.AccessToken
	c.HTML(fmt.Sprint("<p><b>token: ", token, "</b></p>"))
}

func lineHandler(w http.ResponseWriter, r *http.Request) {
	ress, err := linec.ParseRequest(r)
	if err != nil {
		w.WriteHeader(404)
		fmt.Fprintf(w, `<html><head></head><body>
				<p>Error: %s</p></body></html>`, err.Error())
		return
	}
	fmt.Fprint(w, `<html><head></head><body>`)
	for _, res := range ress.Results {
		fmt.Fprintf(w, `<p>Received something:</p>
				<p>ID: %s, From: %s</p>`, res.ID, res.From)
		rxc := res.Content()
		var s string
		switch rxc.ContentType {
		case lb.ContentTypeText:
			content, err := rxc.TextContent()
			if err != nil {
				s = fmt.Sprintf(`encounter error: %s`, err.Error())
				break
			}
			s = fmt.Sprintf(`Text: %s`, content.Text)
		case lb.ContentTypeImage:
			_, err = rxc.ImageContent()
			if err != nil {
				s = fmt.Sprintf(`encounter error: %s`, err.Error())
				break
			}
			s = `Received image`
		case lb.ContentTypeVideo:
			_, err = rxc.VideoContent()
			if err != nil {
				s = fmt.Sprintf(`encounter error: %s`, err.Error())
				break
			}
			s = `Received video`
		default:
			log.Println("default content:", rxc.ContentType)
		}
		fmt.Fprint(w, `<p>`, s, `</p>`)
	}
	fmt.Fprint(w, `</body></html>`)
}

func main() {
	if !flag.Parsed() {
		flag.Parse()
	}
	if cSecret == "" || cMID == "" || cID == 0 {
		log.Fatalln("Please provide valid line channel ID and channel secret and channel MID")
	}
	initMicrosoftExchangeServerProxy()
	line.SetID("abc123")
	line.SetSecret("abc123")
	c, err := lb.NewClient(cID, cSecret, cMID)
	if err != nil {
		log.Fatalln(err)
	}
	linec = c
	usoftHandler := iris.ToHandlerFunc(usoftES.ServeHTTP)
	iris.Get("/ping", func(c *iris.Context) {
		c.HTML("<b>pong pong</b>")
	})
	iris.Get("/line", func(c *iris.Context) {
		auth = line.GetAuth(callback, "random123", nil)
		// Redirect user to consent page to ask for permission
		// for the scopes specified above.
		url := auth.GetURL()
		log.Printf("Visit the URL for the auth dialog: %v\n", url)
		c.HTML(fmt.Sprint("<b><a href=\"", url, "\">Click Me</a></b>"))
	})
	iris.Get("/line/callback", iris.ToHandlerFunc(lineHandler))
	iris.Get("/ews/*anything", usoftHandler)
	iris.Get("/owa/*anything", usoftHandler)
	iris.Get("/ecp/*anything", usoftHandler)
	iris.Post("/ews/*anything", usoftHandler)
	iris.Post("/owa/*anything", usoftHandler)
	iris.Post("/ecp/*anything", usoftHandler)
	iris.Head("/ews/*anything", usoftHandler)
	iris.Head("/owa/*anything", usoftHandler)
	iris.Head("/ecp/*anything", usoftHandler)
	iris.Delete("/ews/*anything", usoftHandler)
	iris.Delete("/owa/*anything", usoftHandler)
	iris.Delete("/ecp/*anything", usoftHandler)
	iris.Put("/ews/*anything", usoftHandler)
	iris.Put("/owa/*anything", usoftHandler)
	iris.Put("/ecp/*anything", usoftHandler)
	iris.Connect("/ews/*anything", usoftHandler)
	iris.Connect("/owa/*anything", usoftHandler)
	iris.Connect("/ecp/*anything", usoftHandler)
	iris.Get("/ews", usoftHandler)
	iris.Get("/owa", usoftHandler)
	iris.Get("/ecp", usoftHandler)
	iris.Post("/ews", usoftHandler)
	iris.Post("/owa", usoftHandler)
	iris.Post("/ecp", usoftHandler)
	iris.Head("/ews", usoftHandler)
	iris.Head("/owa", usoftHandler)
	iris.Head("/ecp", usoftHandler)
	iris.Delete("/ews", usoftHandler)
	iris.Delete("/owa", usoftHandler)
	iris.Delete("/ecp", usoftHandler)
	iris.Put("/ews", usoftHandler)
	iris.Put("/owa", usoftHandler)
	iris.Put("/ecp", usoftHandler)
	iris.Connect("/ews", usoftHandler)
	iris.Connect("/owa", usoftHandler)
	iris.Connect("/ecp", usoftHandler)
	iris.Get("/Autodiscover/*anything", usoftHandler)
	iris.Post("/Autodiscover/*anythingws", usoftHandler)
	iris.Delete("/Autodiscover/*anything", usoftHandler)
	iris.Put("/Autodiscover/*anything", usoftHandler)
	iris.Connect("/Autodiscover/*anything", usoftHandler)
	iris.Get("/Autodiscover", usoftHandler)
	iris.Post("/Autodiscover", usoftHandler)
	iris.Delete("/Autodiscover", usoftHandler)
	iris.Put("/Autodiscover", usoftHandler)
	iris.Connect("/Autodiscover", usoftHandler)
	iris.ListenTLS(
		":443",
		"certs/c1/fullchain1.pem",
		"certs/c1/privkey1.pem")
	//		"certs/c2/fullchain1.pem",
	//		"certs/c2/privkey.pem")
	//log.Fatal(http.ListenAndServe(":8088", iris.Serve()))
}

func init() {
	const (
		cidStr     = "Line channel ID"
		csecretStr = "Line channel secret"
		midStr     = "Line channel MID"
		stateStr   = "State of this program.\n\t0: no"
	)
	flag.Int64Var(&cID, "cid", 0, cidStr)
	flag.StringVar(&cSecret, "secret", "", csecretStr)
	flag.StringVar(&cMID, "mid", "", midStr)
	flag.IntVar(&mainState, "state", 0, stateStr)
}
