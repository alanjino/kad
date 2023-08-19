package oryclient

import (
	"context"
	"strings"

	"github.com/intelops/go-common/logging"
	"github.com/kelseyhightower/envconfig"
	"github.com/kube-tarian/kad/server/pkg/credential"
	ory "github.com/ory/client-go"
	"golang.org/x/oauth2/clientcredentials"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// Config represents the configuration settings required for
// fetching ory entities from the vault
// also for integration with ORY and create a OryApiClient.
type Config struct {
	OryEntityName        string `envconfig:"ORY_ENTITY_NAME" required:"true"`
	CredentialIdentifier string `envconfig:"CRED_IDENTITY" required:"true"`
}

type Client struct {
	oryPAT string
	conn   *ory.APIClient
	log    logging.Logger
	oryURL string
}

type OryClient interface {
	GetSessionTokenFromContext(ctx context.Context) (string, error)
	Authorize(ctx context.Context, accessToken string) (context.Context, error)
	UnaryInterceptor(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error
}

// NewOryClient returns a OryClient interface
func NewOryClient(log logging.Logger) (OryClient, error) {
	cfg, err := getOryEnv()
	if err != nil {
		return nil, err
	}
	serviceCredential, err := credential.GetServiceUserCredential(context.Background(),
		cfg.OryEntityName, cfg.CredentialIdentifier)
	if err != nil {
		return nil, err
	}
	oryPAT := serviceCredential.AdditionalData["ORY_PAT"]
	oryURL := serviceCredential.AdditionalData["ORY_URL"]
	conn := NewOrySdk(log, oryURL)
	return &Client{
		oryPAT: oryPAT,
		conn:   conn,
		log:    log,
		oryURL: oryURL,
	}, nil
}

func getOryEnv() (*Config, error) {
	cfg := &Config{}
	if err := envconfig.Process("", cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}

// NewOrySdk creates a oryAPIClient using the oryURL
// and returns it
func NewOrySdk(log logging.Logger, oryURL string) *ory.APIClient {
	log.Info("creating a ory client")
	config := ory.NewConfiguration()
	config.Servers = ory.ServerConfigurations{{
		URL: oryURL,
	}}

	return ory.NewAPIClient(config)
}

// GetSessionTokenFromContext fetches the session token from the context
// and returns the token and nil for the error.
// But if any error occurs while fetching the token it returns an empty string and an error.
func (c *Client) GetSessionTokenFromContext(ctx context.Context) (string, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return "", status.Error(codes.Unauthenticated, "Failed to get metadata from context")
	}
	bearerToken := md.Get("authorization")
	if len(bearerToken) == 0 || len(strings.Split(bearerToken[0], " ")) != 2 {
		return "", status.Error(codes.Unauthenticated, "No access token provided")
	}
	accessToken := bearerToken[0]
	if len(accessToken) < 8 || accessToken[:7] != "Bearer " {
		return "", status.Error(codes.Unauthenticated, "Invalid access token")
	}
	return accessToken[7:], nil
}

// Authorize checks whether the accesstoken is valid or Invalid using the ory.APIClient
// It checks token is active or not active
// If token is active its a valid token
// If token is not active its a invalid token
func (c *Client) Authorize(ctx context.Context, accessToken string) (context.Context, error) {
	ctx = context.WithValue(ctx, ory.ContextAccessToken, c.oryPAT)
	sessionInfo, _, err := c.conn.IdentityApi.GetSession(ctx, accessToken).Execute()
	if err != nil {
		c.log.Errorf("Error occured while getting session info for session id - "+accessToken+"+\nError - %v", err.Error())
		return ctx, status.Errorf(codes.Unauthenticated, "Failed to introspect session id - %v", err)
	}
	c.log.Infof("session id: %v", sessionInfo.Id)
	if !sessionInfo.GetActive() {
		c.log.Errorf("Error occured while getting session info for session id - "+accessToken+"+\nError - %v", err.Error())
		return ctx, status.Error(codes.Unauthenticated, "session id is not active")
	}
	return ctx, nil
}
func (c *Client) GetOryTokenUrl() string {
	tokenUrl := c.oryURL + "/oauth2/token"
	return tokenUrl
}
func (c *Client) GetOauthToken(ctx context.Context) (context.Context, error) {
	clientid, secret, err := credential.GetIamOauthCredential(ctx)
	if err != nil {
		c.log.Errorf("error while getting clientid and secret from vault: %v", err.Error())
		return ctx, err
	}

	conf := &clientcredentials.Config{
		ClientID:     clientid,
		ClientSecret: secret,
		Scopes:       []string{"openid email offline"},
		TokenURL:     c.GetOryTokenUrl(),
	}
	at, err := conf.Token(ctx)
	if err != nil {
		c.log.Errorf("error while fetching oauth token from oryapiclient ERROR: %v", err.Error())
		return ctx, err
	}
	md := metadata.Pairs("oauth_token", at.AccessToken,
		"ory_url", c.oryURL,
		"ory_pat", c.oryPAT,
	)
	newCtx := metadata.NewOutgoingContext(ctx, md)

	return newCtx, nil
}

func (c *Client) UnaryInterceptor(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
	newCtx, err := c.GetOauthToken(ctx)
	if err != nil {
		return err
	}
	return invoker(newCtx, method, req, reply, cc, opts...)
}