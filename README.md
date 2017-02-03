# drone-trigger

[![Build Status](https://travis-ci.org/UKHomeOffice/drone-trigger.svg?branch=master)](https://travis-ci.org/UKHomeOffice/drone-trigger) [![Docker Repository on Quay](https://quay.io/repository/ukhomeofficedigital/drone-trigger/status "Docker Repository on Quay")](https://quay.io/repository/ukhomeofficedigital/drone-trigger)

Drone plugin for triggering downstream builds with custom parameters

This plugin allows for triggering remote or local builds. You can specify
various filters, like `tag`, `branch`, `number`, `commit`, etc. It will find an
existing build which matches specified filters and trigger the build restart or
deployment, in addition, custom parameters can also be set.

## Build

Dependencies are located in the vendor directory and managed using
[govendor](https://github.com/kardianos/govendor) cli tool.

```
go test -v -cover

mkdir -p bin
GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -ldflags "-X main.Version=dev+git" -o bin/drone-trigger_linux_amd64
```

## Configuration

The following parameters are used to configure the plugin:

- `drone_server`: full URL to the drone server, it can be a remote drone server as well
- `drone_token` or `${DRONE_TOKEN}`: drone user secret token. Just create a `DRONE_TOKEN` secret, the plugin will pick it up
- `repo`: git repository in owner/name format
- `status`: build status filter, default is `success`
- `event`: build event type filter. If unset, no event filter will be done
- `deploy_to`: sends a deployment trigger, which also sets a `DRONE_DEPLOY_TO` environment variable in the target job
- `params`: list of custom parameters that will be passed into a build environment as environment variables
- `fork`: create a new build and a build number instead of restarting an existing build. Please note that a deployment trigger always spawns a new build
- `verbose`: displays a more verbose output

Only one filter from the below list can be specified.
- `number`: filter by specific build number
- `commit`: filter by long commit sha
- `branch`: filter by branch name
- `tag`: filter by tag name. Please note that event type will be `tag`.
- `deployed-to`: filter by the environment deployed to. Please note that event type will be `deployment`.


### Drone configuration

```yaml
pipeline:
  build:
    image: golang
    commands:
      - go get -v
      - ...

  docker_build:
    image: docker:1.11
    commands:
      - docker login ...
      - docker build -t foo/bar:${DRONE_COMMIT_SHA} .
      - docker push foo/bar:${DRONE_COMMIT_SHA}

  trigger_deploy:
    image: quay.io/ukhomeofficedigital/drone-trigger:latest
    drone_server: https://drone.example.com
    repo: owner/go-deploy-scripts
    branch: master
    deploy_to: prod
    params: "IMAGE_NAME=foo/bar:${DRONE_COMMIT_SHA},APP_ID=123"
```


Since drone-trigger can be run as a standalone tool, configuration can be
provided via cli flags and arguments as well as environment variables.

```bash
drone-trigger --help
NAME:
   drone-trigger - trigger drone builds or deployments

USAGE:
   drone-trigger_linux_amd64 [global options] command [command options] [arguments...]

COMMANDS:
     help, h  Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --verbose                      verbose output [$VERBOSE, $PLUGIN_VERBOSE]
   --fork                         fork an existing build - drone assigns a new build number [$FORK, $PLUGIN_FORK]
   --drone-server URL, -s URL     drone server URL [$DRONE_SERVER, $PLUGIN_DRONE_SERVER]
   --drone-token TOKEN, -t TOKEN  drone auth TOKEN [$DRONE_TOKEN, $PLUGIN_DRONE_TOKEN]
   --repo REPO, -r REPO           REPO, eg. foo/bar [$REPO, $PLUGIN_REPO]
   --commit value, -c value       filter by commit sha [$FILTER_COMMIT, $PLUGIN_COMMIT]
   --tag value                    filter by tag [$FILTER_TAG, $PLUGIN_TAG]
   --branch value, -b value       filter by branch [$FILTER_BRANCH, $PLUGIN_BRANCH]
   --status value                 filter by build status (default: "success") [$FILTER_STATUS, $PLUGIN_STATUS]
   --number value                 filter by build number (default: 0) [$FILTER_NUMBER, $PLUGIN_NUMBER]
   --event value                  filter by trigger event [$FILTER_EVENT, $PLUGIN_EVENT]
   --deploy-to value, -d value    environment to deploy to, if set a deployment event will be triggered [$DEPLOY_TO, $PLUGIN_DEPLOY_TO]
   --param value, -p value        custom parameters to include in the trigger in KEY=value format [$PARAMS, $PLUGIN_PARAMS]
   --help, -h                     show help
   --version, -v                  print the version

```

## Release process

Push / Merge to master will produce a docker
[image](https://quay.io/repository/ukhomeofficedigital/drone-trigger?tab=tags) with a tag `latest`.

To create a new release, just create a new tag off master.

## Contributing

We welcome pull requests. Please check issues and existing PRs before submitting a patch.

## Author

Vaidas Jablonskis [vaijab](https://github.com/vaijab)

