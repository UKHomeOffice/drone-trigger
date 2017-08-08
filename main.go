package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path"
	"strconv"
	"strings"

	"github.com/drone/drone-go/drone"
	"github.com/urfave/cli"
	"golang.org/x/oauth2"
)

// Version is set at compile time, passing -ldflags "-X main.Version=<build version>"
var Version string

func main() {
	app := cli.NewApp()
	app.Name = "drone-trigger"
	app.Author = "Vaidas Jablonskis <jablonskis@gmail.com>"
	app.Version = Version
	app.Usage = "trigger drone builds or deployments"

	app.Flags = []cli.Flag{
		cli.BoolFlag{
			Name:   "verbose",
			Usage:  "verbose output",
			EnvVar: "VERBOSE,PLUGIN_VERBOSE",
		},
		cli.StringFlag{
			Name:   "drone-server, s",
			Usage:  "drone server `URL`",
			EnvVar: "DRONE_SERVER,PLUGIN_DRONE_SERVER",
		},
		cli.StringFlag{
			Name:   "drone-token, t",
			Usage:  "drone auth `TOKEN`",
			EnvVar: "DRONE_TOKEN,PLUGIN_DRONE_TOKEN",
		},
		cli.StringFlag{
			Name:   "repo, r",
			Usage:  "`REPO`, eg. foo/bar",
			EnvVar: "REPO,PLUGIN_REPO",
		},
		cli.StringFlag{
			Name:   "commit, c",
			Usage:  "filter by commit sha",
			EnvVar: "FILTER_COMMIT,PLUGIN_COMMIT",
		},
		cli.StringFlag{
			Name:   "tag",
			Usage:  "filter by tag",
			EnvVar: "FILTER_TAG,PLUGIN_TAG",
		},
		cli.StringFlag{
			Name:   "branch, b",
			Usage:  "filter by branch",
			EnvVar: "FILTER_BRANCH,PLUGIN_BRANCH",
		},
		cli.StringFlag{
			Name:   "status",
			Usage:  "filter by build status",
			EnvVar: "FILTER_STATUS,PLUGIN_STATUS",
			Value:  "success",
		},
		cli.IntFlag{
			Name:   "number",
			Usage:  "filter by build number",
			EnvVar: "FILTER_NUMBER,PLUGIN_NUMBER",
		},
		cli.StringFlag{
			Name:   "event",
			Usage:  "filter by trigger event",
			EnvVar: "FILTER_EVENT,PLUGIN_EVENT",
		},
		cli.StringFlag{
			Name:   "deployed-to",
			Usage:  "filter by environment deployed to",
			EnvVar: "FILTER_DEPLOYED_TO,PLUGIN_DEPLOYED_TO",
		},
		cli.StringFlag{
			Name:   "deploy-to, d",
			Usage:  "environment to deploy to, if set a deployment event will be triggered",
			EnvVar: "DEPLOY_TO,PLUGIN_DEPLOY_TO",
		},
		cli.StringSliceFlag{
			Name:   "param, p",
			Usage:  "custom parameters to include in the trigger in KEY=value format",
			EnvVar: "PARAMS,PLUGIN_PARAMS",
		},
	}

	app.Action = run
	app.Run(os.Args)
}

// Trigger represent drone trigger.
type Trigger struct {
	client drone.Client
}

func newTrigger(ctx *cli.Context) *Trigger {
	oauthCfg := &oauth2.Config{}
	httpClient := oauthCfg.Client(
		context.Background(),
		&oauth2.Token{
			AccessToken: ctx.String("drone-token"),
		},
	)

	return &Trigger{
		client: drone.NewClient(ctx.String("drone-server"), httpClient),
	}
}

func run(ctx *cli.Context) error {
	// Exit if required flags are not set
	if !ctx.IsSet("drone-server") && !isAnyEnvSet("DRONE_SERVER", "PLUGIN_DRONE_SERVER") {
		if err := cli.ShowAppHelp(ctx); err != nil {
			return cli.NewExitError(err, 1)
		}
		return cli.NewExitError("error: drone server is not set", 3)
	}
	if !ctx.IsSet("drone-token") && !isAnyEnvSet("DRONE_TOKEN", "PLUGIN_DRONE_TOKEN") {
		if err := cli.ShowAppHelp(ctx); err != nil {
			return cli.NewExitError(err, 1)
		}
		return cli.NewExitError("error: drone token is not set", 3)
	}
	if !ctx.IsSet("repo") && !isAnyEnvSet("REPO", "PLUGIN_REPO") {
		if err := cli.ShowAppHelp(ctx); err != nil {
			return cli.NewExitError(err, 1)
		}
		return cli.NewExitError("error: repo is not set", 3)
	}

	loneFilters := 0

	if ctx.IsSet("tag") || isAnyEnvSet("FILTER_TAG", "PLUGIN_TAG") {
		loneFilters++
	}

	if ctx.IsSet("branch") || isAnyEnvSet("FILTER_BRANCH", "PLUGIN_BRANCH") {
		loneFilters++
	}

	if ctx.IsSet("commit") || isAnyEnvSet("FILTER_COMMIT", "PLUGIN_COMMIT") {
		loneFilters++
	}

	if ctx.IsSet("deployed-to") || isAnyEnvSet("FILTER_DEPLOYED_TO", "PLUGIN_DEPLOYED_TO") {
		loneFilters++
	}

	if loneFilters > 1 {
		return cli.NewExitError("error: tag, branch, commit or deployed-to cannot be set at the same time, pick one filter", 3)
	}

	t := newTrigger(ctx)
	build, err := t.findBuild(ctx)
	if err != nil {
		return cli.NewExitError(err.Error(), 1)
	}
	if build == nil {
		return cli.NewExitError("No previous builds found", 1)
	}

	params := parsePairs(ctx.StringSlice("param"))
	var newBuild *drone.Build
	owner, repo, err := parseRepo(ctx.String("repo"))
	if err != nil {
		return cli.NewExitError(err.Error(), 1)
	}

	if ctx.IsSet("deploy-to") || isAnyEnvSet("DEPLOY_TO", "PLUGIN_DEPLOY_TO") {
		b, err := t.client.Deploy(owner, repo, build.Number, ctx.String("deploy-to"), params)
		if err != nil {
			return cli.NewExitError(err.Error(), 1)
		}
		newBuild = b
	} else {
		b, err := t.client.BuildStart(owner, repo, build.Number, params)
		if err != nil {
			return cli.NewExitError(err.Error(), 1)
		}
		newBuild = b
	}

	newBuildURL := path.Join(ctx.String("drone-server"), ctx.String("repo"), strconv.Itoa(newBuild.Number))
	fmt.Fprintf(os.Stderr, "Follow new build status at: %s\n", newBuildURL)

	if ctx.Bool("verbose") {
		j, err := json.MarshalIndent(newBuild, "", "  ")
		if err != nil {
			return cli.NewExitError(err.Error(), 1)
		}
		fmt.Println(string(j))
	}
	return nil
}

// findBuild requests an array of previous repo builds from drone API and finds
// the one that matches given filters
func (t *Trigger) findBuild(ctx *cli.Context) (*drone.Build, error) {
	user, repo, err := parseRepo(ctx.String("repo"))
	if err != nil {
		return nil, err
	}
	builds, err := t.client.BuildList(user, repo)
	if err != nil {
		return nil, err
	}
	for _, b := range builds {
		if match(ctx, b) {
			return b, nil
		}
	}
	return nil, nil
}

// match returns true for the first build that matches cli filter flags.
func match(ctx *cli.Context, build *drone.Build) bool {
	if ctx.String("status") != build.Status {
		return false
	}
	if ctx.IsSet("event") && ctx.String("event") != build.Event {
		return false
	}
	// Build number always takes precedence
	if ctx.IsSet("number") {
		if ctx.Int("number") == build.Number {
			return true
		}
		return false
	}
	if ctx.IsSet("commit") {
		if ctx.String("commit") == build.Commit {
			return true
		}
		return false
	}
	if ctx.IsSet("tag") {
		if "refs/tags/"+ctx.String("tag") == build.Ref {
			return true
		}
		return false
	}
	if ctx.IsSet("deployed-to") {
		if ctx.String("deployed-to") == build.Deploy {
			return true
		}
		return false
	}
	// Matching on branch does not make sense for pull_request events, because
	// the branch for PR events is the base branch.
	if ctx.IsSet("branch") {
		if ctx.String("branch") == build.Branch && build.Event != "pull_request" {
			return true
		}
		return false
	}
	// Return latest successful build if no specific filters are set
	return true
}

// parseRepo parses a owner/repo string into two parts
func parseRepo(s string) (owner, repo string, err error) {
	var parts = strings.Split(s, "/")
	if len(parts) != 2 {
		err = fmt.Errorf("error: invalid or missing repository. eg octocat/hello-world")
		return
	}
	owner = parts[0]
	repo = parts[1]
	return
}

// parsePairs parses an array of KEY=VALUE strings into a map
func parsePairs(p []string) map[string]string {
	params := map[string]string{}
	for _, i := range p {
		parts := strings.Split(i, "=")
		if len(parts) != 2 {
			continue
		}
		params[parts[0]] = parts[1]
	}
	return params
}

// isAnyEnvSet checks if any of given environment vars is set and returns a boolean
func isAnyEnvSet(vars ...string) bool {
	for _, v := range vars {
		_, r := os.LookupEnv(v)
		if r {
			return r
		}
	}
	return false
}
