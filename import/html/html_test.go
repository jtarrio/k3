package html_test

import (
	"testing"

	"github.com/jtarrio/atp"
	"github.com/jtarrio/atp/import/html"
	"github.com/stretchr/testify/assert"
)

func TestEasyText(t *testing.T) {
	post, err := html.NewImporter().Import(`   Some   easy   text.
With inconsistent spacing and


carriage


returns.`)
	assert.NoError(t, err)
	expected := atp.NewPost().AddText(`Some easy text. With inconsistent spacing and carriage returns.`)
	assert.Equal(t, expected, post)
}

func TestHtmlWithLinks(t *testing.T) {
	post, err := html.NewImporter().Import(`Some slightly <a href="url1">harder</a> text <a href="url2">with
links</a>.`)
	assert.NoError(t, err)
	expected := atp.NewPost().AddText(`Some slightly `).AddLink(`harder`, "url1").AddText(` text `).AddLink(`with links`, "url2").AddText(`.`)
	assert.Equal(t, expected, post)
}

func TestHtmlWithFormatting(t *testing.T) {
	post, err := html.NewImporter().Import(`<p>This is a paragraph.</p><p>This is another paragraph.</p><p>And another one.</p>`)
	assert.NoError(t, err)
	expected := atp.NewPost().AddText(`This is a paragraph.
This is another paragraph.
And another one.`)
	assert.Equal(t, expected, post)
}

func TestHtmlWithImplicitParagraphs(t *testing.T) {
	post, err := html.NewImporter().Import(`<p>This is a paragraph.</p>This is another paragraph.</p><p>And another one.<p>Yet another one.</p>`)
	assert.NoError(t, err)
	expected := atp.NewPost().AddText(`This is a paragraph.
This is another paragraph.
And another one.
Yet another one.`)
	assert.Equal(t, expected, post)
}

func TestHtmlWithHeaders(t *testing.T) {
	post, err := html.NewImporter().Import(`
<h1>Title</h1>

<h2>Subtitle</h2>

Some text.

<h3>Sub-sub title</h3>
<p>Some other text.</p><h4>Lost count</h4><h5>For real</h5>
More text.
<h6>There is no more down here</h6>
The end.`)
	assert.NoError(t, err)
	expected := atp.NewPost().AddText(`Title
Subtitle
Some text.
Sub-sub title
Some other text.
Lost count
For real
More text.
There is no more down here
The end.`)
	assert.Equal(t, expected, post)
}

func TestHtmlWithBlockElements(t *testing.T) {
	post, err := html.NewImporter().Import(`
Outside.<p>Paragraph</p><div>Div</div>

Empty paragraphs:

<p></p>
<p></p>
<p></p>

I don't know where to include the HR, so...
<hr>
Ditto with the<br>
   line<br>
   break

<p>Mixing <div>element</div> types.</p>
`)
	assert.NoError(t, err)
	expected := atp.NewPost().AddText(`Outside.
Paragraph
Div
Empty paragraphs:
I don't know where to include the HR, so...
-----
Ditto with the
line
break
Mixing
element
types.`)
	assert.Equal(t, expected, post)
}

func TestHtmlWithLists(t *testing.T) {
	post, err := html.NewImporter().Import(`
<ul>
<li>Unordered element.</li>
<li>Another element.</li>
</ul>
<ol>
<li>Ordered element.</li>
<li>Another ordered element.
<ul><li>Nesting
<ol><li>More elements<li>And more</ol>
<li>Unnesting</ul>
<li>More unnesting
</ol>
The end.`)
	assert.NoError(t, err)
	expected := atp.NewPost().AddText(`  * Unordered element.
  * Another element.
  1. Ordered element.
  2. Another ordered element.
    * Nesting
      1. More elements
      2. And more
    * Unnesting
  3. More unnesting
The end.`)
	assert.Equal(t, expected, post)
}

func TestHtmlWithEscapedCharacters(t *testing.T) {
	post, err := html.NewImporter().Import(`
En esto&comma; descubrieron treinta o cuarenta molinos de viento que hay en aquel campo&comma;
y as&iacute; como don Quijote los vio&comma; dijo a su escudero&colon;<br>
&mdash;La ventura va guiando nuestras cosas mejor de lo que acert&aacute;ramos a desear&comma;
porque ves all&iacute;&comma; amigo Sancho Panza&comma; donde se descubren treinta&comma;
o pocos m&aacute;s&comma; desaforados gigantes&comma; con quien pienso hacer batalla
y quitarles a todos las vidas&comma; con cuyos despojos comenzaremos a enriquecer&semi;
que &eacute;sta es buena guerra&comma; y es gran servicio de Dios
quitar tan mala simiente de sobre la faz de la tierra&period;<br>
&mdash;&iquest;Qu&eacute; gigantes&quest; &mdash;dijo Sancho Panza&period;`)
	assert.NoError(t, err)
	expected := atp.NewPost().AddText(`En esto, descubrieron treinta o cuarenta molinos de viento que hay en aquel campo, y así como don Quijote los vio, dijo a su escudero:
—La ventura va guiando nuestras cosas mejor de lo que acertáramos a desear, porque ves allí, amigo Sancho Panza, donde se descubren treinta, o pocos más, desaforados gigantes, con quien pienso hacer batalla y quitarles a todos las vidas, con cuyos despojos comenzaremos a enriquecer; que ésta es buena guerra, y es gran servicio de Dios quitar tan mala simiente de sobre la faz de la tierra.
—¿Qué gigantes? —dijo Sancho Panza.`)
	assert.Equal(t, expected, post)
}

func TestUnknownTags(t *testing.T) {
	post, err := html.NewImporter().Import(`
Some <b>bold</b> and <i>italic</i> text.
And some <x-madeup>made up</x-madeup> tags.`)
	assert.NoError(t, err)
	expected := atp.NewPost().AddText(`Some bold and italic text. And some made up tags.`)
	assert.Equal(t, expected, post)
}

func TestIgnoredTags(t *testing.T) {
	post, err := html.NewImporter().Import(`
<p>Some text.
<script>foo bar baz</script>
More text.</p>
<!--Comments are ignored too.-->
<p>Several levels of ignoring
<iframe>one two<script>three four</script>five six</iframe>
done.`)
	assert.NoError(t, err)
	expected := atp.NewPost().AddText(`Some text. More text.
Several levels of ignoring done.`)
	assert.Equal(t, expected, post)
}
