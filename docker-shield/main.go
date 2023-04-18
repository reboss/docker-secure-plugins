package main

import (
    "fmt"
    "log"
    "github.com/docker/go-plugins-helpers/authorization"
)

func main() {

    dockerShield, err := NewDockerShield()
    if err != nil {
        log.Fatal(err)
    }

    h := authorization.NewHandler(dockerShield)

    if err := h.ServeUnix("secopt", 0); err != nil {
        log.Fatal(err)
    }
}
