package line

import (
	"encoding/json"
	"errors"
	"fmt"
	"html"
	"strconv"
	"strings"
)

const (
	targetLimit = 150
	// to channel
	TO_Ch = 1383378250

	// content-type
	CT_Txt      = 1
	CT_Image    = 2
	CT_Video    = 3
	CT_Audio    = 4
	CT_Loc      = 7
	CT_Sticker  = 8
	CT_Contact  = 10
	CT_Rich_Msg = 12

	// recipient-type
	RT_Usr = 1

	// event-type
	ET_MultiMedia = "138311608800106203"
	ET_OP_MSG     = "138311609000106303"
	ET_OP_ADD     = "138311609100106403"
	ET_Msgs       = "140177271400161403"

	// operation-type
	OT_ADD_FRIEND = 4
	OT_BLOCK      = 8
)

type Profile struct {
	DisplayName   string `json:"displayName"`
	Mid           string `json:"mid"`
	PictureUrl    string `json:"pictureUrl"`
	StatusMessage string `json:"statusMessage"`
}

type Profiles struct {
	Contacts *Profile `json:"contacts"`
	Count    int      `json:"count"`
	Total    int      `json:"total"`
	Start    int      `json:"start"`
	Display  int      `json:"display"`
}

type Group struct {
	Mid        string `json:"mid"`
	Name       string `json:"name"`
	PictureUrl string `json:"pictureUrl"`
}

type Groups struct {
	Groups  *Group `json:"groups"`
	Count   int    `json:"count"`
	Total   int    `json:"total"`
	Start   int    `json:"start"`
	Display int    `json:"display"`
}

type Placeholder map[string]string
type EscapePlaceholder map[string]string

func (e EscapePlaceholder) MarshalerJSON() ([]byte, error) {
	return placeholderMarshaler(e, true)
}

func (p Placeholder) MarshalerJSON() ([]byte, error) {
	return placeholderMarshaler(p, false)
}

func placeholderMarshaler(p map[string]string, htmlEscape bool) ([]byte, error) {
	if p == nil || len(p) == 0 {
		return nil, errors.New("line: can not marshal empty map")
	}
	a := make([]string, len(p))
	i := 0
	for k, v := range p {
		v = strconv.Quote(v)
		if htmlEscape {
			v = html.EscapeString(v)
		}
		a[i] = fmt.Sprint(strconv.Quote(k), ":", v)
		i++
	}
	s := strings.Join(a, `,`)
	return []byte(fmt.Sprint("{", s, "}")), nil
}

type LinkMsgContent struct {
	TemplateId     string            `json:"templateId"`
	PreviewUrl     string            `json:"previewUrl,omitempty"`
	TextParams     EscapePlaceholder `json:"textParams,omitempty"`
	SubTextParams  Placeholder       `json:"subTextParams,omitempty"`
	AltTextParams  Placeholder       `json:"altTextParams,omitempty"`
	LinkTextParams Placeholder       `json:"linkTextParams,omitempty"`
	ALinkUriParams Placeholder       `json:"aLinkTextParams,omitempty"`
	ILinkUriParams Placeholder       `json:"iLinkUriParams,omitempty"`
	LinkUriParams  Placeholder       `json:"linkUriParams,omitempty"`
}

type SndMsg struct {
	To        []string         `json:"to"`
	ToChannel int              `json:"toChannel"`
	EventType string           `json:"eventType"`
	Content   *json.RawMessage `json:"content"`
}

type PostResponse struct {
	Version   string   `json:"version"`
	Timestamp int64    `json:"timestamp"`
	MessageId string   `json:"messageId"`
	Failed    []string `json:"failed"`
}

type TimelineTemplate struct {
	DyamicObjs Placeholder `json:"dyamicObjs,omitempty"`
	FriendMids []string    `json:"friendMids,omitempty"`
	TitleText  string      `json:"titleText,omitempty"`
	MainText   string      `json:"mainText,omitempty"`
	SubText    string      `json:"subText,omitempty"`
}

type TimelineImage struct {
	Url    string `json:"url"`
	Width  int    `json:"width"`
	Height int    `json:"height"`
}

type TimelineUrl struct {
	Device    string `json:"device"`
	TargetUrl string `json:"targetUrl"`
}

type TimelineContent struct {
	ApiVer    int               `json:"apiVer"`
	Cmd       string            `json:"cmd"`
	UserMid   string            `json:"userMid"`
	Device    string            `json:"device"`
	Region    string            `json:"region"`
	ChannelID int               `json:"channelID"`
	FeedNo    int               `json:"feedNo"`
	Test      bool              `json:"test,omitempty"`
	PostText  string            `json:"postText"`
	Template  *TimelineTemplate `json:"template,omitempty"`
	Thumbnail *TimelineImage    `json:"thumbnail,omitempty"`
	Url       []*TimelineUrl    `json:"url,omitempty"`
}

type TimelineMsg struct {
	To        []string         `json:"to"`
	ToChannel int              `json:"toChannel"`
	EventType string           `json:"eventType"`
	Content   *TimelineContent `json:"content"`
}

// BOT API
type Rx struct {
	From        string           `json:"from"`
	FromChannel string           `json:"fromChannel"`
	To          []string         `json:"to"`
	ToChannel   int              `json:"toChannel"`
	EventType   string           `json:"eventType"`
	Id          string           `json:"id"`
	Content     *json.RawMessage `json:"content"`
}

type Message struct {
	Loc   *Location        `json:"location"`
	Id    string           `json:"id"`
	Typ   int              `json:"contentType"`
	Frm   string           `json:"from"`
	CTime int64            `json:"createdTime"`
	To    []string         `json:"to"`
	ToTyp int              `json:"toType"`
	Meta  *json.RawMessage `json:"contentMetadata"`
	Txt   string           `json:"text"`
}

func (m *Message) ParseMeta() (interface{}, error) {
	switch m.Typ {
	case CT_Sticker:
		return m.ParseSticker()
	case CT_Contact:
		return m.ParseContact()
	}
	return nil, errors.New("line: unknown content type!")
}

func (m *Message) ParseSticker() (*Sticker, error) {
	if m.Typ != CT_Sticker {
		return nil, errors.New("line: not Sticker content type!")
	}
	var stick Sticker
	if err := decodeRawJson(&stick, m.Meta); err != nil {
		return nil, err
	}
	return &stick, nil
}

func (m *Message) ParseContact() (*Contact, error) {
	if m.Typ != CT_Contact {
		return nil, errors.New("line: not Contact content type!")
	}
	var contact Contact
	if err := decodeRawJson(&contact, m.Meta); err != nil {
		return nil, err
	}
	return &contact, nil
}

type Location struct {
	Title string  `json:"title"`
	Addr  string  `json:"address,omitempty"`
	Lat   float32 `json:"latitude"`
	Long  float32 `json:"longitude"`
}

type Sticker struct {
	PId string `json:"STKPKGID"`
	Id  string `json:"STKID"`
	Ver string `json:"STKVER"`
	Txt string `json:"STKTXT,omitempty"`
}

type Contact struct {
	Mid  string `json:"mid"`
	Name string `json:"displayName"`
}

type Operation struct {
	Params [3]string `json:"params"`
	Rev    int       `json:"revision"`
	OTyp   int       `json:"opType"`
}

type JSONRaw interface {
	ToRaw() *json.RawMessage
}

// use in conjunction with SendMsg
type Text struct {
	Typ   int    `json:"contentType"`
	ToTyp int    `json:"toType,omitempty"`
	Txt   string `json:"text"`
}

// func (t *Text) MarshalJSON() ([]byte, error) {
// 	return json.Marshal(t)
// }

func (t *Text) ToRaw() *json.RawMessage {
	b, err := json.Marshal(t)
	if err != nil {
		panic(err)
	}
	raw := json.RawMessage(b)
	return &raw
}

// Image (2) or Video (3) type
type ViewMedia struct {
	Typ      int    `json:"contentType"`
	ToTyp    int    `json:"toType,omitempty"`
	Original string `json:"originalContentUrl"`
	Preview  string `json:"previewImageUrl"`
}

// func (vm *ViewMedia) MarshalJSON() ([]byte, error) {
// 	return json.Marshal(vm)
// }

func (vm *ViewMedia) ToRaw() *json.RawMessage {
	b, err := json.Marshal(vm)
	if err != nil {
		panic(err)
	}
	raw := json.RawMessage(b)
	return &raw
}

type Audio struct {
	Typ      int    `json:"contentType"`
	ToTyp    int    `json:"toType,omitempty"`
	Original string `json:"originalContentUrl"`
	Meta     AudLen `json:"contentMetadata"`
}

// func (a *Audio) MarshalJSON() ([]byte, error) {
// 	return json.Marshal(a)
// }

func (a *Audio) ToRaw() *json.RawMessage {
	b, err := json.Marshal(a)
	if err != nil {
		panic(err)
	}
	raw := json.RawMessage(b)
	return &raw
}

type AudLen struct {
	Audlen string `json:"AUDLEN"`
}

type SndLocation struct {
	Typ   int      `json:"contentType"`
	ToTyp int      `json:"toType,omitempty"`
	Txt   string   `json:"text"`
	Loc   Location `json:"location"`
}

// func (sl *SndLocation) MarshalJSON() ([]byte, error) {
// 	return json.Marshal(sl)
// }

func (sl *SndLocation) ToRaw() *json.RawMessage {
	b, err := json.Marshal(sl)
	if err != nil {
		panic(err)
	}
	raw := json.RawMessage(b)
	return &raw
}

type SndSticker struct {
	Typ   int     `json:"contentType"`
	ToTyp int     `json:"toType,omitempty"`
	Meta  Sticker `json:"contentMetadata"`
}

// func (ss *SndSticker) MarshalJSON() ([]byte, error) {
// 	return json.Marshal(ss)
// }

func (ss *SndSticker) ToRaw() *json.RawMessage {
	b, err := json.Marshal(ss)
	if err != nil {
		panic(err)
	}
	raw := json.RawMessage(b)
	return &raw
}

type MultiMsgs struct {
	Notified int                `json:"messageNotified"`
	Msgs     []*json.RawMessage `json:"messages"`
}

type RichMsg struct {
	Canvas struct {
		W     int    `json:"width"`
		H     int    `json:"height"`
		Scene string `json:"initialScene"`
	} `json:"canvas"`
	Images struct {
		X int `json:"x"`
		Y int `json:"y"`
		W int `json:"w"`
		H int `json:"h"`
	} `json:"image1"`
	Actions struct {
		Open RichAct `json:"openHomepage"`
		Show RichAct `json:"showItem"`
	} `json:"actions"`
	Scenes struct {
		S1 struct {
			Draws struct {
				Imgs []RichImage
			} `json:"draws"`
			Listen struct {
				Lsts []RichListen
			} `json:"listeners"`
		} `json:"scene1"`
	} `json:"scenes"`
}

type RichAct struct {
	Typ  string `json:"type"`
	Txt  string `json:"text"`
	Prms struct {
		LUri string `json:"linkUri"`
	} `json:"params"`
}

type RichImage struct {
	I string `json:"image"` //I always equal "image1"
	X int    `json:"x"`
	Y int    `json:"y"`
	W int    `json:"w"`
	H int    `json:"h"`
}

type RichListen struct {
	Typ  int     `json:"type"`
	Prms [4]int  `json:"params"`
	Act  Sticker `json:"action"`
}

type RichMsgContent struct {
	Typ   int `json:"contentType"`
	ToTyp int `json:"toType"`
	Meta  struct {
		Url string `json:"DOWNLOAD_URL"`
		Rev string `json:"SPEC_REV"`
		Alt string `json:"ALT_TEXT"`
		Jsn string `json:"MARKUP_JSON,string"`
	} `json:"contentMetadata"`
}

type Result struct {
	R []Rx `json:"result"`
}

type ResultOk struct {
	R string `json:"result"`
}
