package line

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"html"
	"io"
	"net/http"
	"os"
)

var botEvent = fmt.Sprint(baseUrl, "/events")
var gBotServer = &BOTServer{}

type BOTServer struct {
	mH func(*Message)   // callback message handler
	oH func(*Operation) // callback operation handler
}

func NewBOTServer() *BOTServer {
	return &BOTServer{}
}

func (b *BOTServer) SetMessageHandler(f func(m *Message)) {
	b.mH = f
}

func (b *BOTServer) SetOperationHandler(f func(p *Operation)) {
	b.oH = f
}

func (b *BOTServer) Serve(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprintf(w, "invalid page, %q", html.EscapeString(r.URL.Path))
		return
	}
	h := r.Header
	s := h.Get("X-Line-Channelsignature")
	if h.Get("Content-Type") != "application/json" || s == "" {
		invalidPost(w, r.URL.Path, "")
		return
	}
	rc, f := newParallelValid(r.Body, s)
	var result Result
	e := json.NewDecoder(rc)
	if err := e.Decode(result); err != nil {
		invalidPost(w, r.URL.Path, err.Error())
		return
	}
	if err := rc.Close(); err != nil {
		invalidPost(w, r.URL.Path, err.Error())
		return
	}
	if ok, err, _ := f(true); err != nil {
		invalidPost(w, r.URL.Path, err.Error())
		return
	} else {
		if !ok {
			invalidPost(w, r.URL.Path, "invalid signature")
			return
		}
	}
	// process the received data
	for i, _ := range result.R {
		switch result.R[i].EventType {
		case ET_OP_MSG:
			var msg Message
			if err := decodeRawJson(&msg, result.R[i].Content); err != nil {
				invalidPost(w, r.URL.Path, err.Error())
				return
			}
			b.mH(&msg)
		case ET_OP_ADD:
			var opr Operation
			if err := decodeRawJson(&opr, result.R[i].Content); err != nil {
				invalidPost(w, r.URL.Path, err.Error())
				return
			}
			b.oH(&opr)
		}
	}
	w.WriteHeader(http.StatusOK)
	return
}

func invalidPost(w http.ResponseWriter, s, err string) {
	w.WriteHeader(470)
	fmt.Fprintf(w, "invalid post request, %q error: %s", html.EscapeString(s), err)
	return
}

func BOTServe(w http.ResponseWriter, r *http.Request) {
	gBotServer.Serve(w, r)
}

func SetMessageHandler(f func(m *Message)) {
	gBotServer.mH = f
}

func SetOperationHandler(f func(p *Operation)) {
	gBotServer.oH = f
}

func setHeader(h http.Header) {
	h["Content-Type"] = []string{"application/json", "charset=UTF-8"}
	h["X-Line-ChannelID"] = []string{channelID}
	h["X-Line-ChannelSecret"] = []string{channelSecret}
	h["X-Line-Trusted-User-With-ACL"] = []string{channelMID}
}

func sendMessages(to []string, evt string, r json.RawMessage) error {
	v := SndMsg{
		To:        to,
		ToChannel: TO_Ch,
		EventType: evt,
		Content:   r,
	}
	b := new(bytes.Buffer)
	e := json.NewEncoder(b)
	if err := e.Encode(v); err != nil {
		return err
	}
	req, err := http.NewRequest("POST", botEvent, b)
	if err != nil {
		return err
	}
	req.Header["Content-Type"] = []string{"application/json", "charset=UTF-8"}
	req.Header["X-Line-ChannelID"] = []string{channelID}
	req.Header["X-Line-ChannelSecret"] = []string{channelSecret}
	req.Header["X-Line-Trusted-User-With-ACL"] = []string{channelMID}
	c := http.DefaultClient
	resp, err := c.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return errors.New(fmt.Sprint("bot: invalid response ", resp.Status))
	}
	return nil
}

func chkLimit(to []string) error {
	if len(to) > targetLimit {
		return errors.New("line: exist target users limit")
	}
	return nil
}

func SendText(to []string, txt string) error {
	if err := chkLimit(to); err != nil {
		return err
	}
	v := Text{
		Typ:   CT_Txt,
		ToTyp: RT_Usr,
		Txt:   txt,
	}
	b := new(bytes.Buffer)
	e := json.NewEncoder(b)
	if err := e.Encode(v); err != nil {
		return err
	}
	return sendMessages(to, ET_MultiMedia, b.Bytes())
}

func sendMedia(to []string, curl, preurl string, media int) error {
	if err := chkLimit(to); err != nil {
		return err
	}
	v := ViewMedia{
		Typ:      media,
		ToTyp:    RT_Usr,
		Original: curl,
		Preview:  preurl,
	}
	b := new(bytes.Buffer)
	e := json.NewEncoder(b)
	if err := e.Encode(v); err != nil {
		return err
	}
	return sendMessages(to, ET_MultiMedia, b.Bytes())
}

func SendImage(to []string, curl, preurl string) error {
	return sendMedia(to, curl, preurl, CT_Image)
}

func SendVideo(to []string, curl, preurl string) error {
	return sendMedia(to, curl, preurl, CT_Video)
}

func SendAudio(to []string, url, d string) error {
	if err := chkLimit(to); err != nil {
		return err
	}
	v := Audio{
		Typ:      CT_Audio,
		ToTyp:    RT_Usr,
		Original: url,
		Meta:     AudLen{Audlen: d},
	}
	b := new(bytes.Buffer)
	e := json.NewEncoder(b)
	if err := e.Encode(v); err != nil {
		return err
	}
	return sendMessages(to, ET_MultiMedia, b.Bytes())
}

func SendLocation(to []string, s string, lat, long float32) error {
	if err := chkLimit(to); err != nil {
		return err
	}
	v := SndLocation{
		Typ:   CT_Loc,
		ToTyp: RT_Usr,
		Txt:   s,
		Loc:   Location{Title: s, Lat: lat, Long: long},
	}
	b := new(bytes.Buffer)
	e := json.NewEncoder(b)
	if err := e.Encode(v); err != nil {
		return err
	}
	return sendMessages(to, ET_MultiMedia, b.Bytes())
}

func SendSticker(to []string, id, pid, ver string) error {
	if err := chkLimit(to); err != nil {
		return err
	}
	v := SndSticker{
		Typ:   CT_Sticker,
		ToTyp: RT_Usr,
		Meta:  Sticker{PId: pid, Id: id, Ver: ver},
	}
	b := new(bytes.Buffer)
	e := json.NewEncoder(b)
	if err := e.Encode(v); err != nil {
		return err
	}
	return sendMessages(to, ET_MultiMedia, b.Bytes())
}

func SendMultipleMessages(to []string, v *MultiMsgs) error {
	if err := chkLimit(to); err != nil {
		return err
	}
	b := new(bytes.Buffer)
	e := json.NewEncoder(b)
	if err := e.Encode(v); err != nil {
		return err
	}
	return sendMessages(to, ET_Msgs, b.Bytes())
}

func SendRichMessages(to []string, url, s string, rm *RichMsg) error {
	if err := chkLimit(to); err != nil {
		return err
	}
	b := new(bytes.Buffer)
	e := json.NewEncoder(b)
	if err := e.Encode(rm); err != nil {
		return err
	}
	v := RichMsgContent{
		Typ:   CT_Rich_Msg,
		ToTyp: RT_Usr,
		Meta: &RichMsgMeta{
			Url: url,
			Rev: "1",
			Alt: s,
			Jsn: b.String(),
		},
	}
	b = new(bytes.Buffer)
	e = json.NewEncoder(b)
	if err := e.Encode(v); err != nil {
		return err
	}
	return sendMessages(to, ET_MultiMedia, b.Bytes())
}

func getMessageReader(msgid string, isPre bool) (io.ReadCloser, error) {
	var u string
	if isPre {
		u = fmt.Sprint(baseUrl, "/v1/bot/message/", msgid, "/content/preview")
	} else {
		u = fmt.Sprint(baseUrl, "/v1/bot/message/", msgid, "/content")
	}
	req, err := http.NewRequest("GET", u, nil)
	if err != nil {
		return nil, err
	}
	req.Header["Content-Type"] = []string{"application/json", "charset=UTF-8"}
	req.Header["X-Line-ChannelID"] = []string{channelID}
	req.Header["X-Line-ChannelSecret"] = []string{channelSecret}
	req.Header["X-Line-Trusted-User-With-ACL"] = []string{channelMID}
	c := http.DefaultClient
	resp, err := c.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != 200 {
		resp.Body.Close()
		return nil, errors.New(fmt.Sprint("line: invalid response ", resp.Status))
	}
	return resp.Body, nil
}

// Get image or video message content (full) and save it into saveTo filename.
func GetMessageSave(msgid, saveTo string) error {
	r, err := GetMessageContentReader(msgid)
	if err != nil {
		return err
	}
	defer r.Close()
	f, err := os.OpenFile(saveTo, os.O_WRONLY, os.ModePerm)
	if err != nil {
		return err
	}
	defer f.Close()
	b := bufio.NewReader(r)
	if _, err = b.WriteTo(f); err != nil {
		return err
	}
	return nil
}

func GetMessageContentReader(msgid string) (io.ReadCloser, error) {
	return getMessageReader(msgid, false)
}

func GetMessagePreviewReader(msgid string) (io.ReadCloser, error) {
	return getMessageReader(msgid, true)
}
