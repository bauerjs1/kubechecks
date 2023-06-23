package github_client

import (
	"context"
	"net/http"
	"sync"

	"github.com/google/go-github/github"
	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
	"github.com/zapier/kubechecks/pkg/repo"
	"golang.org/x/oauth2"
)

var githubClient *Client
var githubTokenUser string
var once sync.Once // used to ensure we don't reauth this

type Client struct {
	*github.Client
}

func GetGithubClient() (*Client, string) {
	once.Do(func() {
		githubClient = createGithubClient()
		githubTokenUser = getTokenUser()
	})
	return githubClient, githubTokenUser
}

// We require a username to use with git locally, so get the current auth'd user
func getTokenUser() string {
	user, _, err := githubClient.Users.Get(context.Background(), "")
	if err != nil {
		if err != nil {
			log.Fatal().Err(err).Msg("could not create Github token user")
		}
	}
	return *user.Name
}

// Create a new github client using the auth token provided. We
// can't validate the token at this point, so if it exists we assume it works
func createGithubClient() *Client {
	// Initialize the GitLab client with access token
	t := viper.GetString("vcs-token")
	if t == "" {
		log.Fatal().Msg("github token needs to be set")
	}
	log.Debug().Msgf("Token Length - %d", len(t))
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: t},
	)
	tc := oauth2.NewClient(ctx, ts)
	c := github.NewClient(tc) // If this has failed, we'll catch it on first call

	return &Client{c}
}

func (c *Client) VerifyHook(secret string, p echo.Context) error {
	// Github provides the SHA256 of the secret + payload body, so we extract the body and compare
	if secret != "" {
		_, err := github.ValidatePayload(p.Request(), []byte(secret))
		if err != nil {
			return p.String(http.StatusUnauthorized, "Unauthorized")
		}
	}

	return nil // Success
}

func (c *Client) ParseHook(r *http.Request, payload []byte) (interface{}, error) {
	return github.ParseWebHook(github.WebHookType(r), payload)
}

func (c *Client) CreateRepo(ctx context.Context, payload interface{}) (*repo.Repo, error) {
	switch p := payload.(type) {
	case *github.PullRequestEvent:
		return buildRepoFromEvent(p), nil
	default:
		return nil, nil
	}
}

func buildRepoFromEvent(event *github.PullRequestEvent) *repo.Repo {
	return &repo.Repo{
		BaseRef:       *event.PullRequest.Base.Ref,
		HeadRef:       *event.PullRequest.Head.Ref,
		DefaultBranch: *event.Repo.DefaultBranch,
		CloneURL:      *event.Repo.CloneURL,
		OwnerName:     event.Repo.GetFullName(),
		Name:          event.Repo.GetName(),
		CheckID:       int(*event.PullRequest.Number),
		SHA:           *event.PullRequest.Head.SHA,
		Username:      *event.Sender.Login,
		Email:         *event.Sender.Email,
	}
}
