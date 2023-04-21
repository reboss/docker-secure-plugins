package main

import (
    "fmt"
    "bytes"
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
    var resBytes []byte

    reqBytes, err = json.Marshal(body)
    if err != nil {
        goto internalServerError
    }

    newReq, err = getForwardReq(body, req, reqBytes)
    if err != nil {
        goto internalServerError
    }

    res, err = http.DefaultClient.Do(newReq)
    if err != nil {
        http.Error(w, err.Error(), http.StatusBadGateway)
        return
    }
    defer res.Body.Close()

    resBytes, err = getForwardRes(res, w)
    if err != nil {
        goto internalServerError
    }

    _, err = w.Write(resBytes)
    if err != nil {
        goto internalServerError
    }

internalServerError:
    http.Error(w, err.Error(), http.StatusInternalServerError)
    return
}

func main() {
    http.HandleFunc("/", proxyHandler)
    fmt.Println("Listening on port 8080...")
    log.Fatal(http.ListenAndServe(":8080", nil))
}
