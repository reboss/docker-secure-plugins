package main

import (
    "github.com/docker/go-plugins-helpers/authorization"
)

func NewDockerShield() (*DockerShield) {
    return &DockerShield{}
}

type DockerShield struct {}

func (p *DockerShield) AuthZReq(req authorization.Request) authorization.Response {
    return authorization.Response{Allow: true}
}

func (p *authobot) AuthZRes(req authorization.Request) authorization.Response {
        return authorization.Response{Allow: true}
}
