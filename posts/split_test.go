package posts_test

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/jtarrio/k3"
	"github.com/jtarrio/k3/posts"
	"github.com/stretchr/testify/assert"
)

var creationTime = time.Date(2025, time.January, 2, 12, 34, 56, 789000000, time.UTC)

func TestSplitShort(t *testing.T) {
	post := k3.NewPost().SetCreationTime(creationTime).AddLanguage("es").AddText(
		`En esto, descubrieron treinta o cuarenta molinos de viento que hay en aquel campo, y así como don Quijote los vio, dijo a su escudero:`)
	split := posts.Split(post)
	assert.Equal(t, []*k3.Post{post}, split)
}

func TestSplitIntoTwo(t *testing.T) {
	post := k3.NewPost().SetCreationTime(creationTime).AddLanguage("es").AddText(
		`xxxxxx En esto, descubrieron treinta o cuarenta molinos de viento que hay en aquel campo, y así como don Quijote los vio, dijo a su escudero:
—La ventura va guiando nuestras cosas mejor de lo que acertáramos a desear, porque ves allí, amigo Sancho Panza, donde se descubren treinta, o pocos más, desaforados gigantes, con quien pienso hacer batalla y quitarles a todos las vidas, con cuyos despojos comenzaremos a enriquecer; que ésta es buena guerra, y es gran servicio de Dios quitar tan mala simiente de sobre la faz de la tierra.
—¿Qué gigantes? —dijo Sancho Panza.`)
	split := posts.Split(post)
	expected := []*k3.Post{
		k3.NewPost().SetCreationTime(creationTime).AddLanguage("es").AddText(
			`[1/2] xxxxxx En esto, descubrieron treinta o cuarenta molinos de viento que hay en aquel campo, y así como don Quijote los vio, dijo a su escudero:
—La ventura va guiando nuestras cosas mejor de lo que acertáramos a desear, porque ves allí, amigo Sancho Panza, donde se descubren treinta, o pocos`),
		k3.NewPost().SetCreationTime(creationTime).AddLanguage("es").AddText(
			`[2/2] más, desaforados gigantes, con quien pienso hacer batalla y quitarles a todos las vidas, con cuyos despojos comenzaremos a enriquecer; que ésta es buena guerra, y es gran servicio de Dios quitar tan mala simiente de sobre la faz de la tierra.
—¿Qué gigantes? —dijo Sancho Panza.`),
	}
	assert.Equal(t, expected, split)
}

func TestSplitIntoEleven(t *testing.T) {
	post := k3.NewPost().SetCreationTime(creationTime).AddLanguage("es").AddText(
		`xxxxxx En esto, descubrieron treinta o cuarenta molinos de viento que hay en aquel campo, y así como don Quijote los vio, dijo a su escudero:
—La ventura va guiando nuestras cosas mejor de lo que acertáramos a desear, porque ves allí, amigo Sancho Panza, donde se descubren treinta, o pocos más, desaforados gigantes, con quien pienso hacer batalla y quitarles a todos las vidas, con cuyos despojos comenzaremos a enriquecer; que ésta es buena guerra, y es gran servicio de Dios quitar tan mala simiente de sobre la faz de la tierra.
—¿Qué gigantes? —dijo Sancho Panza.
—Aquellos que allí ves, —respondió su amo, —de los brazos largos, que los suelen tener algunos de casi dos leguas.
—Mire vuestra merced, —respondió Sancho, —que aquellos que allí se parecen no son gigantes, sino molinos de viento, y lo que en ellos parecen brazos son las aspas, que, volteadas del viento, hacen andar la piedra del molino.
—Bien parece, —respondió don Quijote, —que no estás cursado en esto de las aventuras: ellos son gigantes; y si tienes miedo, quítate de ahí, y ponte en oración en el espacio que yo voy a entrar con ellos en fiera y desigual batalla.
Y, diciendo esto, dio de espuelas a su caballo Rocinante, sin atender a las voces que su escudero Sancho le daba, advirtiéndole que, sin duda alguna, eran molinos de viento, y no gigantes, aquellos que iba a acometer. Pero él iba tan puesto en que eran gigantes que ni oía las voces de su escudero Sancho ni echaba de ver, aunque estaba ya bien cerca, lo que eran; antes iba diciendo en voces altas:
—Non fuyades, cobardes y viles criaturas, que un solo caballero es el que os acomete.
Levantóse en esto un poco de viento y las grandes aspas comenzaron a moverse, lo cual visto por don Quijote, dijo:
—Pues, aunque mováis más brazos que los del gigante Briareo, me lo habéis de pagar.
Y diciendo esto, y encomendándose de todo corazón a su señora Dulcinea, pidiéndole que en tal trance le socorriese, bien cubierto de su rodela, con la lanza en el ristre, arremetió a todo el galope de Rocinante y embistió con el primero molino que estaba delante; y, dándole una lanzada en el aspa, la volvió el viento con tanta furia que hizo la lanza pedazos, llevándose tras sí al caballo y al caballero, que fue rodando muy maltrecho por el campo. Acudió Sancho Panza a socorrerle, a todo el correr de su asno, y cuando llegó halló que no se podía menear: tal fue el golpe que dio con él Rocinante.
—¡Válame Dios! —dijo Sancho. —¿No le dije yo a vuestra merced que mirase bien lo que hacía, que no eran sino molinos de viento, y no lo podía ignorar sino quien llevase otros tales en la cabeza?
—Calla, amigo Sancho, —respondió don Quijote; —que las cosas de la guerra, más que otras, están sujetas a continua mudanza; cuanto más, que yo pienso, y es así verdad, que aquel sabio Frestón que me robó el aposento y los libros ha vuelto estos gigantes en molinos por quitarme la gloria de su vencimiento: tal es la enemistad que me tiene; mas al cabo al cabo, han de poder poco sus malas artes contra la bondad de mi espada.
—Dios lo haga como puede, —respondió Sancho Panza.`)
	split := posts.Split(post)
	expected := []*k3.Post{
		k3.NewPost().SetCreationTime(creationTime).AddLanguage("es").AddText(
			`[1/11] xxxxxx En esto, descubrieron treinta o cuarenta molinos de viento que hay en aquel campo, y así como don Quijote los vio, dijo a su escudero:
—La ventura va guiando nuestras cosas mejor de lo que acertáramos a desear, porque ves allí, amigo Sancho Panza, donde se descubren treinta, o pocos`),
		k3.NewPost().SetCreationTime(creationTime).AddLanguage("es").AddText(
			`[2/11] más, desaforados gigantes, con quien pienso hacer batalla y quitarles a todos las vidas, con cuyos despojos comenzaremos a enriquecer; que ésta es buena guerra, y es gran servicio de Dios quitar tan mala simiente de sobre la faz de la tierra.
—¿Qué gigantes? —dijo Sancho Panza.
—Aquellos que`),
		k3.NewPost().SetCreationTime(creationTime).AddLanguage("es").AddText(
			`[3/11] allí ves, —respondió su amo, —de los brazos largos, que los suelen tener algunos de casi dos leguas.
—Mire vuestra merced, —respondió Sancho, —que aquellos que allí se parecen no son gigantes, sino molinos de viento, y lo que en ellos parecen brazos son las aspas, que, volteadas del viento,`),
		k3.NewPost().SetCreationTime(creationTime).AddLanguage("es").AddText(
			`[4/11] hacen andar la piedra del molino.
—Bien parece, —respondió don Quijote, —que no estás cursado en esto de las aventuras: ellos son gigantes; y si tienes miedo, quítate de ahí, y ponte en oración en el espacio que yo voy a entrar con ellos en fiera y desigual batalla.
Y, diciendo esto, dio de`),
		k3.NewPost().SetCreationTime(creationTime).AddLanguage("es").AddText(
			`[5/11] espuelas a su caballo Rocinante, sin atender a las voces que su escudero Sancho le daba, advirtiéndole que, sin duda alguna, eran molinos de viento, y no gigantes, aquellos que iba a acometer. Pero él iba tan puesto en que eran gigantes que ni oía las voces de su escudero Sancho ni echaba de`),
		k3.NewPost().SetCreationTime(creationTime).AddLanguage("es").AddText(
			`[6/11] ver, aunque estaba ya bien cerca, lo que eran; antes iba diciendo en voces altas:
—Non fuyades, cobardes y viles criaturas, que un solo caballero es el que os acomete.
Levantóse en esto un poco de viento y las grandes aspas comenzaron a moverse, lo cual visto por don Quijote, dijo:
—Pues,`),
		k3.NewPost().SetCreationTime(creationTime).AddLanguage("es").AddText(
			`[7/11] aunque mováis más brazos que los del gigante Briareo, me lo habéis de pagar.
Y diciendo esto, y encomendándose de todo corazón a su señora Dulcinea, pidiéndole que en tal trance le socorriese, bien cubierto de su rodela, con la lanza en el ristre, arremetió a todo el galope de Rocinante y`),
		k3.NewPost().SetCreationTime(creationTime).AddLanguage("es").AddText(
			`[8/11] embistió con el primero molino que estaba delante; y, dándole una lanzada en el aspa, la volvió el viento con tanta furia que hizo la lanza pedazos, llevándose tras sí al caballo y al caballero, que fue rodando muy maltrecho por el campo. Acudió Sancho Panza a socorrerle, a todo el correr de`),
		k3.NewPost().SetCreationTime(creationTime).AddLanguage("es").AddText(
			`[9/11] su asno, y cuando llegó halló que no se podía menear: tal fue el golpe que dio con él Rocinante.
—¡Válame Dios! —dijo Sancho. —¿No le dije yo a vuestra merced que mirase bien lo que hacía, que no eran sino molinos de viento, y no lo podía ignorar sino quien llevase otros tales en la cabeza?`),
		k3.NewPost().SetCreationTime(creationTime).AddLanguage("es").AddText(
			`[10/11] —Calla, amigo Sancho, —respondió don Quijote; —que las cosas de la guerra, más que otras, están sujetas a continua mudanza; cuanto más, que yo pienso, y es así verdad, que aquel sabio Frestón que me robó el aposento y los libros ha vuelto estos gigantes en molinos por quitarme la gloria de`),
		k3.NewPost().SetCreationTime(creationTime).AddLanguage("es").AddText(
			`[11/11] su vencimiento: tal es la enemistad que me tiene; mas al cabo al cabo, han de poder poco sus malas artes contra la bondad de mi espada.
—Dios lo haga como puede, —respondió Sancho Panza.`),
	}
	assert.Equal(t, expected, split)
}

func TestSplitMultipleBlocks(t *testing.T) {
	post := k3.NewPost().SetCreationTime(creationTime).AddLanguage("es").
		AddText(`xxxxxx En esto, descubrieron treinta o cuarenta molinos de viento que hay en aquel campo, y así como don Quijote los vio, `).
		AddText("dijo a su escudero:\n—La ventura va guiando nuestras cosas mejor de lo que acertáramos a desear, ").
		AddText(`porque ves allí, amigo Sancho Panza, donde se descubren treinta, o pocos más, desaforados gigantes, `).
		AddText(`con quien pienso hacer batalla y quitarles a todos las vidas, `).
		AddText(`con cuyos despojos comenzaremos a enriquecer; que ésta es buena guerra, `).
		AddText("y es gran servicio de Dios quitar tan mala simiente de sobre la faz de la tierra.\n").
		AddLink(`—¿Qué gigantes?`, "https://example.com").
		AddText(` —dijo Sancho Panza.`)
	split := posts.Split(post)
	expected := []*k3.Post{
		k3.NewPost().SetCreationTime(creationTime).AddLanguage("es").AddText(
			`[1/2] xxxxxx En esto, descubrieron treinta o cuarenta molinos de viento que hay en aquel campo, y así como don Quijote los vio, dijo a su escudero:
—La ventura va guiando nuestras cosas mejor de lo que acertáramos a desear, porque ves allí, amigo Sancho Panza, donde se descubren treinta, o pocos`),
		k3.NewPost().SetCreationTime(creationTime).AddLanguage("es").AddText(
			`[2/2] más, desaforados gigantes, con quien pienso hacer batalla y quitarles a todos las vidas, con cuyos despojos comenzaremos a enriquecer; que ésta es buena guerra, y es gran servicio de Dios quitar tan mala simiente de sobre la faz de la tierra.
`).AddLink(`—¿Qué gigantes?`, "https://example.com").AddText(` —dijo Sancho Panza.`),
	}
	assert.Equal(t, expected, split)
}

func TestSplitCustomPartFunctionPrefix(t *testing.T) {
	post := k3.NewPost().SetCreationTime(creationTime).AddLanguage("es").AddText(
		`En esto, descubrieron treinta o cuarenta molinos de viento que hay en aquel campo, y así como don Quijote los vio, dijo a su escudero:
—La ventura va guiando nuestras cosas mejor de lo que acertáramos a desear, porque ves allí, amigo Sancho Panza, donde se descubren treinta, o pocos más, desaforados gigantes, con quien pienso hacer batalla y quitarles a todos las vidas, con cuyos despojos comenzaremos a enriquecer; que ésta es buena guerra, y es gran servicio de Dios quitar tan mala simiente de sobre la faz de la tierra.
—¿Qué gigantes? —dijo Sancho Panza.`)
	partFn := func(num, total int) string { return fmt.Sprintf("(Part %d of %d)", num, total) }
	split := posts.Split(post, posts.WithPrefix(partFn))
	expected := []*k3.Post{
		k3.NewPost().SetCreationTime(creationTime).AddLanguage("es").AddText(
			`(Part 1 of 2) En esto, descubrieron treinta o cuarenta molinos de viento que hay en aquel campo, y así como don Quijote los vio, dijo a su escudero:
—La ventura va guiando nuestras cosas mejor de lo que acertáramos a desear, porque ves allí, amigo Sancho Panza, donde se descubren treinta, o pocos`),
		k3.NewPost().SetCreationTime(creationTime).AddLanguage("es").AddText(
			`(Part 2 of 2) más, desaforados gigantes, con quien pienso hacer batalla y quitarles a todos las vidas, con cuyos despojos comenzaremos a enriquecer; que ésta es buena guerra, y es gran servicio de Dios quitar tan mala simiente de sobre la faz de la tierra.
—¿Qué gigantes? —dijo Sancho Panza.`),
	}
	assert.Equal(t, expected, split)
}

func TestSplitCustomPartFunctionSuffix(t *testing.T) {
	post := k3.NewPost().SetCreationTime(creationTime).AddLanguage("es").AddText(
		`En esto, descubrieron treinta o cuarenta molinos de viento que hay en aquel campo, y así como don Quijote los vio, dijo a su escudero:
—La ventura va guiando nuestras cosas mejor de lo que acertáramos a desear, porque ves allí, amigo Sancho Panza, donde se descubren treinta, o pocos más, desaforados gigantes, con quien pienso hacer batalla y quitarles a todos las vidas, con cuyos despojos comenzaremos a enriquecer; que ésta es buena guerra, y es gran servicio de Dios quitar tan mala simiente de sobre la faz de la tierra.
—¿Qué gigantes? —dijo Sancho Panza.`)
	partFn := func(num, total int) string { return fmt.Sprintf("(Part %d of %d)", num, total) }
	split := posts.Split(post, posts.WithSuffix(partFn))
	expected := []*k3.Post{
		k3.NewPost().SetCreationTime(creationTime).AddLanguage("es").AddText(
			`En esto, descubrieron treinta o cuarenta molinos de viento que hay en aquel campo, y así como don Quijote los vio, dijo a su escudero:
—La ventura va guiando nuestras cosas mejor de lo que acertáramos a desear, porque ves allí, amigo Sancho Panza, donde se descubren treinta, o pocos (Part 1 of 2)`),
		k3.NewPost().SetCreationTime(creationTime).AddLanguage("es").AddText(
			`más, desaforados gigantes, con quien pienso hacer batalla y quitarles a todos las vidas, con cuyos despojos comenzaremos a enriquecer; que ésta es buena guerra, y es gran servicio de Dios quitar tan mala simiente de sobre la faz de la tierra.
—¿Qué gigantes? —dijo Sancho Panza. (Part 2 of 2)`),
	}
	assert.Equal(t, expected, split)
}

func TestSplitMaximumSize(t *testing.T) {
	body := strings.Repeat(" a", 290)
	for l := range 20 {
		prefix := strings.Repeat("p", l)
		post := k3.NewPost().AddText(prefix + body)
		split := posts.Split(post)
		assert.LessOrEqual(t, split[0].GetGraphemeLength(), 300)
		runes := []rune(split[0].GetPlainText())
		assert.LessOrEqual(t, len(runes), 300)
	}
}
