package akismet

import (
	"errors"
	"io/ioutil"
	"net/http"
	"net/url"
)

const (
	BaseURL = "rest.akismet.com/1.1/"
)

var (
	ErrNoBlog = errors.New("No blog name given")
	ErrNoIP   = errors.New("No user IP given")
)

// TODO: User-Agent generation

// Verify API key
func VerifyKey(apiKey, blogURL string) (ok bool, err error) {
	var v url.Values
	v.Set("key", apiKey)
	v.Set("blog", blogURL)

	resp, err := http.PostForm(BaseURL+"verify-key", v)
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return false, err
	}
	return string(body) == "valid", nil
}

type Options struct {
	Blog        string // Required
	UserIP      string // Required
	Referrer    string
	Permalink   string
	Type        string
	Author      string
	AuthorEmail string
	AuthorURL   string
	Content     string
}

func (self *Options) values() (v url.Values, err error) {
	if len(self.Blog) == 0 {
		err = ErrNoBlog
		return
	}
	if len(self.UserIP) == 0 {
		err = ErrNoIP
		return
	}
	v.Set("blog", self.Blog)
	v.Set("user_ip", self.UserIP)
	if len(self.Referrer) > 0 {
		v.Set("referrer", self.Referrer)
	}
	if len(self.Permalink) > 0 {
		v.Set("permalink", self.Permalink)
	}
	if len(self.Type) > 0 {
		v.Set("comment_type", self.Type)
	}
	if len(self.Author) > 0 {
		v.Set("comment_author", self.Author)
	}
	if len(self.AuthorEmail) > 0 {
		v.Set("comment_author_email", self.AuthorEmail)
	}
	if len(self.AuthorURL) > 0 {
		v.Set("comment_author_url", self.AuthorURL)
	}
	if len(self.Content) > 0 {
		v.Set("comment_content", self.Content)
	}

	return
}

// Submit comment for spam check. Returns a true if comment is spam, or an error if one occurred.
func CommentCheck(apiKey string, opts *Options) (spam bool, err error) {
	path := apiKey + "." + BaseURL + "comment-check"
	v, err := opts.values()
	if err != nil {
		return false, err
	}

	resp, err := http.PostForm(path, v)
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return false, err
	}
	switch string(body) {
	case "true":
		return true, nil
	case "invalid":
		return false, errors.New(resp.Header.Get("X-Akismet-Debug-Help"))
	}
	return false, nil
}

// Let Akismet know they missed something.
func SubmitSpam(apiKey string, opts *Options) (ok bool, err error) {
	return submit(apiKey, opts, "spam")
}

// Let Akismet know this comment triggered a false positive.
func SubmitHam(apiKey string, opts *Options) (ok bool, err error) {
	return submit(apiKey, opts, "ham")
}

func submit(apiKey string, opts *Options, hamOrSpam string) (ok bool, err error) {
	v, err := opts.values()
	if err != nil {
		return false, err
	}
	path := apiKey + "." + BaseURL + "submit-" + hamOrSpam
	resp, err := http.PostForm(path, v)
	if err != nil {
		return false, err
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return false, err
	}
	return string(body) == "Thanks for making the web a better place.", nil
}
