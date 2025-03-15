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
	"unicode"

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
		tagResolver:    DefaultTagResolver,
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

// WithTagResolver sets the function to use to resolve hashtags.
//
// By default, strings starting with # (optionally ending with #) and followed by non-whitespace characters (unless it's entirely composed of numbers) is considered a hash tag.
func WithTagResolver(r TagResolver) ImporterOption {
	return func(i *importer) {
		i.tagResolver = r
	}
}

// HandleResolver is a type for a function that takes a Bluesky handle and returns its DID, if it exists, or nil if it doesn't.
type HandleResolver func(string) *string

// UrlResolver is a type for a function that takes a string (with or without an initial `http://` or `https://`) and parses it as an absolute URL, returning nil if it's not possible to parse.
type UrlResolver func(string) *url.URL

// UrlFormatter is a type for a function that takes a URL and returns the string to show in the post's text.
type UrlFormatter func(*url.URL) string

// TagResolver is a type for a function that takes a string (with initial '#') and returns the tag it corresponds to, or nil if none.
type TagResolver func(string) *string

type ImporterOption func(*importer)

type importer struct {
	handleResolver HandleResolver
	urlResolver    UrlResolver
	urlFormatter   UrlFormatter
	tagResolver    TagResolver
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
			hashtag := text[f.start:f.end]
			tag := i.tagResolver(hashtag)
			if tag != nil {
				out.AddText(text[p:f.start])
				out.AddTag(hashtag, *tag)
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
		for end > start && isPunctuation(str[end-1]) {
			end--
		}
		if start == end {
			continue
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
	`#[^\s#]+#?`,
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

// DefaultHandleResolver recognizes no handles, so they don't get turned into links.
func DefaultHandleResolver(h string) *string {
	return nil
}

// DefaultUrlResolver returns every string that can be parsed as a URL, with or without an 'http(s)://' prefix.
func DefaultUrlResolver(u string) *url.URL {
	n := strings.Index(u, "://")
	if n < 0 {
		u = "https://" + u
	}
	parsed, err := url.Parse(u)
	if err != nil {
		return nil
	}
	if !parsed.IsAbs() {
		return nil
	}
	return parsed
}

// NetworkUrlResolver parses the URL like DefaultUrlResolver, but then checks that the hostname has an IP address.
func NetworkUrlResolver(u string) *url.URL {
	parsed := DefaultUrlResolver(u)
	ips, err := net.LookupIP(parsed.Hostname())
	if err != nil || len(ips) == 0 {
		return nil
	}
	return parsed
}

// DefaultUrlFormatter returns the url without scheme, cut to 20 characters if it's longer than 24.
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

// DefaultTagResolver cuts the initial '#' and final '#', if it exists, and returns the rest unless it's entirely composed of numbers.
func DefaultTagResolver(hashtag string) *string {
	hashtag, _ = strings.CutPrefix(hashtag, "#")
	hashtag, _ = strings.CutSuffix(hashtag, "#")
	for _, r := range hashtag {
		if !unicode.IsNumber(r) {
			return &hashtag
		}
	}
	return nil
}

// NoTagResolver always returns nil so no hashtags are turned into links.
func NoTagResolver(hashtag string) *string {
	return nil
}
