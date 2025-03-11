package client

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/bluesky-social/indigo/api/atproto"
	"github.com/bluesky-social/indigo/lex/util"
	"github.com/bluesky-social/indigo/xrpc"
	"github.com/jtarrio/atp"
	"github.com/jtarrio/atp/posts"
	"github.com/lestrrat-go/jwx/v3/jwt"
)

// Client is an interface for talking to a Bluesky server.
type Client interface {
	// GetAccessToken authenticates to the server and retrieves authentication tokens, if needed.
	// If the client already has valid tokens, this operation is a no-op.
	GetAccessToken(ctx context.Context) error
	// Publish saves the given post in the user's timeline, returning the post's CID and URI.
	Publish(ctx context.Context, post *atp.Post) (*PublishResult, error)
	// FindUserByHandle returns the handle and DID of the user with the given handle.
	FindUserByHandle(ctx context.Context, handle string) (*UserData, error)
	// FindUserByDid returns the handle and DID of the user with the given handle.
	FindUserByDid(ctx context.Context, did string) (*UserData, error)
}

// PublishResult holds the result of the Publish method.
type PublishResult struct {
	// Uri contains the post's URI.
	Uri string
	// Cid contains the post's CID.
	Cid string
}

// UserData holds the result of the FindUserByXxx methods.
type UserData struct {
	// Handle contains the user's handle.
	Handle string
	// Did contains the user's DID.
	Did string
}

// New creates a new client instance, authenticated with the given username and password, and with the given options.
func New(identifier string, password string, options ...ClientOption) Client {
	client := &clientImpl{
		identifier: identifier,
		password:   password,
		clock:      atp.SystemClock(),
		converter:  posts.NewConverter(posts.WithClock(atp.SystemClock())),
		xrpc:       &xrpc.Client{Host: "https://bsky.social"},
	}
	for _, option := range options {
		option(client)
	}
	return client
}

// WithHost makes the client use a different host to connect to Bluesky.
func WithHost(host string) ClientOption {
	return func(p *clientImpl) {
		p.xrpc.Host = host
	}
}

// WithHttpClient makes the client use a different http.Client to connect to Bluesky.
func WithHttpClient(client *http.Client) ClientOption {
	return func(p *clientImpl) {
		p.xrpc.Client = client
	}
}

// WithClock makes the client use a different time source.
func WithClock(clock atp.Clock) ClientOption {
	return func(p *clientImpl) {
		p.clock = clock
		p.converter = posts.NewConverter(posts.WithClock(clock))
	}
}

// ClientOption is a modifier for NewClient.
type ClientOption func(*clientImpl)

type clientImpl struct {
	identifier string
	password   string
	clock      atp.Clock
	converter  *posts.Converter
	xrpc       *xrpc.Client
	xrpcMutex  sync.RWMutex
}

func (c *clientImpl) GetAccessToken(ctx context.Context) error {
	c.xrpcMutex.Lock()
	defer c.xrpcMutex.Unlock()
	now := c.clock.Now().Add(5 * time.Second)
	if c.xrpc.Auth != nil {
		if expired, err := isJwtExpired(c.xrpc.Auth.AccessJwt, now); err != nil {
			return err
		} else if !expired {
			return nil
		}
		if expired, err := isJwtExpired(c.xrpc.Auth.RefreshJwt, now); err != nil {
			return err
		} else if !expired {
			c.xrpc.Auth.AccessJwt = c.xrpc.Auth.RefreshJwt
			output, err := atproto.ServerRefreshSession(ctx, c.xrpc)
			if err != nil {
				return fmt.Errorf("could not refresh session: %w", err)
			}
			c.xrpc.Auth = &xrpc.AuthInfo{
				AccessJwt:  output.AccessJwt,
				RefreshJwt: output.RefreshJwt,
				Handle:     output.Handle,
				Did:        output.Did,
			}
			return nil
		}
	}
	input := &atproto.ServerCreateSession_Input{
		Identifier: c.identifier,
		Password:   c.password,
	}
	output, err := atproto.ServerCreateSession(ctx, c.xrpc, input)
	if err != nil {
		return fmt.Errorf("could not create session: %w", err)
	}
	c.xrpc.Auth = &xrpc.AuthInfo{
		AccessJwt:  output.AccessJwt,
		RefreshJwt: output.RefreshJwt,
		Handle:     output.Handle,
		Did:        output.Did,
	}
	return nil
}

func (c *clientImpl) Publish(ctx context.Context, post *atp.Post) (*PublishResult, error) {
	if err := c.GetAccessToken(ctx); err != nil {
		return nil, err
	}
	c.xrpcMutex.RLock()
	defer c.xrpcMutex.RUnlock()
	input := &atproto.RepoCreateRecord_Input{
		Collection: "app.bsky.feed.post",
		Record:     &util.LexiconTypeDecoder{Val: c.converter.ToFeedPost(post)},
		Repo:       c.xrpc.Auth.Did,
	}
	output, err := atproto.RepoCreateRecord(ctx, c.xrpc, input)
	if err != nil {
		return nil, fmt.Errorf("could not publish a post: %w", err)
	}
	result := &PublishResult{
		Uri: output.Uri,
		Cid: output.Cid,
	}
	return result, err
}

func (c *clientImpl) FindUserByHandle(ctx context.Context, username string) (*UserData, error) {
	if err := c.GetAccessToken(ctx); err != nil {
		return nil, err
	}
	c.xrpcMutex.RLock()
	defer c.xrpcMutex.RUnlock()
	output, err := atproto.IdentityResolveHandle(ctx, c.xrpc, username)
	if err != nil {
		return nil, fmt.Errorf("could not find username '%s': %w", username, err)
	}
	result := &UserData{
		Handle: username,
		Did:    output.Did,
	}
	return result, nil
}

func (c *clientImpl) FindUserByDid(ctx context.Context, did string) (*UserData, error) {
	if err := c.GetAccessToken(ctx); err != nil {
		return nil, err
	}
	c.xrpcMutex.RLock()
	defer c.xrpcMutex.RUnlock()
	output, err := atproto.IdentityResolveDid(ctx, c.xrpc, did)
	if err != nil {
		return nil, fmt.Errorf("could not find DID '%s': %w", did, err)
	}
	didDocument, ok := output.DidDoc.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("could not extract DID document: %+v", output.DidDoc)
	}
	rawAkas, found := didDocument["alsoKnownAs"]
	if !found {
		return nil, fmt.Errorf("the DID document does not contain an alsoKnownAs field")
	}
	akas, ok := rawAkas.([]any)
	if !ok {
		return nil, fmt.Errorf("the DID document contains a non-list alsoKnownAs field")
	}
	handle := ""
	for _, rawAka := range akas {
		if aka, ok := rawAka.(string); ok {
			if suffix, ok := strings.CutPrefix(aka, "at://"); ok {
				handle = suffix
				break
			}
		}
	}
	result := &UserData{
		Handle: handle,
		Did:    did,
	}
	return result, nil
}

func isJwtExpired(jwtString string, now time.Time) (bool, error) {
	if len(jwtString) == 0 {
		return true, nil
	}
	token, err := jwt.ParseString(jwtString, jwt.WithVerify(false), jwt.WithValidate(false))
	if err != nil {
		return false, err
	}
	exp, found := token.Expiration()
	return found && now.After(exp), nil
}
