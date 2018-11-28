package main

import (
    "bytes"
    "fmt"
    "log"
    "net/http"
    "os/exec"
    "regexp"
    "strconv"
)

// Load version at startup
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

// Provide information about webservice
func versionHandler(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json")
    fmt.Fprintf(w, `{"major":%d,"minor":%d,"revision":%d}`, major, minor, revision)
}

// Extract PDF content
func extractHandler(w http.ResponseWriter, r *http.Request) {
    
    // Prepare command
    // TODO use query string to select appropriate output (plain text, html, xml...)
    cmd := exec.Command("pdftotext", "-enc", "UTF-8", "-eol", "unix", "-layout", "-", "-")
    
    // Pipe attachment content to input
    file, _, _ := r.FormFile("file")
    defer file.Close()
    cmd.Stdin = file
    
    // Pipe output to buffers
    var stdout, stderr bytes.Buffer
    cmd.Stdout = &stdout
    cmd.Stderr = &stderr
    
    // Run
    err := cmd.Run()
    if err != nil {
        w.WriteHeader(400)
        w.Write(stderr.Bytes())
    } else {
        w.Write(stdout.Bytes())
    }
    // TODO log query at some point
    
}

// Entry point
func main() {
    http.HandleFunc("/version", versionHandler)
    http.HandleFunc("/extract", extractHandler)
    log.Fatal(http.ListenAndServe(":8080", nil))
}
