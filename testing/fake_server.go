package testing

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"time"

	"github.com/bluesky-social/indigo/api/atproto"
	"github.com/bluesky-social/indigo/api/bsky"
	"github.com/bluesky-social/indigo/atproto/identity"
	"github.com/bluesky-social/indigo/xrpc"
	"github.com/jtarrio/atp"
	"github.com/lestrrat-go/jwx/v3/jwt"
)

// NewFakeServer returns a fake Bluesky server for testing.
func NewFakeServer(options ...FakeServerOption) *FakeServer {
	fs := &FakeServer{
		clock:   atp.SystemClock(),
		users:   map[string]string{},
		methods: map[methodKey]methodFunc{},
	}
	for _, option := range options {
		option(fs)
	}
	fs.server = httptest.NewServer(fs)
	fs.Register(NewFunction(xrpc.Procedure, "com.atproto.server.createSession", fs.serverCreateSession))
	fs.Register(NewCommand(xrpc.Procedure, "com.atproto.server.refreshSession", fs.serverRefreshSession))
	fs.Register(NewFunction(xrpc.Procedure, "com.atproto.repo.createRecord", fs.repoCreateRecord))
	fs.Register(NewCommand(xrpc.Query, "com.atproto.identity.resolveHandle", fs.identityResolveHandle))
	return fs
}

// WithClock makes the FakeServer use the given clock.
func WithClock(clock atp.Clock) FakeServerOption {
	return func(fs *FakeServer) {
		fs.clock = clock
	}
}

type FakeServerOption func(*FakeServer)

// FakeServer is a fake Bluesky server for testing.
type FakeServer struct {
	// Calls contains information about all the methods that were called in the fake server.
	Calls []Call
	// Posts contains all the posts that were published to the server.
	Posts    []Post
	clock    atp.Clock
	users    map[string]string
	methods  map[methodKey]methodFunc
	userDids []identity.DIDDocument
	server   *httptest.Server
}

// Call contains information about a method call.
type Call struct {
	Method string
	User   *string
	Params map[string][]string
	Input  any
}

// Post contains information about a published Bluesky post.
type Post struct {
	Repo   string
	Rkey   string
	Record *bsky.FeedPost
}

// URL returns the server's URL.
func (f *FakeServer) URL() string {
	return f.server.URL
}

// Close shuts down the server.
func (f *FakeServer) Close() {
	f.server.Close()
}

// AddUser adds a user that can be authenticated to the server.
func (f *FakeServer) AddUser(identifier string, password string) *FakeServer {
	f.users[identifier] = password
	return f
}

// AddUserDid adds information about another user on the server.
func (f *FakeServer) AddUserDid(userDid *identity.DIDDocument) *FakeServer {
	f.userDids = append(f.userDids, *userDid)
	return f
}

// Register adds a method to the server.
func (f *FakeServer) Register(method *FakeServerMethod) *FakeServer {
	f.methods[method.key] = method.fn
	return f
}

func (f *FakeServer) serverCreateSession(user *string, params map[string][]string, input *atproto.ServerCreateSession_Input) (*atproto.ServerCreateSession_Output, error) {
	if pass, found := f.users[input.Identifier]; !found || pass != input.Password {
		return nil, fmt.Errorf("user not found: %s", input.Identifier)
	}
	accessJwt, err := createJwt(input.Identifier, f.clock.Now().Add(2*time.Minute))
	if err != nil {
		return nil, fmt.Errorf("error generating accessJwt: %w", err)
	}
	refreshJwt, err := createJwt(input.Identifier, f.clock.Now().Add(1*time.Hour))
	if err != nil {
		return nil, fmt.Errorf("error generating refreshJwt: %w", err)
	}
	output := &atproto.ServerCreateSession_Output{
		AccessJwt:  accessJwt,
		RefreshJwt: refreshJwt,
		Handle:     input.Identifier,
		Did:        "did:web:" + input.Identifier,
	}
	return output, nil
}

func (f *FakeServer) serverRefreshSession(user *string, params map[string][]string) (*atproto.ServerRefreshSession_Output, error) {
	if user == nil {
		return nil, fmt.Errorf("no valid JWT in request")
	}
	accessJwt, err := createJwt(*user, f.clock.Now().Add(2*time.Minute))
	if err != nil {
		return nil, fmt.Errorf("error generating accessJwt: %w", err)
	}
	refreshJwt, err := createJwt(*user, f.clock.Now().Add(1*time.Hour))
	if err != nil {
		return nil, fmt.Errorf("error generating refreshJwt: %w", err)
	}
	output := &atproto.ServerRefreshSession_Output{
		AccessJwt:  accessJwt,
		RefreshJwt: refreshJwt,
		Handle:     *user,
		Did:        "did:web:" + *user,
	}
	return output, nil
}

func (f *FakeServer) repoCreateRecord(user *string, params map[string][]string, input *atproto.RepoCreateRecord_Input) (*atproto.RepoCreateRecord_Output, error) {
	if user == nil {
		return nil, fmt.Errorf("no valid JWT in request")
	}
	if input.Collection != "app.bsky.feed.post" {
		return nil, fmt.Errorf("invalid collection: %s", input.Collection)
	}
	rkey := ""
	if input.Rkey != nil {
		rkey = *input.Rkey
	} else {
		rkey = fmt.Sprintf("%x", len(f.Posts))
	}
	f.Posts = append(f.Posts, Post{
		Repo:   input.Repo,
		Rkey:   rkey,
		Record: input.Record.Val.(*bsky.FeedPost),
	})
	output := &atproto.RepoCreateRecord_Output{
		Cid:    rkey,
		Commit: &atproto.RepoDefs_CommitMeta{},
		Uri:    fmt.Sprintf("https://uri/%s/%s", input.Repo, rkey),
	}
	return output, nil
}

func (f *FakeServer) identityResolveHandle(user *string, params map[string][]string) (*atproto.IdentityResolveHandle_Output, error) {
	handle, found := params["handle"]
	if !found {
		return nil, errors.New("handle not specified")
	}
	for i := range f.userDids {
		userDid := &f.userDids[i]
		for _, aka := range userDid.AlsoKnownAs {
			if suffix, found := strings.CutPrefix(aka, "at://"); found {
				if suffix == handle[0] {
					result := &atproto.IdentityResolveHandle_Output{
						Did: userDid.DID.String(),
					}
					return result, nil
				}
				break
			}
		}
	}
	return nil, fmt.Errorf("user with handle '%s' not found", handle[0])
}

func (f *FakeServer) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	if len(req.URL.Path) < 6 || req.URL.Path[0:6] != "/xrpc/" {
		outputError(rw, "invalid method", fmt.Errorf("invalid path: %s", req.URL.Path))
		return
	}
	key := methodKey{
		httpMethod: req.Method,
		name:       req.URL.Path[6:],
	}
	if method, found := f.methods[key]; found {
		method(rw, req, f)
		return
	}
	outputError(rw, "method not registered", fmt.Errorf("not registered: %s %s", key.httpMethod, key.name))
}

type ProcedureDefinition[O any] func(user *string, params map[string][]string) (*O, error)

// NewCommand is used to define a method that doesn't take an input object.
func NewCommand[O any](reqType xrpc.XRPCRequestType, name string, def ProcedureDefinition[O]) *FakeServerMethod {
	key := methodKey{
		name: name,
	}
	if reqType == xrpc.Query {
		key.httpMethod = http.MethodGet
	} else {
		key.httpMethod = http.MethodPost
	}
	fn := func(rw http.ResponseWriter, req *http.Request, fs *FakeServer) {
		user := getUser(req, fs.clock.Now())
		params := req.URL.Query()
		if len(params) == 0 {
			params = nil
		}
		fs.Calls = append(fs.Calls, Call{
			Method: name,
			User:   user,
			Params: params,
			Input:  nil,
		})
		output, err := def(user, params)
		if err != nil {
			outputError(rw, "method returned error", err)
			return
		}
		b, err := json.Marshal(output)
		if err != nil {
			outputError(rw, "could not convert response to JSON", err)
			return
		}
		rw.Write(b)
	}
	return &FakeServerMethod{key: key, fn: fn}
}

type FunctionDefinition[I any, O any] func(user *string, params map[string][]string, input *I) (*O, error)

// NewFunction is used to define a method that takes an input object.
func NewFunction[I any, O any](reqType xrpc.XRPCRequestType, name string, def FunctionDefinition[I, O]) *FakeServerMethod {
	key := methodKey{
		name: name,
	}
	if reqType == xrpc.Query {
		key.httpMethod = http.MethodGet
	} else {
		key.httpMethod = http.MethodPost
	}
	fn := func(rw http.ResponseWriter, req *http.Request, fs *FakeServer) {
		user := getUser(req, fs.clock.Now())
		params := req.URL.Query()
		if len(params) == 0 {
			params = nil
		}
		var input *I
		defer req.Body.Close()
		err := json.NewDecoder(req.Body).Decode(&input)
		if err != nil {
			outputError(rw, "could not convert request to JSON", err)
			return
		}
		fs.Calls = append(fs.Calls, Call{
			Method: name,
			User:   user,
			Params: params,
			Input:  input,
		})
		output, err := def(user, params, input)
		if err != nil {
			outputError(rw, "method returned error", err)
			return
		}
		b, err := json.Marshal(output)
		if err != nil {
			outputError(rw, "could not convert response to JSON", err)
			return
		}
		rw.Write(b)
	}
	return &FakeServerMethod{key: key, fn: fn}
}

// FakeServerMethod contains a method that can be registered using Register.
//
// Use NewCommand and NewFunction to create FakeServerMethods.
type FakeServerMethod struct {
	key methodKey
	fn  methodFunc
}

type methodKey struct {
	httpMethod string
	name       string
}

type methodFunc func(rw http.ResponseWriter, req *http.Request, fs *FakeServer)

func outputError(rw http.ResponseWriter, str string, err error) {
	rw.WriteHeader(400)
	xrpcErr := xrpc.XRPCError{
		ErrStr:  str,
		Message: err.Error(),
	}
	b, err := json.Marshal(xrpcErr)
	if err == nil {
		rw.Write(b)
	}
}

func getUser(req *http.Request, now time.Time) *string {
	hdr := req.Header.Get("authorization")
	if token, found := strings.CutPrefix(hdr, "Bearer "); found {
		iss, err := decodeJwt(token, now)
		if err == nil {
			user, _ := strings.CutPrefix(iss, "did:web:")
			return &user
		}
	}
	return nil
}

func decodeJwt(jwtStr string, now time.Time) (string, error) {
	token, err := jwt.ParseString(jwtStr, jwt.WithVerify(false), jwt.WithValidate(false))
	if err != nil {
		return "", err
	}
	if exp, found := token.Expiration(); found && exp.Before(now) {
		return "", fmt.Errorf("token expired on %s", exp)
	}
	if iss, found := token.Issuer(); !found {
		return "", errors.New("token does not have issuer information")
	} else if len(iss) <= 8 || iss[0:8] != "did:web:" {
		return "", fmt.Errorf("token does not have a valid issuer: %s", iss)
	} else {
		return iss, nil
	}
}

func createJwt(identifier string, exp time.Time) (string, error) {
	token, err := jwt.NewBuilder().Issuer("did:web:" + identifier).Expiration(exp).Build()
	if err != nil {
		return "", err
	}
	b, err := jwt.Sign(token, jwt.WithInsecureNoSignature())
	if err != nil {
		return "", err
	}
	return string(b), nil
}
