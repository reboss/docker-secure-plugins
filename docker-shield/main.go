package main

import (
    "log"
    "github.com/docker/go-plugins-helpers/authorization"
)

func main() {

    dockerShield := NewDockerShield()
    h := authorization.NewHandler(dockerShield)

    if err := h.ServeUnix("secopt", 0); err != nil {
        log.Fatal(err)
    }
}
