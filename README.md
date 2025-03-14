# K₃

K₃ — A Go library to post on Bluesky

## Purpose

The K₃ library is designed to let you easily write bots and other programs that can post on Bluesky.

It comprises the following components:

- Create, import, split, and convert posts:
  - [`k3.Post`](post.go) — Easily create new Bluesky posts from scratch.
  - [`import.text.Importer`](import/text/text.go) — Import posts from a text document. This importer recognizes URIs, username mentions, and hashtags.
  - [`import.html.Importer`](import/html/html.go) — Import posts from HTML content. This importer applies some basic formatting and can recognize links.
  - [`posts.Split`](posts/split.go) — Split a long post into multiple posts.
  - [`posts.Converter`](posts/converter.go) — Convert a `k3.Post` into a `bsky.FeedPost`, Bluesky's native post format.
- Publish posts:
  - [`client.Client`](client/client.go) — Connect to Bluesky, resolve usernames, and publish posts.
  - [`multiposter.Multiposter`](multiposter/multiposter.go) — Publish multiple posts as a sequence or as a thread.

## How to use this library

```go
import "github.com/jtarrio/k3"
```

### Create posts from scratch

```go
post := k3.NewPost().
    AddText(`This is an example of a post. `).
    AddText(`It can `).AddLink(`include links`, `https://github.com/jtarrio/k3`).AddText(`, `).
    AddText(`as well as mentions like `).AddMention(`@jacobo.tarrio.org`, `did:plc:ltjxg754ia655dp73hohri2r`).
    AddText(` (though you can use `).AddMention(`anything`, `did:plc:ltjxg754ia655dp73hohri2r`).
    AddText(` as the text) and `).AddTag(`tags`, `tags`).AddText(` and `).AddTag(`#hashtags`, `hashtags`).AddText(`.`)
```

### Import posts from text strings

```go
text := `This is another example of a post.
A link: https://github.com/jtarrio/k3
A username mention: @jacobo.tarrio.org
A hashtag: #hashtag`

importer := text.NewImporter(
    // Specify a handle resolver to be able to convert username mentions.
    // The client.HandleResolver uses the given client.Client to resolve usernames.
    text.WithHandleResolver(client.HandleResolver(myClient)))
post := importer.Import(text)
```

### Import posts from HTML strings

```go
htmlDoc := `<p>This post was written in HTML.</p>
<p>Links <a href="https://github.com/jtarrio/k3">can be converted</a>.</p>`

importer := html.NewImporter()
post := importer.Import(htmlDoc)
```

### Split a long post

```go
importer := html.NewImporter()
post := importer.Import(`<p>
  En esto, descubrieron treinta o cuarenta molinos de viento que hay en aquel
  campo, y así como don Quijote los vio, dijo a su escudero:
</p>
<p>
  -La ventura va guiando nuestras cosas mejor de lo que acertáramos a desear;
  porque ves allí, amigo Sancho Panza, donde se descubren treinta, o pocos más,
  desaforados gigantes, con quien pienso hacer batalla y quitarles a todos las
  vidas, con cuyos despojos comenzaremos a enriquecer; que ésta es buena guerra,
  y es gran servicio de Dios quitar tan mala simiente de sobre la faz de la
  tierra.
</p>
<p>-¿Qué gigantes? -dijo Sancho Panza.</p>
<p>
  -Aquéllos que allí ves -respondió su amo- de los brazos largos, que los suelen
  tener algunos de casi dos leguas.
</p>
<p>
  -Mire vuestra merced -respondió Sancho- que aquéllos que allí se parecen no
  son gigantes, sino molinos de viento, y lo que en ellos parecen brazos son las
  aspas, que, volteadas del viento, hacen andar la piedra del molino.
</p>`)

allPosts := posts.Split(post)
```

### Convert a post to a `bsky.FeedPost`

```go
post := k3.NewPost().AddText(`This is a post created with K₃`)
converter := posts.NewConverter()
feedPost := converter.Convert(post)
```

### Connect to Bluesky and publish a post

```go
c := client.New(identifier, password)
result, err := c.Publish(ctx, feedPost)
```

### Publish a series of posts as a thread

```go
feedPosts := converter.ToFeedPosts(posts.Split(post))

c := client.New(identifier, password)
mp := multiposter.New(cl, multiposter.AsThread())
result := mp.Publish(ctx, feedPosts)
if len(result.Remaining) > 0 {
    if err != nil {
        log.Printf("Publish returned error: %s; retrying after 5 seconds", err.Error())
    }
    time.Sleep(5)
    result = mp.Resume(ctx, result)
}
```

## License

K₃ is Copyright 2025 [Jacobo Tarrío Barreiro](https://jacobo.tarrio.org), and it's made available under the terms of the Apache License, version 2.0.
