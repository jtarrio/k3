package client_test

import (
	"context"
	"testing"
	"time"

	"github.com/bluesky-social/indigo/api/atproto"
	"github.com/bluesky-social/indigo/atproto/identity"
	"github.com/bluesky-social/indigo/lex/util"
	"github.com/jtarrio/atp"
	"github.com/jtarrio/atp/client"
	"github.com/jtarrio/atp/posts"
	atptesting "github.com/jtarrio/atp/testing"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConnect(t *testing.T) {
	username := "testuser"
	password := "testpass"
	clock := &atptesting.FakeClock{Time: time.Date(2025, time.January, 2, 12, 34, 56, 0, time.UTC)}
	fakeServer := atptesting.NewFakeServer(atptesting.WithClock(clock))
	fakeServer.AddUser(username, password)
	defer fakeServer.Close()

	ctx := context.Background()
	c := client.New(username, password, client.WithHost(fakeServer.URL()), client.WithClock(clock))

	// We can get access tokens
	err := c.GetAccessToken(ctx)
	assert.NoError(t, err)
	expectedCalls := []atptesting.Call{{
		Method: "com.atproto.server.createSession",
		Input: &atproto.ServerCreateSession_Input{
			Identifier: username,
			Password:   password,
		},
	}}
	assert.Equal(t, expectedCalls, fakeServer.Calls)
	fakeServer.Calls = nil

	// Access token still valid
	clock.Time = clock.Time.Add(1 * time.Minute)
	err = c.GetAccessToken(ctx)
	assert.NoError(t, err)
	assert.Empty(t, fakeServer.Calls)

	// We can renew the access token
	clock.Time = clock.Time.Add(5 * time.Minute)
	err = c.GetAccessToken(ctx)
	assert.NoError(t, err)
	expectedCalls = []atptesting.Call{{
		Method: "com.atproto.server.refreshSession",
		User:   &username,
	}}
	assert.Equal(t, expectedCalls, fakeServer.Calls)
	fakeServer.Calls = nil

	// We can get new tokens if the refresh token expires
	clock.Time = clock.Time.Add(2 * time.Hour)
	err = c.GetAccessToken(ctx)
	assert.NoError(t, err)
	expectedCalls = []atptesting.Call{{
		Method: "com.atproto.server.createSession",
		Input: &atproto.ServerCreateSession_Input{
			Identifier: username,
			Password:   password,
		},
	}}
	assert.Equal(t, expectedCalls, fakeServer.Calls)
}

func TestPublish(t *testing.T) {
	username := "testuser"
	password := "testpass"
	clock := &atptesting.FakeClock{Time: time.Date(2025, time.January, 2, 12, 34, 56, 0, time.UTC)}
	fakeServer := atptesting.NewFakeServer(atptesting.WithClock(clock))
	fakeServer.AddUser(username, password)
	defer fakeServer.Close()

	ctx := context.Background()
	c := client.New(username, password, client.WithHost(fakeServer.URL()), client.WithClock(clock))
	err := c.GetAccessToken(ctx)
	require.NoError(t, err)
	fakeServer.Calls = nil

	post := atp.NewPost().SetCreationTime(clock.Now()).AddText(`y así como don Quijote los vio, dijo a su escudero`)
	feedPost := posts.NewConverter(posts.WithClock(clock)).ToFeedPost(post)
	_, err = c.Publish(ctx, feedPost)
	require.NoError(t, err)

	expectedCalls := []atptesting.Call{{
		Method: "com.atproto.repo.createRecord",
		User:   &username,
		Input: &atproto.RepoCreateRecord_Input{
			Collection: "app.bsky.feed.post",
			Record:     &util.LexiconTypeDecoder{Val: feedPost},
			Repo:       "did:web:" + username,
		},
	}}
	assert.Equal(t, expectedCalls, fakeServer.Calls)

	expectedPosts := []atptesting.Post{
		{
			Repo:   "did:web:" + username,
			Rkey:   "0",
			Record: feedPost,
		},
	}
	assert.Equal(t, expectedPosts, fakeServer.Posts)
}

func TestSearchUsersByHandle(t *testing.T) {
	username := "testuser"
	password := "testpass"
	clock := &atptesting.FakeClock{Time: time.Date(2025, time.January, 2, 12, 34, 56, 0, time.UTC)}
	fakeServer := atptesting.NewFakeServer(atptesting.WithClock(clock))
	fakeServer.AddUser(username, password)
	fakeServer.AddUserDid(&identity.DIDDocument{
		DID:         "did:web:username.xyz",
		AlsoKnownAs: []string{"at://username"},
	})
	defer fakeServer.Close()

	ctx := context.Background()
	c := client.New(username, password, client.WithHost(fakeServer.URL()), client.WithClock(clock))
	err := c.GetAccessToken(ctx)
	require.NoError(t, err)
	fakeServer.Calls = nil

	result, err := c.FindUserByHandle(ctx, "username")
	require.NoError(t, err)
	expected := &client.UserData{
		Handle: "username",
		Did:    "did:web:username.xyz",
	}
	assert.Equal(t, expected, result)

	_, err = c.FindUserByHandle(ctx, "xxxxx")
	assert.Error(t, err)
}

func TestSearchUsersByDid(t *testing.T) {
	username := "testuser"
	password := "testpass"
	clock := &atptesting.FakeClock{Time: time.Date(2025, time.January, 2, 12, 34, 56, 0, time.UTC)}
	fakeServer := atptesting.NewFakeServer(atptesting.WithClock(clock))
	fakeServer.AddUser(username, password)
	fakeServer.AddUserDid(&identity.DIDDocument{
		DID:         "did:web:username.xyz",
		AlsoKnownAs: []string{"at://username"},
	})
	defer fakeServer.Close()

	ctx := context.Background()
	c := client.New(username, password, client.WithHost(fakeServer.URL()), client.WithClock(clock))
	err := c.GetAccessToken(ctx)
	require.NoError(t, err)
	fakeServer.Calls = nil

	result, err := c.FindUserByDid(ctx, "did:web:username.xyz")
	require.NoError(t, err)
	expected := &client.UserData{
		Handle: "username",
		Did:    "did:web:username.xyz",
	}
	assert.Equal(t, expected, result)

	_, err = c.FindUserByDid(ctx, "did:web:xxxxxx")
	assert.Error(t, err)
}
