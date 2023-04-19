package main

import (
    "bytes"
    "fmt"
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

// TODO: We can just use the interface provided by authorization
type DockerShield struct {}

type configWrapper struct {
    HostConfig *container.HostConfig
}

func (p *DockerShield) AuthZReq(req authorization.Request) authorization.Response {

    uri, err := url.QueryUnescape(req.RequestURI)
    if err != nil {
        return authorization.Response{Err: err.Error()}
    }

    // TODO: Need to prevent --privileged option when running `docker exec`
    if req.RequestMethod == "POST" && req.RequestBody != nil && createAPI.MatchString(uri) {
        body := &configWrapper{}
        if err := json.NewDecoder(bytes.NewReader(req.RequestBody)).Decode(body); err != nil {
            return authorization.Response{Err: err.Error()}
        }

        // Reject privileged containers
        if body.HostConfig.Privileged {
            return authorization.Response{Msg: "Privileged containers not allowed"}
        }

        // Reject custom apparmor or seccomp
        if body.HostConfig.SecurityOpt != nil {
            for _, secopt := range body.HostConfig.SecurityOpt {
                apparmor := strings.Contains(secopt, "apparmor")
                seccomp := strings.Contains(secopt, "seccomp")
                if apparmor || seccomp {
                    return authorization.Response{Msg: "Not allowed to modify security options"}
               }
            }
        }
    }
    return authorization.Response{Allow: true}
}

func (p *DockerShield) AuthZRes(req authorization.Request) authorization.Response {
        return authorization.Response{Allow: true}
}
