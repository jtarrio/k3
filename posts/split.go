package posts

import (
	"fmt"
	"unicode"

	"github.com/jtarrio/k3"
)

// Split takes a post that is too long to be published on Bluesky, and splits it into multiple posts.
//
// You can use WithPrefix and WithSuffix to pass functions that will modify now the parts are numbered.
// Note that this code assumes that numbering functions return strings that increase monotonically
// with the post number. That means: the string for post N must be the same size or larger than
// the string for post N-1.
//
// This happens trivially if your function returns strings like "Part 3 of 9", but not if it returns
// strings like "Part three of nine", because the next string would be "part four of nine",
// which is shorter.
func Split(post *k3.Post, options ...SplitOption) []*k3.Post {
	if post.GetGraphemeLength() <= maxPostGraphemeLength {
		return []*k3.Post{post}
	}

	partFn := DefaultPartFunction
	partPrefix := true
	for _, option := range options {
		partFn, partPrefix = option()
	}

	blocks := splitBlocks(post.Blocks)
	solid := groupBlocks(blocks, partFn, partPrefix)
	var out []*k3.Post
	for _, postBlocks := range solid {
		newPost := k3.NewPost()
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
func splitBlocks(blocks []k3.PostBlock) []k3.PostBlock {
	var output []k3.PostBlock
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
func groupBlocks(blocks []k3.PostBlock, partFn PartFunction, partPrefix bool) [][]k3.PostBlock {
	maxCount := 9
	for {
		var outGroups [][]k3.PostBlock
		var group []k3.PostBlock
		var groupLen int
		startGroup := func(i int) {
			n := len(outGroups) + 1
			group = []k3.PostBlock{blocks[i]}
			prefixSize := len(partFn(n, maxCount)) + 1
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
			var group []k3.PostBlock
			if partPrefix {
				group = append(
					[]k3.PostBlock{k3.NewBlock(partFn(i+1, len(outGroups)) + " ")},
					outGroups[i]...)
			} else {
				group = append(outGroups[i], k3.NewBlock(" "+partFn(i+1, len(outGroups))))
			}
			outGroups[i] = group
		}
		return outGroups
	}
}

// WithPrefix uses the given function as a part numbering function, and prepends its result to each message.
func WithPrefix(fn PartFunction) SplitOption {
	return func() (partFn PartFunction, prefix bool) {
		return fn, true
	}
}

// WithSuffix uses the given function as a part numbering function, and appends its result to each message.
func WithSuffix(fn PartFunction) SplitOption {
	return func() (partFn PartFunction, prefix bool) {
		return fn, false
	}
}

// DefaultPartFunction is the default part numbering function.
func DefaultPartFunction(num, total int) string {
	return fmt.Sprintf("[%d/%d]", num, total)
}

// PartFunction is the type for a part numbering function.
//
// Note that this code assumes that numbering functions return strings that increase monotonically
// with the post number. That means: the string for post N must be the same size or larger than
// the string for post N-1.
//
// This happens trivially if your function returns strings like "Part 3 of 9", but not if it returns
// strings like "Part three of nine", because the next string would be "part four of nine",
// which is shorter even though the number is bigger.
type PartFunction func(num, total int) string

type SplitOption func() (partFn PartFunction, prefix bool)

const maxPostGraphemeLength = 300
