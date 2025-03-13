package k3_test

import (
	"testing"
	"time"

	"github.com/jtarrio/k3"
	"github.com/stretchr/testify/assert"
)

func TestNewPostSimple(t *testing.T) {
	post := k3.NewPost().
		AddText(`En un lugar de la Mancha`)

	expected := &k3.Post{Blocks: []k3.PostBlock{{Text: `En un lugar de la Mancha`}}}
	assert.Equal(t, expected, post)
}

func TestNewPostAllBlocks(t *testing.T) {
	post := k3.NewPost().
		SetCreationTime(time.Date(2025, time.January, 2, 12, 34, 56, 789000000, time.UTC)).
		AddText(`En esto, descubrieron `).
		AddLink(`treinta o cuarenta`, `https://url1`).
		AddText(` `).
		AddMention(`molinos`, `did1`).
		AddText(` de `).
		AddTag(`#viento`, `viento`).
		AddLanguage("es")

	expected := &k3.Post{
		CreationTime: ptr(time.Date(2025, time.January, 2, 12, 34, 56, 789000000, time.UTC)),
		Blocks: []k3.PostBlock{
			{Text: `En esto, descubrieron `},
			{Text: `treinta o cuarenta`, Link: ptr(`https://url1`)},
			{Text: ` `},
			{Text: `molinos`, Mention: ptr(`did1`)},
			{Text: ` de `},
			{Text: `#viento`, Tag: ptr(`viento`)},
		},
		Languages: []string{"es"},
	}
	assert.Equal(t, expected, post)
}

func TestCombineLikeBlocks(t *testing.T) {
	post := k3.NewPost().
		AddText(`En esto, `).
		AddText(`descubrieron `).
		AddLink(`treinta `, `https://url1`).
		AddLink(`o cuarenta`, `https://url1`).
		AddText(` `).
		AddMention(`mol`, `did1`).
		AddMention(`inos`, `did1`).
		AddText(` de `).
		AddTag(`#vie`, `viento`).
		AddTag(`nto`, `viento`)

	expected := &k3.Post{
		CreationTime: nil,
		Blocks: []k3.PostBlock{
			{Text: `En esto, descubrieron `},
			{Text: `treinta o cuarenta`, Link: ptr(`https://url1`)},
			{Text: ` `},
			{Text: `molinos`, Mention: ptr(`did1`)},
			{Text: ` de `},
			{Text: `#viento`, Tag: ptr(`viento`)},
		},
	}
	assert.Equal(t, expected, post)
}

func ptr[T any](v T) *T {
	return &v
}
