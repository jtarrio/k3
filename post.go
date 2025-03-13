package k3

import (
	"strings"
	"time"
)

// Post contains a piece of content that can be published on Bluesky.
type Post struct {
	// CreationTime is the date and time associated with this post.
	CreationTime *time.Time
	// Blocks is the text content of the post, along with any links, tags, and mentions.
	Blocks []PostBlock
	// Languages is the list of codes of languages this post is written in.
	Languages []string
}

// PostBlock is a unit of content for a post.
type PostBlock struct {
	// Text contains the mandatory text for this block.
	Text string
	// Link contains a URI that the text for this block links to.
	Link *string
	// Mention contains a DID which is mentioned in this block.
	Mention *string
	// Tag contains a keyword that this block is tagged with.
	Tag *string
}

// NewPost creates a post.
func NewPost() *Post {
	post := &Post{}
	return post
}

// SetCreationTime changes the post's creation time.
func (p *Post) SetCreationTime(creationTime time.Time) *Post {
	p.CreationTime = &creationTime
	return p
}

// AddText adds a text block to the post with the given content.
func (p *Post) AddText(text string) *Post {
	return p.AddBlock(NewBlock(text))
}

// AddLink adds a link to the post with the given text and URI.
func (p *Post) AddLink(text string, uri string) *Post {
	return p.AddBlock(NewBlock(text, WithLink(uri)))
}

// AddMention adds a mention to the post with the given text and DID.
func (p *Post) AddMention(text string, did string) *Post {
	return p.AddBlock(NewBlock(text, WithMention(did)))
}

// AddTag adds a tag to the post with the given text and keyword.
func (p *Post) AddTag(text string, tag string) *Post {
	return p.AddBlock(NewBlock(text, WithTag(tag)))
}

// AddBlock adds a block to the post. If two consecutive blocks have the same link URI, mention DID, and tag, they are combined.
func (p *Post) AddBlock(block PostBlock) *Post {
	emptyStrPtr := func(a *string) bool { return a == nil || len(*a) == 0 }
	eqStrPtr := func(a, b *string) bool { return a == b || (a != nil && b != nil && *a == *b) }
	if len(block.Text) == 0 {
		return p
	}
	if emptyStrPtr(block.Link) {
		block.Link = nil
	}
	if emptyStrPtr(block.Mention) {
		block.Mention = nil
	}
	if emptyStrPtr(block.Tag) {
		block.Tag = nil
	}
	if len(p.Blocks) == 0 {
		p.Blocks = append(p.Blocks, block)
		return p
	}
	latest := &p.Blocks[len(p.Blocks)-1]
	if eqStrPtr(latest.Link, block.Link) && eqStrPtr(latest.Mention, block.Mention) && eqStrPtr(latest.Tag, block.Tag) {
		latest.Text += block.Text
	} else {
		p.Blocks = append(p.Blocks, block)
	}
	return p
}

// AddLanguage adds a code to the list of languages the post was written in.
func (p *Post) AddLanguage(lang string) *Post {
	p.Languages = append(p.Languages, lang)
	return p
}

// GetPlainText returns the plain text of the post.
func (p Post) GetPlainText() string {
	sb := strings.Builder{}
	for _, block := range p.Blocks {
		sb.WriteString(block.GetPlainText())
	}
	return sb.String()
}

// GetByteLength returns the length of the post, in bytes.
func (p Post) GetByteLength() int {
	l := 0
	for _, block := range p.Blocks {
		l += block.GetByteLength()
	}
	return l
}

// GetGraphemeLength returns the length of the post, in graphemes (Unicode characters).
func (p Post) GetGraphemeLength() int {
	l := 0
	for _, block := range p.Blocks {
		l += block.GetGraphemeLength()
	}
	return l
}

// NewBlock creates a new block with the given text and features.
func NewBlock(text string, features ...BlockFeature) PostBlock {
	block := PostBlock{Text: text}
	for _, feature := range features {
		feature(&block)
	}
	return block
}

// WithLink returns a 'link' feature with the given URI.
func WithLink(link string) BlockFeature {
	return func(b *PostBlock) {
		b.Link = &link
	}
}

// WithMention returns a 'mention' feature with the given DID.
func WithMention(did string) BlockFeature {
	return func(b *PostBlock) {
		b.Mention = &did
	}
}

// WithTag returns a 'tag' feature with the given keyword.
func WithTag(tag string) BlockFeature {
	return func(b *PostBlock) {
		b.Tag = &tag
	}
}

// BlockFeature is the type for features used in NewBlock.
type BlockFeature func(*PostBlock)

// GetPlainText returns the block's text
func (b PostBlock) GetPlainText() string {
	return b.Text
}

// GetByteLength returns the length of the block's text, in bytes.
func (b PostBlock) GetByteLength() int {
	return len(b.Text)
}

// GetGraphemeLength returns the length of the block's text, in graphemes.
func (b PostBlock) GetGraphemeLength() int {
	return len([]rune(b.Text))
}
