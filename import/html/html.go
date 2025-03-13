// Package html provides an Importer that converts HTML code into a Bluesky post.
//
// The importer parses the provided HTML code and applies some simple formatting,
// converting links wherever found.
//
// Supported tags include <p>, <div>, <h1> through <h6>, <hr>, <ul>, <ol>, <li>, <pre>, and <br>.
//
// Other tags are ignored but their content is still added to the post. Those include <b>, <i>, <span>, etc.
//
// Some tags are ignored and their content is not added to the post. Those include <script>, <link>, <iframe>, etc.
package html

import (
	"fmt"
	"strings"
	"unicode"

	"github.com/jtarrio/atp"
	"golang.org/x/net/html"
)

// Importer is an interface to convert some HTML into a Bluesky post.
type Importer interface {
	// Import converts the given HTML code into a post.
	Import(html string) (*atp.Post, error)
}

// NewImporter creates a new HTML importer.
func NewImporter() Importer {
	return &importer{}
}

type importer struct{}

func (*importer) Import(input string) (*atp.Post, error) {
	doc, err := html.Parse(strings.NewReader(input))
	if err != nil {
		return nil, &HtmlParseError{err}
	}

	conv := &converter{post: atp.NewPost()}
	conv.convert(doc)
	return conv.post, nil
}

type converter struct {
	post       *atp.Post
	inPara     bool
	inPre      bool
	wantSpace  bool
	linkTarget []string
	listIndex  []int
	ignore     int
}

func (c *converter) convert(n *html.Node) {
	switch n.Type {
	case html.TextNode:
		c.addText(n.Data)
	case html.ElementNode:
		tag, ok := htmlTags[n.Data]
		if !ok {
			tag = htmlTag{doNothing, doNothing}
		}
		tag.preFn(c, n)
		for s := n.FirstChild; s != nil; s = s.NextSibling {
			c.convert(s)
		}
		tag.postFn(c, n)
	default:
		for s := n.FirstChild; s != nil; s = s.NextSibling {
			c.convert(s)
		}
	}
}

var htmlTags = map[string]htmlTag{
	"a":      {startAnchor, endAnchor},
	"br":     {paragraphBoundary, doNothing},
	"p":      {paragraphBoundary, paragraphBoundary},
	"div":    {paragraphBoundary, paragraphBoundary},
	"h1":     {paragraphBoundary, paragraphBoundary},
	"h2":     {paragraphBoundary, paragraphBoundary},
	"h3":     {paragraphBoundary, paragraphBoundary},
	"h4":     {paragraphBoundary, paragraphBoundary},
	"h5":     {paragraphBoundary, paragraphBoundary},
	"h6":     {paragraphBoundary, paragraphBoundary},
	"pre":    {startPre, endPre},
	"hr":     {startHr, doNothing},
	"ol":     {startOl, endList},
	"ul":     {startUl, endList},
	"li":     {startLi, paragraphBoundary},
	"head":   {startIgnore, endIgnore},
	"script": {startIgnore, endIgnore},
	"applet": {startIgnore, endIgnore},
	"object": {startIgnore, endIgnore},
	"svg":    {startIgnore, endIgnore},
	"style":  {startIgnore, endIgnore},
	"link":   {startIgnore, endIgnore},
	"iframe": {startIgnore, endIgnore},
}

type htmlTag struct {
	preFn  func(*converter, *html.Node)
	postFn func(*converter, *html.Node)
}

func doNothing(*converter, *html.Node) {}

func paragraphBoundary(c *converter, _ *html.Node) {
	c.inPara = false
	c.wantSpace = false
}

func startIgnore(c *converter, _ *html.Node) {
	c.ignore++
}

func endIgnore(c *converter, _ *html.Node) {
	if c.ignore > 0 {
		c.ignore--
	}
}

func startAnchor(c *converter, n *html.Node) {
	if c.inPara && c.wantSpace {
		c.addTextBlock(" ")
		c.wantSpace = false
	}
	for _, a := range n.Attr {
		if a.Key == "href" {
			c.linkTarget = append(c.linkTarget, a.Val)
			return
		}
	}
	c.linkTarget = append(c.linkTarget, "")
}

func endAnchor(c *converter, _ *html.Node) {
	if len(c.linkTarget) > 0 {
		c.linkTarget = c.linkTarget[:len(c.linkTarget)-1]
	}
}

func startPre(c *converter, n *html.Node) {
	paragraphBoundary(c, n)
	c.inPre = true
}

func endPre(c *converter, n *html.Node) {
	paragraphBoundary(c, n)
	c.inPre = false
}

func startHr(c *converter, n *html.Node) {
	paragraphBoundary(c, n)
	c.addText(`-----`)
	paragraphBoundary(c, n)
}

func startUl(c *converter, n *html.Node) {
	paragraphBoundary(c, n)
	c.listIndex = append(c.listIndex, 0)
}

func startOl(c *converter, n *html.Node) {
	paragraphBoundary(c, n)
	c.listIndex = append(c.listIndex, 1)
}

func endList(c *converter, n *html.Node) {
	paragraphBoundary(c, n)
	if len(c.listIndex) > 0 {
		c.listIndex = c.listIndex[:len(c.listIndex)-1]
	}
}

func startLi(c *converter, n *html.Node) {
	paragraphBoundary(c, n)
	if len(c.listIndex) > 0 {
		idx := c.listIndex[len(c.listIndex)-1]
		if idx == 0 {
			c.addText("* ")
		} else {
			c.addText(fmt.Sprintf("%d. ", idx))
			c.listIndex[len(c.listIndex)-1] = idx + 1
		}
	}
}

func (c *converter) addText(txt string) {
	if c.inPre {
		if !c.inPara {
			c.addTextBlock("\n")
			c.inPara = true
		}
		c.addTextBlock(txt)
		return
	}
	const zeroWidthSpace rune = '\u200b'
	for _, r := range txt {
		if unicode.IsSpace(r) {
			if r != zeroWidthSpace && c.inPara {
				c.wantSpace = true
			}
		} else {
			if !c.inPara {
				if len(c.post.Blocks) > 0 {
					c.addTextBlock("\n")
				}
				if len(c.listIndex) > 0 {
					c.addTextBlock(strings.Repeat("  ", len(c.listIndex)))
				}
			} else if c.wantSpace {
				c.addTextBlock(" ")
				c.wantSpace = false
			}
			c.addTextBlock(string([]rune{r}))
			c.inPara = true
		}
	}
}

func (c *converter) addTextBlock(txt string) {
	if c.ignore > 0 {
		return
	}
	if len(c.linkTarget) > 0 {
		c.post.AddLink(txt, c.linkTarget[len(c.linkTarget)-1])
		return
	}
	c.post.AddText(txt)
}

type HtmlParseError struct {
	err error
}

func (e *HtmlParseError) Error() string {
	return fmt.Sprintf("error parsing HTML for import: %s", e.err.Error())
}

func (e *HtmlParseError) Unwrap() error {
	return e.err
}
