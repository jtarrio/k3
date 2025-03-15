package text_test

import (
	"testing"

	"github.com/jtarrio/k3"
	"github.com/jtarrio/k3/import/text"
	"github.com/stretchr/testify/assert"
)

func TestConvertPlainText(t *testing.T) {
	post := text.NewImporter().Import(`This is a simple string`)
	expected := k3.NewPost().AddText(`This is a simple string`)
	assert.Equal(t, expected, post)
}

func TestConvertFullUrls(t *testing.T) {
	post := text.NewImporter().Import(`
Hostname https://example.com
Hostname slash https://example.com/
Path http://example.net/foo/bar
Port http://example:8080/
Query http://example?abc=def
Fragment http://example#abcdef
Pqf http://example/foo?abc=def&def=ghi#jklmno
IPv6 http://[2001::1]:8080/
IPv4 http://127.0.0.1:8080/
Escapes http://example.net/ab%63?abc=d%20ef#ghi%20jkl
Final punctuation http://example.com/abc?foo=bar.,!?)more.,!?)
The end`)
	expected := k3.NewPost().AddText(`
Hostname `).AddLink(`example.com`, "https://example.com").AddText(`
Hostname slash `).AddLink(`example.com`, "https://example.com/").AddText(`
Path `).AddLink(`example.net/foo/bar`, "http://example.net/foo/bar").AddText(`
Port `).AddLink(`example:8080`, "http://example:8080/").AddText(`
Query `).AddLink(`example`, "http://example?abc=def").AddText(`
Fragment `).AddLink(`example`, "http://example#abcdef").AddText(`
Pqf `).AddLink(`example/foo`, "http://example/foo?abc=def&def=ghi#jklmno").AddText(`
IPv6 `).AddLink(`[2001::1]:8080`, "http://[2001::1]:8080/").AddText(`
IPv4 `).AddLink(`127.0.0.1:8080`, "http://127.0.0.1:8080/").AddText(`
Escapes `).AddLink(`example.net/abc`, "http://example.net/ab%63?abc=d%20ef#ghi%20jkl").AddText(`
Final punctuation `).AddLink(`example.com/abc`, "http://example.com/abc?foo=bar.,!?)more").AddText(`.,!?)
The end`)
	assert.Equal(t, expected, post)
}

func TestConvertShortUrls(t *testing.T) {
	post := text.NewImporter().Import(`
Hostname example.com
Hostname slash example.com/
Path example.net/foo/bar
Full name required example/foo/bar
Port example.com:8080
Query example.com?abc=def
Fragment example.com#abcdef
Pqf example.com/foo?abc=def&def=ghi#jklmno
IPv6 [2001::1]:8080
IPv4 127.0.0.1:8080
Not any random number 123.456.789:8080
Escapes example.net/ab%63?abc=d%20ef#ghi%20jkl
Final punctuation example.com/abc?foo=bar.,!?)more.,!?)
The end`)
	expected := k3.NewPost().AddText(`
Hostname `).AddLink(`example.com`, "https://example.com").AddText(`
Hostname slash `).AddLink(`example.com`, "https://example.com/").AddText(`
Path `).AddLink(`example.net/foo/bar`, "https://example.net/foo/bar").AddText(`
Full name required example/foo/bar
Port `).AddLink(`example.com:8080`, "https://example.com:8080").AddText(`
Query `).AddLink(`example.com`, "https://example.com?abc=def").AddText(`
Fragment `).AddLink(`example.com`, "https://example.com#abcdef").AddText(`
Pqf `).AddLink(`example.com/foo`, "https://example.com/foo?abc=def&def=ghi#jklmno").AddText(`
IPv6 `).AddLink(`[2001::1]:8080`, "https://[2001::1]:8080").AddText(`
IPv4 `).AddLink(`127.0.0.1:8080`, "https://127.0.0.1:8080").AddText(`
Not any random number 123.456.789:8080
Escapes `).AddLink(`example.net/abc`, "https://example.net/ab%63?abc=d%20ef#ghi%20jkl").AddText(`
Final punctuation `).AddLink(`example.com/abc`, "https://example.com/abc?foo=bar.,!?)more").AddText(`.,!?)
The end`)
	assert.Equal(t, expected, post)
}

func TestConvertUsernames(t *testing.T) {
	// Use a fake resolver where usernames starting with 'v' are valid
	fakeResolver := func(u string) *string {
		if u[0] == 'v' {
			cid := "cid:web:" + u
			return &cid
		}
		return nil
	}
	post := text.NewImporter(text.WithHandleResolver(fakeResolver)).Import(`
An invalid @username
A @valid.username
This is @invalid.url.shaped
Something else`)
	expected := k3.NewPost().AddText(`
An invalid @username
A `).AddMention(`@valid.username`, `cid:web:valid.username`).AddText(`
This is @`).AddLink(`invalid.url.shaped`, `https://invalid.url.shaped`).AddText(`
Something else`)
	assert.Equal(t, expected, post)
}

func TestIgnoreUsernamesByDefault(t *testing.T) {
	post := text.NewImporter().Import(`
It doesn't matter whether the @username
is @valid.username or @invalid.username
it is not converted as a username.`)
	expected := k3.NewPost().AddText(`
It doesn't matter whether the @username
is @`).AddLink(`valid.username`, `https://valid.username`).AddText(` or @`).AddLink(`invalid.username`, `https://invalid.username`).AddText(`
it is not converted as a username.`)
	assert.Equal(t, expected, post)
}

func TestConvertTags(t *testing.T) {
	post := text.NewImporter().Import(`
Some #hashtag
A#doubleHashtag#isAllowed
You can have #123numbers but not only #123 numbers
You can also have #🙂emoji and #punc.tuation
Last line`)
	expected := k3.NewPost().AddText(`
Some `).AddTag(`#hashtag`, `hashtag`).AddText(`
A`).AddTag(`#doubleHashtag#`, `doubleHashtag`).AddText(`isAllowed
You can have `).AddTag(`#123numbers`, `123numbers`).AddText(` but not only #123 numbers
You can also have `).AddTag(`#🙂emoji`, `🙂emoji`).AddText(` and `).AddTag(`#punc.tuation`, `punc.tuation`).AddText(`
Last line`)
	assert.Equal(t, expected, post)
}

func TestNoTags(t *testing.T) {
	post := text.NewImporter(text.WithTagResolver(text.NoTagResolver)).Import(`
Some #hashtag
A#doubleHashtag#isAllowed
You can have #123numbers but not only #123 numbers
You can also have #🙂emoji and #punc.tuation
Last line`)
	expected := k3.NewPost().AddText(`
Some #hashtag
A#doubleHashtag#isAllowed
You can have #123numbers but not only #123 numbers
You can also have #🙂emoji and #`).AddLink(`punc.tuation`, `https://punc.tuation`).AddText(`
Last line`)
	assert.Equal(t, expected, post)
}

func TestConflicts(t *testing.T) {
	resolveAll := func(h string) *string { return &h }
	post := text.NewImporter(text.WithHandleResolver(resolveAll)).Import(`
URL or hashtag? http://example.com#hash example.com#hash
Handle or URL? @example.com
Hashtag or URL? #example#example.com #example#example.com#example
Hashtag or handle? #example@example.com
`)
	expected := k3.NewPost().AddText(`
URL or hashtag? `).AddLink(`example.com`, `http://example.com#hash`).AddText(` `).AddLink(`example.com`, `https://example.com#hash`).AddText(`
Handle or URL? `).AddMention(`@example.com`, `example.com`).AddText(`
Hashtag or URL? `).AddTag(`#example#`, `example`).AddLink(`example.com`, `https://example.com`).AddText(` `).AddTag(`#example#`, `example`).AddLink(`example.com`, `https://example.com#example`).AddText(`
Hashtag or handle? `).AddTag(`#example@example.com`, `example@example.com`).AddText(`
`)
	assert.Equal(t, expected, post)
}
