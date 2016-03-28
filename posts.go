package line

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"golang.org/x/oauth2"
)

func postMsg(a *API, v interface{}) (*PostResponse, error) {
	var p PostResponse
	b := new(bytes.Buffer)
	e := json.NewEncoder(b)
	if err := e.Encode(v); err != nil {
		return nil, err
	}
	req, err := http.NewRequest("POST", fmt.Sprint(baseUrl, "/events"), b)
	if err != nil {
		return nil, err
	}
	req.Header["Content-Type"] = []string{"application/json"}
	req.Header["X-Line-ChannelToken"] = []string{a.Token.AccessToken}
	c := a.Config.Client(oauth2.NoContext, a.Token)
	resp, err := c.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	d := json.NewDecoder(resp.Body)
	err = d.Decode(&p)
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func (a *API) SendLinkMsg(m *TimelineContent) (*PostResponse, error) {
	return postMsg(a, m)
}

func (a *API) PostTimeline(m *TimelineMsg) (*PostResponse, error) {
	return postMsg(a, m)
}
