package posts

import (
	"fmt"
	"unicode"

	"github.com/jtarrio/atp"
)

const maxPostGraphemeLength = 300

// Split takes a post that is too long to be published on Bluesky, and splits it into multiple posts.
func Split(post *atp.Post) []*atp.Post {
	if post.GetGraphemeLength() <= maxPostGraphemeLength {
		return []*atp.Post{post}
	}

	blocks := splitBlocks(post.Blocks)
	solid := groupBlocks(blocks)
	var out []*atp.Post
	for _, postBlocks := range solid {
		newPost := atp.NewPost()
		newPost.CreationTime = post.CreationTime
		newPost.Languages = post.Languages
		for _, block := range postBlocks {
			newPost.AddBlock(block)
		}
		out = append(out, newPost)
	}
	return out
}

// splitBlocks creates a list of blocks that contain words and spacing between words, in that order.
func splitBlocks(blocks []atp.PostBlock) []atp.PostBlock {
	var output []atp.PostBlock
	isText := true
	start := 0
	for _, block := range blocks {
		runes := []rune(block.Text)
		addBlock := func(end int) {
			newBlock := block
			newBlock.Text = string(runes[start:end])
			output = append(output, newBlock)
			start = end
		}
		for i, r := range runes {
			switch isText {
			case false:
				if !unicode.IsSpace(r) {
					addBlock(i)
					isText = true
				}
			case true:
				if unicode.IsSpace(r) {
					addBlock(i)
					isText = false
				} else if i-start >= 100 {
					// Split long words into two or more blocks
					addBlock(i)
					addBlock(i)
				}
			}
		}
		addBlock(len(runes))
	}
	return output
}

// groupBlocks creates groups of blocks that, together with the item count, fit within the limits.
func groupBlocks(blocks []atp.PostBlock) [][]atp.PostBlock {
	getPrefix := func(i, n int) string { return fmt.Sprintf("[%d/%d] ", i, n) }
	maxCount := 9
	for {
		var outGroups [][]atp.PostBlock
		var group []atp.PostBlock
		var groupLen int
		startGroup := func(i int) {
			n := len(outGroups) + 1
			group = []atp.PostBlock{blocks[i]}
			prefixSize := len(getPrefix(n, maxCount))
			groupLen = blocks[i].GetGraphemeLength() + prefixSize
		}
		startGroup(0)
		for i := 2; i < len(blocks); i += 2 {
			bl := blocks[i].GetGraphemeLength()
			if groupLen+bl > maxPostGraphemeLength {
				outGroups = append(outGroups, group)
				startGroup(i)
			} else {
				groupLen += blocks[i-1].GetGraphemeLength() + bl
				group = append(group, blocks[i-1], blocks[i])
			}
		}
		outGroups = append(outGroups, group)
		if len(outGroups) > maxCount {
			maxCount = maxCount*10 + 9
			continue
		}
		for i := range outGroups {
			group := append(
				[]atp.PostBlock{atp.NewBlock(getPrefix(i+1, len(outGroups)))},
				outGroups[i]...)
			outGroups[i] = group
		}
		return outGroups
	}
}
