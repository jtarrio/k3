package multiposter_test

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/bluesky-social/indigo/api/atproto"
	"github.com/bluesky-social/indigo/api/bsky"
	"github.com/jtarrio/k3"
	"github.com/jtarrio/k3/client"
	"github.com/jtarrio/k3/multiposter"
	"github.com/jtarrio/k3/posts"
	atptesting "github.com/jtarrio/k3/testing"
	"github.com/stretchr/testify/assert"
)

func TestPostSequence(t *testing.T) {
	c := &fakeClient{failAfter: -1}
	m := multiposter.New(c)
	posts := getPosts(10)
	result := m.Publish(context.Background(), posts)
	expectedResult := &multiposter.PublishResult{}
	for i := range 10 {
		expectedResult.Published = append(expectedResult.Published, &client.PublishResult{
			Uri: uriOf(i),
			Cid: cidOf(i),
		})
	}
	assert.Equal(t, expectedResult, result)
	var expectedPublished []publishedPost
	for i := range posts {
		expectedPublished = append(expectedPublished, publishedPost{
			cid:  cidOf(i),
			uri:  uriOf(i),
			post: posts[i],
		})
	}
	assert.Equal(t, expectedPublished, c.posts)
}

func TestResumeSequence(t *testing.T) {
	c := &fakeClient{failAfter: 3}
	m := multiposter.New(c)
	posts := getPosts(10)
	result := m.Publish(context.Background(), posts)
	expectedResult := &multiposter.PublishResult{}
	for i := range 3 {
		expectedResult.Published = append(expectedResult.Published, &client.PublishResult{
			Uri: uriOf(i),
			Cid: cidOf(i),
		})
	}
	expectedResult.Remaining = posts[3:]
	expectedResult.Error = errPublish
	assert.Equal(t, expectedResult, result)
	assert.Len(t, c.posts, 3)

	c.failAfter = -1

	result2 := m.Resume(context.Background(), result)
	expectedResult = &multiposter.PublishResult{}
	for i := range 10 {
		expectedResult.Published = append(expectedResult.Published, &client.PublishResult{
			Uri: uriOf(i),
			Cid: cidOf(i),
		})
	}
	assert.Equal(t, expectedResult, result2)
	var expectedPublished []publishedPost
	for i := range posts {
		expectedPublished = append(expectedPublished, publishedPost{
			cid:  cidOf(i),
			uri:  uriOf(i),
			post: posts[i],
		})
	}
	assert.Equal(t, expectedPublished, c.posts)
}

func TestPostThread(t *testing.T) {
	c := &fakeClient{failAfter: -1}
	m := multiposter.New(c, multiposter.AsThread())
	posts := getPosts(10)
	result := m.Publish(context.Background(), posts)
	expectedResult := &multiposter.PublishResult{}
	for i := range 10 {
		expectedResult.Published = append(expectedResult.Published, &client.PublishResult{
			Uri: uriOf(i),
			Cid: cidOf(i),
		})
	}
	assert.Equal(t, expectedResult, result)
	var expectedPublished []publishedPost
	for i := range posts {
		postWithReply := *posts[i]
		if i > 0 {
			postWithReply.Reply = &bsky.FeedPost_ReplyRef{
				Parent: &atproto.RepoStrongRef{
					LexiconTypeID: "com.atproto.repo.strongRef",
					Cid:           cidOf(i - 1),
					Uri:           uriOf(i - 1),
				},
				Root: &atproto.RepoStrongRef{
					LexiconTypeID: "com.atproto.repo.strongRef",
					Cid:           cidOf(0),
					Uri:           uriOf(0),
				},
			}
		}
		expectedPublished = append(expectedPublished, publishedPost{
			cid:  cidOf(i),
			uri:  uriOf(i),
			post: &postWithReply,
		})
	}
	assert.Equal(t, expectedPublished, c.posts)
}

func TestResumeThread(t *testing.T) {
	c := &fakeClient{failAfter: 3}
	m := multiposter.New(c, multiposter.AsThread())
	posts := getPosts(10)

	result := m.Publish(context.Background(), posts)
	expectedResult := &multiposter.PublishResult{}
	for i := range 3 {
		expectedResult.Published = append(expectedResult.Published, &client.PublishResult{
			Uri: uriOf(i),
			Cid: cidOf(i),
		})
	}
	expectedResult.Remaining = posts[3:]
	expectedResult.Error = errPublish
	assert.Equal(t, expectedResult, result)
	assert.Len(t, c.posts, 3)

	c.failAfter = -1

	result2 := m.Resume(context.Background(), result)
	expectedResult = &multiposter.PublishResult{}
	for i := range 10 {
		expectedResult.Published = append(expectedResult.Published, &client.PublishResult{
			Uri: uriOf(i),
			Cid: cidOf(i),
		})
	}
	assert.Equal(t, expectedResult, result2)
	var expectedPublished []publishedPost
	for i := range posts {
		postWithReply := *posts[i]
		if i > 0 {
			postWithReply.Reply = &bsky.FeedPost_ReplyRef{
				Parent: &atproto.RepoStrongRef{
					LexiconTypeID: "com.atproto.repo.strongRef",
					Cid:           cidOf(i - 1),
					Uri:           uriOf(i - 1),
				},
				Root: &atproto.RepoStrongRef{
					LexiconTypeID: "com.atproto.repo.strongRef",
					Cid:           cidOf(0),
					Uri:           uriOf(0),
				},
			}
		}
		expectedPublished = append(expectedPublished, publishedPost{
			cid:  cidOf(i),
			uri:  uriOf(i),
			post: &postWithReply,
		})
	}
	assert.Equal(t, expectedPublished, c.posts)
}

func getPosts(count int) []*bsky.FeedPost {
	converter := posts.NewConverter(posts.WithClock(&atptesting.FakeClock{Time: time.Date(2025, time.January, 2, 12, 34, 56, 789000000, time.UTC)}))
	var out []*bsky.FeedPost
	for i := range count {
		post := k3.NewPost().AddText(fmt.Sprintf("This is post #%d", i+1))
		out = append(out, converter.ToFeedPost(post))
	}
	return out
}

type fakeClient struct {
	posts     []publishedPost
	failAfter int
}

type publishedPost struct {
	cid  string
	uri  string
	post *bsky.FeedPost
}

var errPublish = errors.New("failure posting")

func (f *fakeClient) Publish(ctx context.Context, post *bsky.FeedPost) (*client.PublishResult, error) {
	if f.failAfter == 0 {
		return nil, errPublish
	}
	pp := publishedPost{
		cid:  cidOf(len(f.posts)),
		uri:  uriOf(len(f.posts)),
		post: post,
	}
	f.posts = append(f.posts, pp)
	if f.failAfter > 0 {
		f.failAfter--
	}
	return &client.PublishResult{
		Uri: pp.uri,
		Cid: pp.cid,
	}, nil
}

func cidOf(i int) string {
	return fmt.Sprintf("%d", i)
}

func uriOf(i int) string {
	return fmt.Sprintf("at://xxxx/yyyy/%d", i)
}

func (f *fakeClient) FindUserByHandle(ctx context.Context, handle string) (*client.UserData, error) {
	panic("unimplemented")
}

func (f *fakeClient) GetAccessToken(ctx context.Context) error {
	panic("unimplemented")
}
