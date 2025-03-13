package multiposter

import (
	"context"

	"github.com/bluesky-social/indigo/api/atproto"
	"github.com/bluesky-social/indigo/api/bsky"
	"github.com/jtarrio/k3/client"
)

// Multiposter is an interface to post multiple messages at once, either as a sequence of individual messages or as a thread.
type Multiposter interface {
	// Publish sends a series of posts to Bluesky.
	Publish(ctx context.Context, posts []*bsky.FeedPost) *PublishResult
	// Resume sends a series of posts to Bluesky, picking up where a previous call to Publish failed.
	Resume(ctx context.Context, previousResult *PublishResult) *PublishResult
}

// PublishResult contains the result of a Publish or Resume operation.
type PublishResult struct {
	// Published contains the result of publishing each individual post. If the operation failed at some point,
	// the results of the successful operations are here. Posts are always published serially, so if you tried
	// to publish 10 posts and post number 4 failed, you will see the results of posts 1-3 here.
	Published []*client.PublishResult
	// Remaining contains the list of all posts that are left to be published. As an example, if you tried
	// to publish 10 posts and post number 4 failed, you will see posts 4-10 here.
	Remaining []*bsky.FeedPost
	// Error contains the error returned by the last publish operation, if any.
	Error error
}

// New creates a new Multiposter with the given client and options.
//
// By default, posts are published as a sequence of posts, but you can use the AsThread option to publish them as a thread.
func New(client client.Client, options ...MultiposterOption) Multiposter {
	m := &multiposter{client: client}
	for _, option := range options {
		option(m)
	}
	return m
}

// AsThread indicates that the posts will be published as a thread, with the first post as its root.
func AsThread() MultiposterOption {
	return func(m *multiposter) {
		m.threaded = true
	}
}

// AsSequence indicates that the posts will be published as a sequence of individual posts.
//
// The first post in the list will be published first, but note that the actual ordering
// of the posts in the Bluesky view may depend on the order of the dates on the individual posts,
// if it differs from the posting order.
//
// This is the default option.
func AsSequence() MultiposterOption {
	return func(m *multiposter) {
		m.threaded = false
	}
}

type MultiposterOption func(*multiposter)

type multiposter struct {
	client   client.Client
	threaded bool
}

func (m multiposter) Publish(ctx context.Context, posts []*bsky.FeedPost) *PublishResult {
	return m.doPublish(ctx, posts, nil)
}

func (m multiposter) Resume(ctx context.Context, previousResult *PublishResult) *PublishResult {
	return m.doPublish(ctx, previousResult.Remaining, previousResult.Published)
}

func (m multiposter) doPublish(ctx context.Context, posts []*bsky.FeedPost, previousResults []*client.PublishResult) *PublishResult {
	var threadRoot, threadParent *client.PublishResult
	if len(previousResults) > 0 {
		threadRoot = previousResults[0]
		threadParent = previousResults[len(previousResults)-1]
	}
	result := &PublishResult{Published: previousResults}
	for i := range posts {
		var thisPost *bsky.FeedPost
		if m.threaded && threadParent != nil && threadRoot != nil {
			thisPost = setReplyField(posts[i], threadParent, threadRoot)
		} else {
			thisPost = posts[i]
		}
		singleResult, err := m.client.Publish(ctx, thisPost)
		if err != nil {
			result.Remaining = posts[i:]
			result.Error = err
			return result
		}
		threadParent = singleResult
		if threadRoot == nil {
			threadRoot = singleResult
		}
		result.Published = append(result.Published, singleResult)
	}
	return result
}

func setReplyField(thisPost *bsky.FeedPost, threadParent *client.PublishResult, threadRoot *client.PublishResult) *bsky.FeedPost {
	postCopy := *thisPost
	postCopy.Reply = &bsky.FeedPost_ReplyRef{
		Parent: &atproto.RepoStrongRef{
			LexiconTypeID: "com.atproto.repo.strongRef",
			Cid:           threadParent.Cid,
			Uri:           threadParent.Uri,
		},
		Root: &atproto.RepoStrongRef{
			LexiconTypeID: "com.atproto.repo.strongRef",
			Cid:           threadRoot.Cid,
			Uri:           threadRoot.Uri,
		},
	}
	return &postCopy
}
