package line

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strconv"
	"testing"
	"testing/quick"
	"time"

	"github.com/stretchr/testify/assert"
)

// test BOT server message data
var tstBOTSvrMsgDt = []struct {
	name           string
	json           string
	content        []string
	receive        []Rx
	contentContent []interface{}
}{{
	name: "Single Message",
	json: `{"result":[{
    "from":"u2ddf2eb3c959e561f6c9fa2ea732e7eb8",
    "fromChannel":"1341301815",
    "to":["u0cc15697597f61dd8b01cea8b027050e"],
    "toChannel":1441301333,
    "eventType":"138311609000106303",
    "id":"ABCDEF-12345678901",
    "content": %s]}`,
	content: []string{`{
      "location":null,
      "id":"326718",
      "contentType":1,
      "from":"fff2aec188e58752ee1fb0f9507c6529a",
      "createdTime":1332394961610,
      "to":["u0a556cffd4da0dd89c94fb36e36e1cdd"],
      "toType":1,
      "contentMetadata":null,
      "text":"Hello, BOT API Server!"
    }
  }`},
	receive: []Rx{Rx{
		From:        "u2ddf2eb3c959e561f6c9fa2ea732e7eb8",
		FromChannel: "1341301815",
		To:          []string{"u0cc15697597f61dd8b01cea8b027050e"},
		ToChannel:   1441301333,
		EventType:   "138311609000106303",
		Id:          "ABCDEF-12345678901",
		Content:     nil,
	}},
	contentContent: []interface{}{Message{
		Id:    "326718",
		Typ:   1,
		Frm:   "fff2aec188e58752ee1fb0f9507c6529a",
		CTime: 1332394961610,
		To:    []string{"u0a556cffd4da0dd89c94fb36e36e1cdd"},
		ToTyp: 1,
		Txt:   "Hello, BOT API Server!"}},
}, {
	name: "Two Messages",
	json: `{"result":[{
    "from":"u2ddf2eb3c959e561f6c9fa2ea732e7999",
    "fromChannel":"4441301815",
    "to":["cccc15697597f61dd8b01cea8b027050e"],
    "toChannel":1441301222,
    "eventType":"138311609000106303",
    "id":"AACDEF-12345678902",
    "content": %s},{
    "from":"u2ddf2eb3c959e561f6c9fa2ea732e7777",
    "fromChannel":"3331301815",
    "to":["uuuu15697597f61dd8b01cea8b027050e"],
    "toChannel":5551301333,
    "eventType":"138311609000106303",
    "id":"AFFDEF-12345678900",
    "content": %s}]}`,
	content: []string{`{
      "location":null,
      "id":"555708",
      "contentType":1,
      "from":"uff2aec188e58752ee2fb0f9507c6529a",
      "createdTime":1333394961610,
      "to":["u0a556cffd4da0dd89c94fb36e36e1cdc"],
      "toType":1,
      "contentMetadata":null,
      "text":"Hello, BOT API Server1!"
    }`, `{
      "location":null,
      "id":"325708",
      "contentType":1,
      "from":"uff2aec188e58752ee1fb0f9507c6529a",
      "createdTime":1332394961611,
      "to":["u1a556cffd4da0dd89c94fb36e36e1cdc"],
      "toType":1,
      "contentMetadata":null,
      "text":"Hello, BOT API Server2!"
    }`},
	receive: []Rx{Rx{
		From:        "u2ddf2eb3c959e561f6c9fa2ea732e7999",
		FromChannel: "4441301815",
		To:          []string{"cccc15697597f61dd8b01cea8b027050e"},
		ToChannel:   1441301222,
		EventType:   "138311609000106303",
		Id:          "AACDEF-12345678902",
		Content:     nil,
	}, Rx{
		From:        "u2ddf2eb3c959e561f6c9fa2ea732e7777",
		FromChannel: "3331301815",
		To:          []string{"uuuu15697597f61dd8b01cea8b027050e"},
		ToChannel:   5551301333,
		EventType:   "138311609000106303",
		Id:          "AFFDEF-12345678900",
		Content:     nil,
	}},
	contentContent: []interface{}{
		Message{
			Id:    "555708",
			Typ:   1,
			Frm:   "uff2aec188e58752ee2fb0f9507c6529a",
			CTime: 1333394961610,
			To:    []string{"u0a556cffd4da0dd89c94fb36e36e1cdc"},
			ToTyp: 1,
			Txt:   "Hello, BOT API Server1!"},
		Message{
			Id:    "325708",
			Typ:   1,
			Frm:   "uff2aec188e58752ee1fb0f9507c6529a",
			CTime: 1332394961611,
			To:    []string{"u1a556cffd4da0dd89c94fb36e36e1cdc"},
			ToTyp: 1,
			Txt:   "Hello, BOT API Server2!"}},
}, {
	name: "Add friend operation",
	json: `{"result":[{
    "from":"uefb896062d34df287b220e7b581d2466",
    "fromChannel":"1341311815",
    "to":["u0cc55697597f61dd8b01cea8b027050e"],
    "toChannel":1441331333,
    "eventType":"138311609100106403",
    "id":"ABDDEF-22345678901",
    "content": %s}]}`,
	content: []string{`{
      "params":[
        "u0f3bfc598b061eba02183bfc5280886a",
        null,
        null
      ],
      "revision":2469,
      "opType":4
    }`},
	receive: []Rx{Rx{
		From:        "uefb896062d34df287b220e7b581d2466",
		FromChannel: "1341311815",
		To:          []string{"u0cc55697597f61dd8b01cea8b027050e"},
		ToChannel:   1441331333,
		EventType:   "138311609100106403",
		Id:          "ABDDEF-22345678901",
		Content:     nil,
	}},
	contentContent: []interface{}{Operation{
		Params: [3]string{"u0f3bfc598b061eba02183bfc5280886a", "", ""},
		Rev:    2469,
		OTyp:   4,
	}},
}, {
	name: "Block friend operation",
	json: `{"result":[{
    "from":"uefb89606dd34df287b220e7b581d2466",
    "fromChannel":"1341333815",
    "to":["u0cc55697597f61dd8b01cea8b027050e"],
    "toChannel":1441331333,
    "eventType":"138311609100106403",
    "id":"ABDDFF-22345678991",
    "content": %s}]}`,
	content: []string{`{
      "params":["u0f3bfc599b061eba02183bfc5280886a", null, null],
      "revision":2470,
      "opType":8
    }`},
	receive: []Rx{Rx{
		From:        "uefb89606dd34df287b220e7b581d2466",
		FromChannel: "1341333815",
		To:          []string{"u0cc55697597f61dd8b01cea8b027050e"},
		ToChannel:   1441331333,
		EventType:   "138311609100106403",
		Id:          "ABDDFF-22345678991",
		Content:     nil,
	}},
	contentContent: []interface{}{Operation{
		Params: [3]string{"u0f3bfc599b061eba02183bfc5280886a", "", ""},
		Rev:    2470,
		OTyp:   8,
	}},
}}

func hlprArrayStrToArrayEmptyItf(as []string) []interface{} {
	itf := make([]interface{}, len(as))
	for i, _ := range as {
		itf[i] = as[i]
	}
	return itf
}

func hlprFatal(t *testing.T, i int, s string, err error) {
	t.Fatal("Test #", i+1, " ", s, err)
}

func hlprWriteError(w http.ResponseWriter, err error) {
	w.WriteHeader(http.StatusBadRequest)
	fmt.Fprint(w, "Encounter ERROR:", err)
}

func TestBOTServe(t *testing.T) {
	SetID("123456")
	SetSecret("789012")
	SetMID("ABCDEF")
	for i, d := range tstBOTSvrMsgDt {
		b1 := new(bytes.Buffer)
		b2 := new(bytes.Buffer)
		w := io.MultiWriter(b1, b2)
		fmt.Fprintf(w, d.json, hlprArrayStrToArrayEmptyItf(d.content)...)
		req, err := http.NewRequest("POST", "/test-line", b1)
		if err != nil {
			hlprFatal(t, i+1, d.name, err)
		}
		req.Header["Content-Type"] = []string{"application/json", "charset=UTF-8"}
		sha, err := genHMACSHA256(b2)
		if err != nil {
			hlprFatal(t, i+1, d.name, err)
		}
		req.Header["X-Line-Channelsignature"] = []string{sha}
		rr := httptest.NewRecorder()
		rr.Body = new(bytes.Buffer)
		handler := http.HandlerFunc(BOTServe) // cast to become HandlerFunc type
		handler.ServeHTTP(rr, req)
		if status := rr.Code; status != http.StatusOK {
			t.Errorf("Test #%d name: %s. `handler` returned wrong status code: got %v (%s) want %v",
				i+1, d.name, status, rr.Body.String(), http.StatusOK)
		}
	}
}

// Test sending helper functions
func hlprTestServer() (*httptest.Server, *func(*http.Request) error) {
	var f = func(*http.Request) error {
		return errors.New("line-test: not initialize")
	}
	ts := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := f(r); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	return ts, &f
}

func hlprError(t *testing.T, idx int, name string, err error) {
	t.Error("Test #"+strconv.Itoa(idx+1)+")", name, err)
}

func hlprDecSndMsgStruct(r io.Reader) (m *SndMsg, err error) {
	e := json.NewDecoder(r)
	m = new(SndMsg)
	if err = e.Decode(m); err != nil {
		return nil, err
	}
	return
}

func hlprGenSndMsgOp(to []string, e string) *SndMsg {
	return &SndMsg{
		To:        to,
		ToChannel: TO_Ch,
		EventType: e,
		Content:   nil,
	}
}

// Test SendText
var tstSendText = []struct {
	name string
	to   []string
	txt  string
}{{
	name: "Simple SendText test#1",
	to:   []string{"1234", "5678"},
	txt:  "Hello all.",
}, {
	name: "Simple SendText test#2",
	to:   []string{"9012", "3456"},
	txt:  "Thank you all.",
}}

func hlprDecSendText(r io.Reader) (m *SndMsg, t *Text) {
	var err error
	m, err = hlprDecSndMsgStruct(r)
	if err != nil {
		return nil, nil
	}
	t = new(Text)
	err = decodeRawJson(t, m.Content)
	if err != nil {
		return nil, nil
	}
	// for testing, we split the json.RawMessage and compare
	// instead of combine it in same struct to compare.
	m.Content = nil
	return
}

func hlprGenSendTextOp(to []string, txt string) (*SndMsg, *Text) {
	return &SndMsg{
			To:        to,
			ToChannel: TO_Ch,
			EventType: ET_MultiMedia,
			Content:   nil,
		}, &Text{
			Typ:   CT_Txt,
			ToTyp: RT_Usr,
			Txt:   txt,
		}
}

func TestSendText(t *testing.T) {
	ts, af := hlprTestServer()
	ts.StartTLS()
	defer ts.Close()
	botEvent = ts.URL
	for i, tt := range tstSendText {
		*af = func(r *http.Request) error {
			inM, inT := hlprDecSendText(r.Body)
			m, txt := hlprGenSendTextOp(tt.to, tt.txt)
			if !assert.Equal(t, inM, m) {
				err := errors.New("SndMsg object inconsistent")
				hlprError(t, i, tt.name, err)
				return err
			}
			if !assert.Equal(t, inT, txt) {
				err := errors.New("Text object inconsistent")
				hlprError(t, i, tt.name, err)
				return err
			}
			return nil
		}
		if err := SendText(tt.to, tt.txt); err != nil {
			hlprError(t, i, tt.name, err)
			continue
		}
	}
}

// Test SendImage
func hlprDecSendViewMedia(r io.Reader) (m *SndMsg, vm *ViewMedia) {
	var err error
	m, err = hlprDecSndMsgStruct(r)
	if err != nil {
		return nil, nil
	}
	vm = new(ViewMedia)
	err = decodeRawJson(vm, m.Content)
	if err != nil {
		return nil, nil
	}
	// for testing, we split the json.RawMessage and compare
	// instead of combine it in same struct to compare.
	m.Content = nil
	return
}

var tstSendImage = []struct {
	name      string
	to        []string
	curl, pre string
}{{
	name: "Fake SendImage test#1",
	to:   []string{"u5912407b444e54885d00111f7b0ce375"},
	curl: "http://example.com/original.jpg",
	pre:  "http://example.com/preview.jpg",
}, {
	name: "Fake SendImage test#2",
	to:   []string{"V89012", "V23456"},
	curl: "http://images.com/original.jpg",
	pre:  "http://images.com/preview.jpg",
}}

func hlprGenSendImageOp(to []string, curl, pre string) (*SndMsg, *ViewMedia) {
	return &SndMsg{
			To:        to,
			ToChannel: TO_Ch,
			EventType: ET_MultiMedia,
			Content:   nil,
		}, &ViewMedia{
			Typ:      CT_Image,
			ToTyp:    RT_Usr,
			Original: curl,
			Preview:  pre,
		}
}

func TestSendImage(t *testing.T) {
	ts, af := hlprTestServer()
	ts.StartTLS()
	defer ts.Close()
	botEvent = ts.URL
	for i, tt := range tstSendImage {
		*af = func(r *http.Request) error {
			inM, inVM := hlprDecSendViewMedia(r.Body)
			m, vm := hlprGenSendImageOp(tt.to, tt.curl, tt.pre)
			if !assert.Equal(t, inM, m) {
				err := errors.New("SndMsg object inconsistent")
				hlprError(t, i, tt.name, err)
				return err
			}
			if !assert.Equal(t, inVM, vm) {
				err := errors.New("ViewMedia object inconsistent")
				hlprError(t, i, tt.name, err)
				return err
			}
			return nil
		}
		if err := SendImage(tt.to, tt.curl, tt.pre); err != nil {
			hlprError(t, i, tt.name, err)
			continue
		}
	}
}

// Test SendVideo
var tstSendVideo = []struct {
	name      string
	to        []string
	curl, pre string
}{{
	name: "Fake SendVideo test#1",
	to:   []string{"A12345", "A56789"},
	curl: "http://example.com/original.mp4",
	pre:  "http://example.com/preview.jpg",
}, {
	name: "Fake SendVideo test#2",
	to:   []string{"V89012", "V23456"},
	curl: "http://test.com/original.mkv",
	pre:  "http://test.com/preview.jpg",
}}

func hlprGenSendVideoOp(to []string, curl, pre string) (*SndMsg, *ViewMedia) {
	return &SndMsg{
			To:        to,
			ToChannel: TO_Ch,
			EventType: ET_MultiMedia,
			Content:   nil,
		}, &ViewMedia{
			Typ:      CT_Video,
			ToTyp:    RT_Usr,
			Original: curl,
			Preview:  pre,
		}
}

func TestSendVideo(t *testing.T) {
	ts, af := hlprTestServer()
	ts.StartTLS()
	defer ts.Close()
	botEvent = ts.URL
	for i, tt := range tstSendVideo {
		*af = func(r *http.Request) error {
			inM, inVM := hlprDecSendViewMedia(r.Body)
			m, vm := hlprGenSendVideoOp(tt.to, tt.curl, tt.pre)
			if !assert.Equal(t, inM, m) {
				err := errors.New("SndMsg object inconsistent")
				hlprError(t, i, tt.name, err)
				return err
			}
			if !assert.Equal(t, inVM, vm) {
				err := errors.New("ViewMedia object inconsistent")
				hlprError(t, i, tt.name, err)
				return err
			}
			return nil
		}
		if err := SendVideo(tt.to, tt.curl, tt.pre); err != nil {
			hlprError(t, i, tt.name, err)
			continue
		}
	}
}

// Test SendAudio
var tstSendAudio = []struct {
	name   string
	to     []string
	url, d string
}{{
	name: "Fake SendAudio test#1",
	to:   []string{"12345", "56789"},
	url:  "http://listen.all",
	d:    "360000",
}, {
	name: "Fake SendAudio test#2",
	to:   []string{"89012", "23456"},
	url:  "http://thank.you.all/for/listening",
	d:    "240000",
}}

func hlprDecSendAudio(r io.Reader) (m *SndMsg, a *Audio) {
	var err error
	m, err = hlprDecSndMsgStruct(r)
	if err != nil {
		return nil, nil
	}
	a = new(Audio)
	err = decodeRawJson(a, m.Content)
	if err != nil {
		return nil, nil
	}
	// for testing, we split the json.RawMessage and compare
	// instead of combine it in same struct to compare.
	m.Content = nil
	return
}

func hlprGenSendAudioOp(to []string, url, d string) (*SndMsg, *Audio) {
	return &SndMsg{
			To:        to,
			ToChannel: TO_Ch,
			EventType: ET_MultiMedia,
			Content:   nil,
		}, &Audio{
			Typ:      CT_Audio,
			ToTyp:    RT_Usr,
			Original: url,
			Meta:     AudLen{Audlen: d},
		}
}

func TestSendAudio(t *testing.T) {
	ts, af := hlprTestServer()
	ts.StartTLS()
	defer ts.Close()
	botEvent = ts.URL
	for i, tt := range tstSendAudio {
		*af = func(r *http.Request) error {
			inM, inA := hlprDecSendAudio(r.Body)
			m, aud := hlprGenSendAudioOp(tt.to, tt.url, tt.d)
			if !assert.Equal(t, inM, m) {
				err := errors.New("SndMsg object inconsistent")
				hlprError(t, i, tt.name, err)
				return err
			}
			if !assert.Equal(t, inA, aud) {
				err := errors.New("Audio object inconsistent")
				hlprError(t, i, tt.name, err)
				return err
			}
			return nil
		}
		if err := SendAudio(tt.to, tt.url, tt.d); err != nil {
			hlprError(t, i, tt.name, err)
			continue
		}
	}
}

// Test SendLocation
var tstSendLocation = []struct {
	name      string
	to        []string
	title     string
	lat, long float32
}{{
	name:  "Fake SendLocation test#1",
	to:    []string{"012345", "456789"},
	title: "Kuala Lumpur",
	lat:   0.3456,
	long:  180.123,
}, {
	name:  "Fake SendLocation test#2",
	to:    []string{"A89012", "A23456"},
	title: "Selayang, Selangor",
	lat:   0.3401,
	long:  179.999,
}}

func hlprDecSendLocation(r io.Reader) (m *SndMsg, l *SndLocation) {
	var err error
	m, err = hlprDecSndMsgStruct(r)
	if err != nil {
		return nil, nil
	}
	l = new(SndLocation)
	err = decodeRawJson(l, m.Content)
	if err != nil {
		return nil, nil
	}
	// for testing, we split the json.RawMessage and compare
	// instead of combine it in same struct to compare.
	m.Content = nil
	return
}

func hlprGenSendLocationOp(to []string, title string, lat, long float32) (*SndMsg, *SndLocation) {
	return &SndMsg{
			To:        to,
			ToChannel: TO_Ch,
			EventType: ET_MultiMedia,
			Content:   nil,
		}, &SndLocation{
			Typ:   CT_Loc,
			ToTyp: RT_Usr,
			Txt:   title,
			Loc: Location{
				Title: title,
				Lat:   lat,
				Long:  long,
			}}
}

func TestSendLocation(t *testing.T) {
	ts, af := hlprTestServer()
	ts.StartTLS()
	defer ts.Close()
	botEvent = ts.URL
	for i, tt := range tstSendLocation {
		*af = func(r *http.Request) error {
			inM, inL := hlprDecSendLocation(r.Body)
			m, loc := hlprGenSendLocationOp(tt.to, tt.title, tt.lat, tt.long)
			if !assert.Equal(t, inM, m) {
				err := errors.New("SndMsg object inconsistent")
				hlprError(t, i, tt.name, err)
				return err
			}
			if !assert.Equal(t, inL, loc) {
				err := errors.New("SndLocation object inconsistent")
				hlprError(t, i, tt.name, err)
				return err
			}
			return nil
		}
		if err := SendLocation(tt.to, tt.title, tt.lat, tt.long); err != nil {
			hlprError(t, i, tt.name, err)
			continue
		}
	}
}

// Test SendSticker
var tstSendSticker = []struct {
	name         string
	to           []string
	id, pid, ver string
}{{
	name: "Fake Sticker test#1",
	to:   []string{"012345", "456789"},
	id:   "3", pid: "332", ver: "100",
}, {
	name: "Fake Sticker test#2",
	to:   []string{"A89012", "A23456"},
	id:   "4", pid: "316", ver: "101",
}}

func hlprDecSendSticker(r io.Reader) (m *SndMsg, s *SndSticker) {
	var err error
	m, err = hlprDecSndMsgStruct(r)
	if err != nil {
		return nil, nil
	}
	s = new(SndSticker)
	err = decodeRawJson(s, m.Content)
	if err != nil {
		return nil, nil
	}
	// for testing, we split the json.RawMessage and compare
	// instead of combine it in same struct to compare.
	m.Content = nil
	return
}

func hlprGenSendStickerOp(to []string, id, pid, ver string) (*SndMsg, *SndSticker) {
	return &SndMsg{
			To:        to,
			ToChannel: TO_Ch,
			EventType: ET_MultiMedia,
			Content:   nil,
		}, &SndSticker{
			Typ:   CT_Sticker,
			ToTyp: RT_Usr,
			Meta: Sticker{
				PId: pid,
				Id:  id,
				Ver: ver,
			}}
}

func TestSendSticker(t *testing.T) {
	ts, af := hlprTestServer()
	ts.StartTLS()
	defer ts.Close()
	botEvent = ts.URL
	for i, tt := range tstSendSticker {
		*af = func(r *http.Request) error {
			inM, inL := hlprDecSendSticker(r.Body)
			m, s := hlprGenSendStickerOp(tt.to, tt.id, tt.pid, tt.ver)
			if !assert.Equal(t, inM, m) {
				err := errors.New("SndMsg object inconsistent")
				hlprError(t, i, tt.name, err)
				return err
			}
			if !assert.Equal(t, inL, s) {
				err := errors.New("SndSticker object inconsistent")
				hlprError(t, i, tt.name, err)
				return err
			}
			return nil
		}
		if err := SendSticker(tt.to, tt.id, tt.pid, tt.ver); err != nil {
			hlprError(t, i, tt.name, err)
			continue
		}
	}
}

// Test MultiMsg
const (
	rtstMaxLen    = 10 // total each random test multi msgs should less than this value
	rtstMMMaxSize = 20 // max random size for To field and Msgs field
)

var rtstText = []*Text{
	&Text{Typ: CT_Txt, Txt: `json:"text1"`},
	&Text{Typ: CT_Txt, Txt: `json:"text2"`},
}

var rtstImage = []*ViewMedia{
	&ViewMedia{Typ: CT_Image, Original: `http://image1.com/img1`, Preview: "image1"},
	&ViewMedia{Typ: CT_Image, Original: `test image2"`, Preview: "http://preview1.com/image2"},
}

var rtstVideo = []*ViewMedia{
	&ViewMedia{Typ: CT_Video, Original: `http://video1.com/img1`, Preview: "video1"},
	&ViewMedia{Typ: CT_Video, Original: `test video2"`, Preview: "http://preview1.com/video2"},
}

var rtstAudio = []*Audio{
	&Audio{Typ: CT_Audio, Original: `http://video1.com/img1`, Meta: AudLen{Audlen: "video1"}},
	&Audio{Typ: CT_Audio, Original: `test video2"`, Meta: AudLen{Audlen: "http://preview1.com/video2"}},
}

var rtstLoc = []*SndLocation{
	&SndLocation{Typ: CT_Loc, Txt: `Selayang`, Loc: Location{Title: "Nothing", Lat: 44.5, Long: 100.1}},
	&SndLocation{Typ: CT_Loc, Txt: `Rawang`, Loc: Location{Title: "Ok", Lat: 66.78, Long: 12.678}},
}

var rtstSticker = []*SndSticker{
	&SndSticker{Typ: CT_Sticker, Meta: Sticker{PId: "123", Id: "55", Ver: "1.0.1", Txt: "Smile"}},
	&SndSticker{Typ: CT_Sticker, Meta: Sticker{PId: "999", Id: "77", Ver: "2.2.2", Txt: "Love"}},
}

var tstIdxToType = []int{
	CT_Txt,
	CT_Image,
	CT_Video,
	CT_Audio,
	CT_Loc,
	CT_Sticker,
}

var rtstToData = []string{
	"012345",
	"abc111",
	"poppsxz",
	"fjglkjre",
	"mnndsepp",
}

// helper func generate random JSONRaw object
func hlprMultiTstRawData(idx, rand int) JSONRaw {
	f := func(i, j int) int {
		if i >= j {
			return i % j
		}
		return i
	}
	idx = f(idx, len(tstIdxToType))
	switch idx {
	case CT_Txt:
		rand = f(rand, len(rtstText))
		return rtstText[rand]
	case CT_Image:
		rand = f(rand, len(rtstImage))
		return rtstImage[rand]
	case CT_Video:
		rand = f(rand, len(rtstVideo))
		return rtstVideo[rand]
	case CT_Audio:
		rand = f(rand, len(rtstAudio))
		return rtstAudio[rand]
	case CT_Loc:
		rand = f(rand, len(rtstLoc))
		return rtstLoc[rand]
	default:
		rand = f(rand, len(rtstSticker))
		return rtstSticker[rand]
	}
	return nil
}

// helper func generate random To field data.
func hlprRTos(r *rand.Rand) []string {
	as := make([]string, r.Intn(rtstMMMaxSize)+1)
	for i, _ := range as {
		as[i] = rtstToData[r.Intn(len(rtstToData))]
	}
	return as
}

func hlprDecMultiMsgs(r io.Reader) (m *SndMsg, mm *MultiMsgs) {
	var err error
	m, err = hlprDecSndMsgStruct(r)
	if err != nil {
		return nil, nil
	}
	mm = new(MultiMsgs)
	err = decodeRawJson(mm, m.Content)
	if err != nil {
		return nil, nil
	}
	// for testing, we split the json.RawMessage and compare
	// instead of combine it in same struct to compare.
	m.Content = nil
	return
}

func TestSendMultiMsgs(t *testing.T) {
	f := func(to []string, notified int, data []JSONRaw) bool {
		raws := make([]*json.RawMessage, len(data))
		for i, d := range data {
			raws[i] = d.ToRaw()
		}
		v := &MultiMsgs{
			Notified: notified,
			Msgs:     raws,
		}
		ts := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method != "POST" {
				w.WriteHeader(http.StatusNotFound)
				return
			}
			sm, mm := hlprDecMultiMsgs(r.Body)
			outSM := hlprGenSndMsgOp(to, ET_Msgs)
			if !assert.Equal(t, sm, outSM) {
				hlprWriteError(w, errors.New("SndMsg object inconsistent"))
				return
			}
			if !assert.Equal(t, v, mm) {
				hlprWriteError(w, errors.New("MultiMsgs object inconsistent"))
				return
			}
			w.WriteHeader(http.StatusOK)
		}))
		ts.StartTLS()
		botEvent = ts.URL
		if err := SendMultipleMessages(to, v); err != nil {
			ts.Close()
			t.Error("TestSendMultiMsgs error:", err)
			return false
		}
		ts.Close()
		return true
	}
	c := quick.Config{
		Rand: rand.New(rand.NewSource(time.Now().Unix())),
		Values: func(v []reflect.Value, r *rand.Rand) {
			raws := make([]JSONRaw, r.Intn(rtstMMMaxSize)+1)
			for i, _ := range raws {
				raws[i] = hlprMultiTstRawData(r.Intn(rtstMaxLen), r.Intn(rtstMaxLen))
			}
			v[2] = reflect.ValueOf(raws)
			v[1] = reflect.ValueOf(r.Intn(len(raws)))
			v[0] = reflect.ValueOf(hlprRTos(r))
		},
	}
	if err := quick.Check(f, &c); err != nil {
		t.Error("failed on TestSendMultiMsgs black box test", err)
	}
}

// Test GetMessage...
const msgContentFmt = `/v1/bot/message/%s/content`

var contentErr error = errors.New("returned content is not same as provided content")
var tstGetMsgContentData = []struct {
	name string
	id   string
	dt   []byte
}{{
	name: "Fake GetMessage data test#1",
	id:   "4567890",
	dt: []byte{0, 1, 3, 4, 5, 6, 7, 8, 9,
		10, 11, 12, 13, 14, 15, 16, 17, 18, 19},
}, {
	name: "Fake GetMessage data test#2",
	id:   "1456789",
	dt:   []byte{128, 129, 130, 140, 150, 160, 170},
}}

func TestGetMsgContent(t *testing.T) {
	for i, tt := range tstGetMsgContentData {
		path := fmt.Sprintf(msgContentFmt, tt.id)
		ts := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path != path || r.Method != "GET" {
				w.WriteHeader(http.StatusNotFound)
				return
			}
			w.WriteHeader(http.StatusOK)
			w.Write(tt.dt)
		}))
		ts.StartTLS()
		baseUrl = ts.URL
		r, err := GetMessageContentReader(tt.id)
		if err != nil {
			ts.Close()
			hlprError(t, i, tt.name, err)
			continue
		}
		b, err := ioutil.ReadAll(r)
		if err != nil {
			ts.Close()
			hlprError(t, i, tt.name, err)
			continue
		}
		if !assert.Equal(t, tt.dt, b) {
			ts.Close()
			hlprError(t, i, tt.name, contentErr)
			continue
		}
		ts.Close()
	}
}

const msgPreviewFmt = `/v1/bot/message/%s/content/preview`

var tstGetMsgPreviewData = []struct {
	name string
	id   string
	dt   []byte
}{{
	name: "Fake GetMessage data test#1",
	id:   "4567890",
	dt: []byte{0, 1, 3, 4, 5, 6, 7, 8, 9,
		10, 11, 12, 13, 14, 15, 16, 17, 18, 19},
}, {
	name: "Fake GetMessage data test#2",
	id:   "456789",
	dt:   []byte{128, 129, 130, 140, 150, 160, 170},
}}

func TestGetMsgPreview(t *testing.T) {
	for i, tt := range tstGetMsgPreviewData {
		path := fmt.Sprintf(msgPreviewFmt, tt.id)
		ts := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path != path || r.Method != "GET" {
				w.WriteHeader(http.StatusNotFound)
				return
			}
			w.WriteHeader(http.StatusOK)
			w.Write(tt.dt)
		}))
		ts.StartTLS()
		baseUrl = ts.URL
		r, err := GetMessagePreviewReader(tt.id)
		if err != nil {
			ts.Close()
			hlprError(t, i, tt.name, err)
			continue
		}
		b, err := ioutil.ReadAll(r)
		if err != nil {
			ts.Close()
			hlprError(t, i, tt.name, err)
			continue
		}
		if !assert.Equal(t, tt.dt, b) {
			ts.Close()
			hlprError(t, i, tt.name, contentErr)
			continue
		}
		ts.Close()
	}
}
