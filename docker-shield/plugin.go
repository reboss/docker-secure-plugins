package main

import (
    "bytes"
    "strings"
    "net/url"
    "regexp"
    "encoding/json"
    "github.com/docker/go-plugins-helpers/authorization"
    "github.com/docker/engine-api/types/container"
)

var createAPI = regexp.MustCompile(`/v.*/containers/create`)
var execAPI = regexp.MustCompile(`/v.*/containers/.+/exec`)

func NewDockerShield() (*DockerShield) {
    return &DockerShield{}
}

// Implements authorization.Plugin interface
type DockerShield struct {}

type configWrapper struct {
    HostConfig *container.HostConfig
}

type execWrapper struct {
    Privileged bool
}

func validateSecOpts(body *configWrapper) bool {
    for _, secopt := range body.HostConfig.SecurityOpt {
        apparmor := strings.Contains(secopt, "apparmor")
        seccomp := strings.Contains(secopt, "seccomp")
        if apparmor || seccomp {
            return false
        }
    }
    return true
}

func (p *DockerShield) AuthZReq(req authorization.Request) authorization.Response {

    uri, err := url.QueryUnescape(req.RequestURI)
    if err != nil {
        return authorization.Response{Err: err.Error()}
    }

    if req.RequestMethod == "POST" && req.RequestBody != nil {

        if createAPI.MatchString(uri) {
            body := &configWrapper{}

            if err := json.NewDecoder(bytes.NewReader(req.RequestBody)).Decode(body); err != nil {
                return authorization.Response{Err: err.Error()}
            }

            // Reject custom apparmor or seccomp
            if body.HostConfig.SecurityOpt != nil && !validateSecOpts(body) {
                return authorization.Response{Msg: "Not allowed to modify security options"}
            }

            // Reject privileged containers
            if body.HostConfig.Privileged {
                return authorization.Response{Msg: "Privileged containers not allowed"}
            }
        }

        else if execAPI.MatchString(uri) {
            body := &execWrapper{}

            if err := json.NewDecoder(bytes.NewReader(req.RequestBody)).Decode(body); err != nil {
                return authorization.Response{Err: err.Error()}
            }

            // Reject execing into containers in privleged mode
            if body.Privileged {
                return authorization.Response{Msg: "Privileged containers not allowed"}
            }

        }
    }
    return authorization.Response{Allow: true}
}

func (p *DockerShield) AuthZRes(req authorization.Request) authorization.Response {
        return authorization.Response{Allow: true}
}
