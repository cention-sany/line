package line

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"

	"golang.org/x/oauth2"
)

const (
	L_BOT = iota
	L_BusinessConnect
)

const (
	businessConnUrl = "https://api.line.me/v1"
	botUrl          = "https://trialbot-api.line.me/v1"
)

var (
	baseUrl  = botUrl
	lineAPI  = L_BOT
	oauthUrl = fmt.Sprint(businessConnUrl, "/oauth")
	LINE_EP  = oauth2.Endpoint{
		AuthURL:  "https://access.line.me/dialog/oauth/weblogin",
		TokenURL: fmt.Sprint(oauthUrl, "/accessToken"),
	}
	channelID, channelSecret, channelMID string
)

type LINE struct {
	API     int    // must set this
	BaseUrl string // optional to set
}

type Channel struct {
	Id, Secret, Mid string
}

// Set channel ID
func SetID(id string) {
	channelID = id
}

// Set channel secret.
func SetSecret(secret string) {
	channelSecret = secret
}

// Set MID
func SetMID(mid string) {
	channelMID = mid
}

// Set API type
func SetAPI(api int) {
	lineAPI = api
	switch lineAPI {
	case L_BusinessConnect:
		baseUrl = businessConnUrl
	default:
		baseUrl = botUrl
	}
}

func validation(r io.Reader, h string) (bool, error) {
	mac := hmac.New(sha256.New, []byte(channelSecret))
	_, err := io.Copy(mac, r)
	if err != nil {
		return false, err
	}
	s := base64.StdEncoding.EncodeToString(mac.Sum(nil))
	if h != s {
		return false, nil
	}
	return true, nil
}

type parallelValid struct {
	pw *io.PipeWriter
	io.Reader
	reserr error
	res    bool
}

// Take in r - io.Reader - and based64-d HmacSHA256 hash h, and return
// callback function with signature "func(`block`) `validation result`, `ready`".
// The hash validation process is executed concurrently while r is being
// read and will only complete when parallelValid.Close is called. `block`
// argument is used to block until validation complete. Or use `ready` to poll
// whether validation process complete.
func newParallelValid(r io.Reader, h string) (io.ReadCloser, func(bool) (bool, error, bool)) {
	var res bool
	var reserr error

	gotRes := make(chan struct{})
	pv := new(parallelValid)
	pr, pw := io.Pipe()
	go func() {
		res, reserr = validation(pr, h)
		gotRes <- struct{}{}
	}()
	pv.Reader = io.TeeReader(r, pw)
	pv.pw = pw
	return pv, func(block bool) (bool, error, bool) {
		if block {
			select {
			case <-gotRes:
			}
		} else {
			select {
			case <-gotRes:
			default:
				return false, errors.New("line: not ready"), false
			}
		}
		return res, reserr, true
	}
}

func (p *parallelValid) Close() error {
	return p.pw.Close()
}

func decodeRawJson(v interface{}, raw json.RawMessage) error {
	r := bytes.NewReader([]byte(raw))
	e := json.NewDecoder(r)
	if err := e.Decode(v); err != nil {
		return err
	}
	return nil
}
