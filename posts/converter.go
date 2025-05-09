package posts

import (
	"time"

	"github.com/bluesky-social/indigo/api/bsky"
	"github.com/jtarrio/k3"
)

// NewConverter creates a new Converter instance with the given options
func NewConverter(options ...ConverterOption) *Converter {
	c := &Converter{clock: k3.NewIncreasingClock(k3.SystemClock())}
	for _, option := range options {
		option(c)
	}
	return c
}

// WithClock makes the converter use the given clock to assign creation times to posts that don't have one.
func WithClock(clock k3.Clock) ConverterOption {
	return func(c *Converter) {
		c.clock = k3.NewIncreasingClock(clock)
	}
}

type ConverterOption func(*Converter)

type Converter struct {
	clock k3.Clock
}

// ToFeedPost generates a Bluesky FeedPost object from the content of the given post.
//
// The creation time, if unset, is populated with an always-increasing clock so that different posts have different creation times.
func (c *Converter) ToFeedPost(post *k3.Post) *bsky.FeedPost {
	var creationTime time.Time
	if post.CreationTime == nil {
		creationTime = c.clock.Now()
	} else {
		creationTime = *post.CreationTime
	}

	out := &bsky.FeedPost{
		LexiconTypeID: "app.bsky.feed.post",
		CreatedAt:     creationTime.Format("2006-01-02T15:04:05.999Z07:00"),
		Text:          post.GetPlainText(),
		Langs:         post.Languages,
	}
	start := 0
	for _, block := range post.Blocks {
		end := start + block.GetByteLength()
		features := getBlockFeatures(&block)
		if len(features) > 0 {
			facet := &bsky.RichtextFacet{
				Features: features,
				Index: &bsky.RichtextFacet_ByteSlice{
					ByteStart: int64(start),
					ByteEnd:   int64(end),
				},
			}
			out.Facets = append(out.Facets, facet)
		}
		start = end
	}
	return out
}

// ToFeedPosts converts a slice of posts into a slice of FeedPost objects.
func (c *Converter) ToFeedPosts(posts []*k3.Post) []*bsky.FeedPost {
	var out []*bsky.FeedPost
	for _, post := range posts {
		out = append(out, c.ToFeedPost(post))
	}
	return out
}

func getBlockFeatures(block *k3.PostBlock) []*bsky.RichtextFacet_Features_Elem {
	var out []*bsky.RichtextFacet_Features_Elem
	if block.Link != nil && len(*block.Link) > 0 {
		out = append(out, &bsky.RichtextFacet_Features_Elem{
			RichtextFacet_Link: &bsky.RichtextFacet_Link{
				LexiconTypeID: "app.bsky.richtext.facet#link",
				Uri:           *block.Link,
			},
		})
	}
	if block.Mention != nil && len(*block.Mention) > 0 {
		out = append(out, &bsky.RichtextFacet_Features_Elem{
			RichtextFacet_Mention: &bsky.RichtextFacet_Mention{
				LexiconTypeID: "app.bsky.richtext.facet#mention",
				Did:           *block.Mention,
			},
		})
	}
	if block.Tag != nil && len(*block.Tag) > 0 {
		out = append(out, &bsky.RichtextFacet_Features_Elem{
			RichtextFacet_Tag: &bsky.RichtextFacet_Tag{
				LexiconTypeID: "app.bsky.richtext.facet#tag",
				Tag:           *block.Tag,
			},
		})
	}
	return out
}
