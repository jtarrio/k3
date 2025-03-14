package posts_test

import (
	"strings"
	"testing"
	"time"

	"github.com/bluesky-social/indigo/api/bsky"
	"github.com/jtarrio/k3"
	"github.com/jtarrio/k3/posts"
	atptesting "github.com/jtarrio/k3/testing"
	"github.com/stretchr/testify/assert"
)

func TestConvertPlainText(t *testing.T) {
	fakeClock := &atptesting.FakeClock{Time: time.Date(2025, time.January, 2, 12, 34, 56, 789000000, time.UTC)}
	c := posts.NewConverter(posts.WithClock(fakeClock))

	post := k3.NewPost().
		AddText(`En esto, descubrieron treinta o cuarenta molinos de viento que hay`).
		AddText(` en aquel campo, y así como don Quijote los vio, dijo a su escudero:`)
	feedPost := c.ToFeedPost(post)
	expected := &bsky.FeedPost{
		LexiconTypeID: "app.bsky.feed.post",
		CreatedAt:     "2025-01-02T12:34:56.789Z",
		Text:          `En esto, descubrieron treinta o cuarenta molinos de viento que hay en aquel campo, y así como don Quijote los vio, dijo a su escudero:`,
	}
	assert.Equal(t, expected, feedPost)
}

func TestUseIncreasingClockWhenCreationTimeIsUnset(t *testing.T) {
	fakeClock := &atptesting.FakeClock{Time: time.Date(2025, time.January, 2, 12, 34, 56, 789000000, time.UTC)}
	c := posts.NewConverter(posts.WithClock(fakeClock))

	feedPost := c.ToFeedPost(k3.NewPost().AddText(`En esto, descubrieron treinta o cuarenta molinos de viento que hay`))
	expected := &bsky.FeedPost{
		LexiconTypeID: "app.bsky.feed.post",
		CreatedAt:     "2025-01-02T12:34:56.789Z",
		Text:          `En esto, descubrieron treinta o cuarenta molinos de viento que hay`,
	}
	assert.Equal(t, expected, feedPost)

	feedPost = c.ToFeedPost(k3.NewPost().AddText(`en aquel campo, y así como don Quijote los vio, dijo a su escudero:`))
	expected = &bsky.FeedPost{
		LexiconTypeID: "app.bsky.feed.post",
		CreatedAt:     "2025-01-02T12:34:56.79Z",
		Text:          `en aquel campo, y así como don Quijote los vio, dijo a su escudero:`,
	}
	assert.Equal(t, expected, feedPost)
}

func TestUseCreationTimeWhenSet(t *testing.T) {
	fakeClock := &atptesting.FakeClock{Time: time.Date(2025, time.January, 2, 12, 34, 56, 789000000, time.UTC)}
	c := posts.NewConverter(posts.WithClock(fakeClock))

	post := k3.NewPost().
		SetCreationTime(time.Date(1999, time.December, 25, 1, 23, 45, 678000000, time.UTC)).
		AddText(`En esto, descubrieron treinta o cuarenta molinos de viento que hay`)
	feedPost := c.ToFeedPost(post)
	expected := &bsky.FeedPost{
		LexiconTypeID: "app.bsky.feed.post",
		CreatedAt:     "1999-12-25T01:23:45.678Z",
		Text:          `En esto, descubrieron treinta o cuarenta molinos de viento que hay`,
	}
	assert.Equal(t, expected, feedPost)

	post = k3.NewPost().
		SetCreationTime(time.Date(1999, time.December, 25, 1, 23, 45, 678000000, time.UTC)).
		AddText(`en aquel campo, y así como don Quijote los vio, dijo a su escudero:`)
	feedPost = c.ToFeedPost(post)
	expected = &bsky.FeedPost{
		LexiconTypeID: "app.bsky.feed.post",
		CreatedAt:     "1999-12-25T01:23:45.678Z",
		Text:          `en aquel campo, y así como don Quijote los vio, dijo a su escudero:`,
	}
	assert.Equal(t, expected, feedPost)
}

func TestConvertLink(t *testing.T) {
	fakeClock := &atptesting.FakeClock{Time: time.Date(2025, time.January, 2, 12, 34, 56, 789000000, time.UTC)}
	c := posts.NewConverter(posts.WithClock(fakeClock))

	post := k3.NewPost().
		AddText(`y así como `).
		AddLink(`don Quijote`, `https://url1`).
		AddText(` los vio, dijo a su escudero`)
	feedPost := c.ToFeedPost(post)
	expectedText := `y así como don Quijote los vio, dijo a su escudero`
	expected := &bsky.FeedPost{
		LexiconTypeID: "app.bsky.feed.post",
		CreatedAt:     "2025-01-02T12:34:56.789Z",
		Text:          expectedText,
		Facets: []*bsky.RichtextFacet{
			{Features: link(`https://url1`), Index: indexOf(expectedText, `don Quijote`)},
		},
	}
	assert.Equal(t, expected, feedPost)
}

func TestConvertMention(t *testing.T) {
	fakeClock := &atptesting.FakeClock{Time: time.Date(2025, time.January, 2, 12, 34, 56, 789000000, time.UTC)}
	c := posts.NewConverter(posts.WithClock(fakeClock))

	post := k3.NewPost().
		AddText(`y así como `).
		AddMention(`don Quijote`, `did1`).
		AddText(` los vio, dijo a `).
		AddMention(`su escudero`, `did2`)
	feedPost := c.ToFeedPost(post)
	const expectedText = `y así como don Quijote los vio, dijo a su escudero`
	expected := &bsky.FeedPost{
		LexiconTypeID: "app.bsky.feed.post",
		CreatedAt:     "2025-01-02T12:34:56.789Z",
		Text:          expectedText,
		Facets: []*bsky.RichtextFacet{
			{Features: mention(`did1`), Index: indexOf(expectedText, `don Quijote`)},
			{Features: mention(`did2`), Index: indexOf(expectedText, `su escudero`)},
		},
	}
	assert.Equal(t, expected, feedPost)
}

func TestConvertTag(t *testing.T) {
	fakeClock := &atptesting.FakeClock{Time: time.Date(2025, time.January, 2, 12, 34, 56, 789000000, time.UTC)}
	c := posts.NewConverter(posts.WithClock(fakeClock))

	post := k3.NewPost().
		AddText(`y así como `).
		AddTag(`#DonQuijote`, `donquijote`).
		AddText(` los vio, dijo a `).
		AddTag(`su escudero`, `suescudero`)
	feedPost := c.ToFeedPost(post)
	const expectedText = `y así como #DonQuijote los vio, dijo a su escudero`
	expected := &bsky.FeedPost{
		LexiconTypeID: "app.bsky.feed.post",
		CreatedAt:     "2025-01-02T12:34:56.789Z",
		Text:          expectedText,
		Facets: []*bsky.RichtextFacet{
			{Features: tag(`donquijote`), Index: indexOf(expectedText, `#DonQuijote`)},
			{Features: tag(`suescudero`), Index: indexOf(expectedText, `su escudero`)},
		},
	}
	assert.Equal(t, expected, feedPost)
}

func TestConvertLanguage(t *testing.T) {
	fakeClock := &atptesting.FakeClock{Time: time.Date(2025, time.January, 2, 12, 34, 56, 789000000, time.UTC)}
	c := posts.NewConverter(posts.WithClock(fakeClock))

	post := k3.NewPost().
		AddText(`y así como don Quijote los vio, he said to his squire`).
		AddLanguage("es").
		AddLanguage("en")
	feedPost := c.ToFeedPost(post)
	expected := &bsky.FeedPost{
		LexiconTypeID: "app.bsky.feed.post",
		CreatedAt:     "2025-01-02T12:34:56.789Z",
		Text:          `y así como don Quijote los vio, he said to his squire`,
		Langs:         []string{"es", "en"},
	}
	assert.Equal(t, expected, feedPost)
}

func link(url string) []*bsky.RichtextFacet_Features_Elem {
	return []*bsky.RichtextFacet_Features_Elem{
		{
			RichtextFacet_Link: &bsky.RichtextFacet_Link{
				LexiconTypeID: "app.bsky.richtext.facet#link",
				Uri:           url,
			},
		},
	}
}

func mention(did string) []*bsky.RichtextFacet_Features_Elem {
	return []*bsky.RichtextFacet_Features_Elem{
		{
			RichtextFacet_Mention: &bsky.RichtextFacet_Mention{
				LexiconTypeID: "app.bsky.richtext.facet#mention",
				Did:           did,
			},
		},
	}
}

func tag(tag string) []*bsky.RichtextFacet_Features_Elem {
	return []*bsky.RichtextFacet_Features_Elem{
		{
			RichtextFacet_Tag: &bsky.RichtextFacet_Tag{
				LexiconTypeID: "app.bsky.richtext.facet#tag",
				Tag:           tag,
			},
		},
	}
}

func indexOf(text, substr string) *bsky.RichtextFacet_ByteSlice {
	index := strings.Index(text, substr)
	return &bsky.RichtextFacet_ByteSlice{
		ByteStart: int64(index),
		ByteEnd:   int64(index + len(substr)),
	}
}
