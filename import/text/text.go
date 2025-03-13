// Package text provides an Importer that converts plain text into a Bluesky post.
//
// The importer looks for URLs, usernames, and hashtags in the text, and converts them into
// links, mentions, and tags. Some formatting is applied to the URLs to avoid displaying
// overlong URLs, and extra validation can be added optionally.
//
// Note that you need to add a HandleResolver to be able to link usernames to their
// profiles. You may want to use the ClientHandleResolver function in package client.
package text

import (
	"net"
	"net/url"
	"regexp"
	"slices"
	"strings"

	"github.com/jtarrio/k3"
)

// Importer is an interface to convert some plain text into a Bluesky post.
type Importer interface {
	// Import converts the given text string into a post.
	Import(text string) *k3.Post
}

// NewImporter creates a new Importer with the given options.
func NewImporter(options ...ImporterOption) Importer {
	i := &importer{
		handleResolver: DefaultHandleResolver,
		urlResolver:    DefaultUrlResolver,
		urlFormatter:   DefaultUrlFormatter,
	}
	for _, option := range options {
		option(i)
	}
	return i
}

// WithHandleResolver sets the function to use to resolve Bluesky handler.
//
// By default, handles are not resolved and, therefore, they are not converted into links to the appropriate profile.
func WithHandleResolver(r HandleResolver) ImporterOption {
	return func(i *importer) {
		i.handleResolver = r
	}
}

// WithUrlResolver sets the function to use to resolve URLs.
//
// By default, URL-shaped strings are turned into URLs. You can use the NetworkUrlResolver to only convert URLs whose hostname exists.
func WithUrlResolver(r UrlResolver) ImporterOption {
	return func(i *importer) {
		i.urlResolver = r
	}
}

// WithUrlFormatter sets the function to use to format URLs in the post text.
//
// By default, the first 20 characters of the hostname+path are shown, adding an ellipsis (…) if necessary.
// However, if the hostname itself is longer than 20 characters, the whole hostname is shown.
func WithUrlFormatter(f UrlFormatter) ImporterOption {
	return func(i *importer) {
		i.urlFormatter = f
	}
}

// HandleResolver is a type for a function that takes a Bluesky handle and returns its DID, if it exists, or nil if it doesn't.
type HandleResolver func(string) *string

// UrlResolver is a type for a function that takes a string (with or without an initial `http://` or `https://`) and parses it as an absolute URL, returning nil if it's not possible to parse.
type UrlResolver func(string) *url.URL

// UrlFormatter is a type for a function that takes a URL and returns the string to show in the post's text.
type UrlFormatter func(url *url.URL) string

type ImporterOption func(*importer)

type importer struct {
	handleResolver HandleResolver
	urlResolver    UrlResolver
	urlFormatter   UrlFormatter
}

func (i *importer) Import(text string) *k3.Post {
	var found []found
	found = append(found, findAll(text, usernameRe, username)...)
	found = append(found, findAll(text, hashtagRe, tag)...)
	found = append(found, findAll(text, fullUrlRe, webUrl)...)
	found = append(found, findAll(text, shortUrlRe, webUrl)...)
	sortFound(found)

	out := k3.NewPost()
	p := 0
	for _, f := range found {
		if p > f.start {
			continue
		}
		switch f.what {
		case webUrl:
			url := i.urlResolver(text[f.start:f.end])
			if url != nil {
				out.AddText(text[p:f.start])
				out.AddLink(i.urlFormatter(url), url.String())
				p = f.end
			}
		case username:
			handle := text[f.start:f.end]
			did := i.handleResolver(handle[1:])
			if did != nil {
				out.AddText(text[p:f.start])
				out.AddMention(handle, *did)
				p = f.end
			}
		case tag:
			tag := text[f.start:f.end]
			if len(tag) > 1 {
				out.AddText(text[p:f.start])
				out.AddTag(tag, tag[1:])
				p = f.end
			}
		}
	}
	if p < len(text) {
		out.AddText(text[p:])
	}
	return out
}

func isPunctuation(chr byte) bool {
	return chr == '.' || chr == ',' || chr == ')' || chr == '?' || chr == '!'
}

func findAll(str string, re *regexp.Regexp, what foundType) []found {
	var out []found
	for _, linkMatch := range re.FindAllStringIndex(str, -1) {
		start := linkMatch[0]
		end := linkMatch[1]
		if what == webUrl {
			if start > 0 && str[start-1] == '@' || str[start-1] == '#' {
				continue
			}
			for end > start && isPunctuation(str[end-1]) {
				end--
			}
			if start == end {
				continue
			}
		}
		out = append(out, found{start: start, end: end, what: what})
	}
	return out
}

// Does not accept the username part of the URL.
var fullUrlRe = regexp.MustCompile(
	`(https?://)` + // scheme
		`(\[[0-9A-Fa-f:]+\]|([0-9]+\.){3}[0-9]+|[A-Za-z][A-Za-z0-9._-]*)` + // host
		`(:[0-9]+)?` + // port
		`((/[A-Za-z0-9._~%!$&'()*+,;=-]*)+)?` + // path
		`(\?[A-Za-z0-9._~%!$&'()*+,;=/?-]*)?` + // query
		`(#[A-Za-z0-9._~%!$&'()*+,;=/?-]*)?`, // fragment
)

// Does not require http:// or https://, but it requires a hostname with at least two components or a full IPv4.
var shortUrlRe = regexp.MustCompile(
	`(\[[0-9A-Fa-f:]+\]|([0-9]+\.){3}[0-9]+|[A-Za-z][A-Za-z0-9._-]*\.[A-Za-z0-9_-]+)` + // host
		`(:[0-9]+)?` + // port
		`((/[A-Za-z0-9._~%!$&'()*+,;=-]*)+)?` + // path
		`(\?[A-Za-z0-9._~%!$&'()*+,;=/?-]*)?` + // query
		`(#[A-Za-z0-9._~%!$&'()*+,;=/?-]*)?`, // fragment
)

var usernameRe = regexp.MustCompile(
	`@[A-Za-z][A-Za-z0-9._-]*\.[A-Za-z0-9_-]+`,
)

var hashtagRe = regexp.MustCompile(
	`#[A-Za-z#][A-Za-z0-9._#-]*`,
)

func sortFound(f []found) {
	slices.SortFunc(f, func(a, b found) int {
		if a.start != b.start {
			return a.start - b.start
		}
		if a.end != b.end {
			return a.end - b.end
		}
		return int(a.what) - int(b.what)
	})
}

type foundType int

const (
	webUrl foundType = iota
	username
	tag
)

type found struct {
	start, end int
	what       foundType
}

func DefaultHandleResolver(h string) *string {
	return nil
}

func DefaultUrlResolver(u string) *url.URL {
	n := strings.Index(u, "://")
	if n < 0 {
		u = "https://" + u
	}
	parsed, err := url.Parse(u)
	if err != nil {
		return nil
	}
	return parsed
}

func NetworkUrlResolver(u string) *url.URL {
	parsed := DefaultUrlResolver(u)
	ips, err := net.LookupIP(parsed.Hostname())
	if err != nil || len(ips) == 0 {
		return nil
	}
	return parsed
}

func DefaultUrlFormatter(u *url.URL) string {
	host := u.Host
	path := u.Path
	if len(host) >= 20 || len(path) == 0 || path == "/" {
		return host
	}
	hostPath := host + path
	if len(hostPath) >= 24 {
		return hostPath[:20] + "…"
	}
	return hostPath
}
