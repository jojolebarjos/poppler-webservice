package main

import (
    "fmt"
    "log"
    "net/http"
    "os/exec"
    "regexp"
    "strconv"
)

func acquireVersion() (major, minor, revision int) {
    cmd := exec.Command("pdftotext", "-v")
    out, err := cmd.CombinedOutput()
    if err != nil {
        panic(err)
    }
    text := string(out)
    regex := regexp.MustCompile(`(\d+)\.(\d+)\.(\d+)`)
    match := regex.FindStringSubmatch(text)
    major, _ = strconv.Atoi(match[1])
    minor, _ = strconv.Atoi(match[2])
    revision, _ = strconv.Atoi(match[3])
    return
}

var major, minor, revision = acquireVersion()

func versionHandler(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json")
    fmt.Fprintf(w, `{"major":%d,"minor":%d,"revision":%d}`, major, minor, revision)
}

func main() {
    http.HandleFunc("/version", versionHandler)
    // TODO handle pdftotext and other commands
    log.Fatal(http.ListenAndServe(":8080", nil))
}
