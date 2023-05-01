package main

import (
    "fmt"
    "bytes"
    "bufio"
    "io"
    "net"
    "encoding/json"
    "io/ioutil"
    "log"
    "regexp"
    "net/http"
)

var createAPI = regexp.MustCompile(`/v.*/containers/create`)

func isNoNewPriv(body requestBody) bool {
    securityOpts := body.SecurityOpt
    for _, secopt := range securityOpts {
        if secopt == "no-new-privileges" {
            return true
        }
    }
    return false
}

func getForwardReq(body requestBody, req *http.Request, reqBytes []byte) (*http.Request, error) {
    reqBytes, err := json.Marshal(body)
    if err != nil {
        return nil, err
    }

    newReq, err := http.NewRequest(req.Method, req.URL.String(), ioutil.NopCloser(
        bytes.NewReader(reqBytes),
    ))
    if err != nil {
        return nil, err
    }
    newReq.Header = req.Header.Clone()

    return newReq, nil
}

// TODO:  Get type for res and resBytes
func getForwardRes(res *http.Response, w http.ResponseWriter) ([]byte, error) {
    for key, values := range res.Header {
        for _, value := range values {
            w.Header().Add(key, value)
        }
    }

    w.WriteHeader(res.StatusCode)

    resBytes, err := ioutil.ReadAll(res.Body)
    if err != nil {
        return nil, err
    }

    return resBytes, nil
}

type requestBody struct {
    SecurityOpt []string `json:"SecurityOpt,omitempty"`
}

func proxyHandler(w http.ResponseWriter, req *http.Request) {
    body := requestBody{}

    err := json.NewDecoder(req.Body).Decode(&body)
    if err != nil {
        // I will want to respond as if this is a docker error
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }

    if body.SecurityOpt != nil && !isNoNewPriv(body) {
        body.SecurityOpt = append(body.SecurityOpt, "no-new-privileges")
    }

    var reqBytes []byte
    var newReq *http.Request
    var res *http.Response
    var conn net.Conn

    reqBytes, err = json.Marshal(body)
    if err != nil {
        goto internalServerError
    }

    newReq, err = getForwardReq(body, req, reqBytes)
    if err != nil {
        goto internalServerError
    }

    // Dial the Unix domain socket
    conn, err = net.Dial("unix", "/var/run/docker.sock")
    if err != nil {
        goto internalServerError
    }
    defer conn.Close()

    // Forward the incoming request to the Unix domain socket
    err = newReq.Write(conn)
    if err != nil {
        goto internalServerError
    }

    // Read the response from the Unix domain socket
    res, err = http.ReadResponse(bufio.NewReader(conn), req)
    if err != nil {
        goto internalServerError
    }
    defer res.Body.Close()

    fmt.Println(req)
    fmt.Println(res)

    // Copy the response headers to the proxy response
    for key, values := range res.Header {
        for _, value := range values {
            w.Header().Add(key, value)
        }
    }

    // Copy the response status code to the proxy response
    w.WriteHeader(res.StatusCode)

    // Copy the response body to the proxy response
    _, err = io.Copy(w, res.Body)
    if err != nil {
        goto internalServerError
    }
    return

internalServerError:
    http.Error(w, err.Error(), http.StatusInternalServerError)
    return
}

func main() {
    http.HandleFunc("/", proxyHandler)
    //socket := "/var/run/docker.sock"
    socket := "/tmp/proxy.sock"
    listener, err := net.Listen("unix", socket)
    if err != nil {
        log.Fatal("Error listening on socket %s: %v", socket, err)
        return
    }
    defer listener.Close()

    fmt.Println("Listening on %s", socket)

    log.Fatal(http.Serve(listener, nil))
}
