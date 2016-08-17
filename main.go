package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path"
	"strconv"

	"github.com/drone/drone/model"
	"github.com/urfave/cli"
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
		cli.BoolFlag{
			Name:   "fork",
			Usage:  "fork an existing build - drone assigns a new build number",
			EnvVar: "FORK,PLUGIN_FORK",
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

func run(ctx *cli.Context) error {
	// Exit if required flags are not set
	if !ctx.IsSet("drone-server") {
		cli.ShowAppHelp(ctx)
		return cli.NewExitError("error: drone server is not set", 3)
	}
	if !ctx.IsSet("drone-token") {
		cli.ShowAppHelp(ctx)
		return cli.NewExitError("error: drone token is not set", 3)
	}
	if !ctx.IsSet("repo") {
		cli.ShowAppHelp(ctx)
		return cli.NewExitError("error: repo is not set", 3)
	}
	if (ctx.IsSet("tag") && (ctx.IsSet("branch") || ctx.IsSet("commit"))) || (ctx.IsSet("branch") && ctx.IsSet("commit")) {
		return cli.NewExitError("error: tag, branch or commit cannot be set at the same time, pick one filter", 3)
	}

	c := newDroneClient(ctx)
	build, err := findBuild(c, ctx)
	if err != nil {
		return cli.NewExitError(err.Error(), 1)
	}
	if build == nil {
		return cli.NewExitError("No previous builds found", 1)
	}

	params := parsePairs(ctx.StringSlice("param"))
	newBuild := &model.Build{}
	owner, repo, err := parseRepo(ctx.String("repo"))
	if ctx.IsSet("deploy-to") {
		b, err := c.Deploy(owner, repo, build.Number, ctx.String("deploy-to"), params)
		if err != nil {
			return cli.NewExitError(err.Error(), 1)
		}
		newBuild = b
	} else {
		if ctx.IsSet("fork") {
			params["fork"] = "true"
		}
		b, err := c.BuildStart(owner, repo, build.Number, params)
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
