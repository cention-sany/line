package line

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"

	"golang.org/x/oauth2"
)

func (a *API) GetUserProfile() (*Profile, error) {
	var p Profile
	req, err := http.NewRequest("GET", fmt.Sprint(baseUrl, "/profile"), nil)
	if err != nil {
		return nil, err
	}
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

func getProfiles(a *API, start, display int, path string) (*Profiles, error) {
	var p Profiles
	v := url.Values{}
	v.Set(`start`, strconv.Itoa(start))
	v.Set(`display`, strconv.Itoa(display))
	u, err := url.Parse(fmt.Sprint(baseUrl, path))
	if err != nil {
		return nil, err
	}
	u.RawQuery = v.Encode()
	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		return nil, err
	}
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

func (a *API) GetFriendLists(start, display int) (*Profiles, error) {
	return getProfiles(a, start, display, "/friends")
}

// GetFavLists get favourites lists of the authorized/authenticaed user which
// should be subset of her/his friends list.
func (a *API) GetFavLists(start, display int) (*Profiles, error) {
	return getProfiles(a, start, display, "/friends/favorite")
}

// GetGroupMembers get group member lists. mid is identifier of the group.
func (a *API) GetGroupMembers(start, display int, mid string) (*Profiles, error) {
	return getProfiles(a, start, display, fmt.Sprint("/group/", mid, "/members"))
}

// GetGroups get group lists of any member ID (mid).
func (a *API) GetGroups(mids ...string) (*Groups, error) {
	var g Groups
	if mids == nil || len(mids) == 0 {
		return nil, errors.New("require MID(s)")
	}
	v := url.Values{}
	for _, mid := range mids {
		v.Add("mids", mid)
	}
	u, err := url.Parse(fmt.Sprint(baseUrl, "/groups"))
	if err != nil {
		return nil, err
	}
	u.RawQuery = v.Encode()
	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		return nil, err
	}
	c := a.Config.Client(oauth2.NoContext, a.Token)
	resp, err := c.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	d := json.NewDecoder(resp.Body)
	err = d.Decode(&g)
	if err != nil {
		return nil, err
	}
	return &g, nil
}
