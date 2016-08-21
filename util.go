package main

import (
	"fmt"
	"os"
	"strings"

	drone "github.com/drone/drone/client"
	"github.com/drone/drone/model"
	"github.com/urfave/cli"
)

// findBuild requests an array of previous repo builds from drone API and finds
// the one that matches given filters
func findBuild(c drone.Client, ctx *cli.Context) (*model.Build, error) {
	user, repo, err := parseRepo(ctx.String("repo"))
	if err != nil {
		return nil, err
	}
	builds, err := c.BuildList(user, repo)
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

// match returns first build that matches cli filter flags
func match(ctx *cli.Context, build *model.Build) bool {
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
	if ctx.String("branch") == build.Branch {
		return true
	}
	// Return latest successful build if no specific filters are set
	return true
}

func newDroneClient(ctx *cli.Context) drone.Client {
	return drone.NewClientToken(ctx.String("drone-server"), ctx.String("drone-token"))
}

// parseRepo parses a owner/repo string into two parts
func parseRepo(str string) (owner, repo string, err error) {
	var parts = strings.Split(str, "/")
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

// isAnyEnvSet sets if any of given environment vars is set and returns a boolean
func isAnyEnvSet(vars ...string) bool {
	var r bool
	for _, v := range vars {
		_, r := os.LookupEnv(v)
		if r {
			return r
		}
	}
	return r
}
