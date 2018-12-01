package main

import (
    "bytes"
    "fmt"
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
    fmt.Printf("%s -> %s %s %s -> 200\n", r.RemoteAddr, r.Proto, r.Method, r.URL)
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
    var code int
    if err != nil {
        code = 400
        w.WriteHeader(code)
        w.Write(stderr.Bytes())
    } else {
        w.Write(stdout.Bytes())
    }
    fmt.Printf("%s -> %s %s %s -> %d\n", r.RemoteAddr, r.Proto, r.Method, r.URL, code)
    
}

// Entry point
func main() {
    fmt.Printf("Using Poppler %d.%d.%d\n", major, minor, revision)
    fmt.Printf("Listening on port 8080\n")
    http.HandleFunc("/version", versionHandler)
    http.HandleFunc("/extract", extractHandler)
    http.ListenAndServe(":8080", nil)
}
