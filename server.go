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
    
    // Get query parameters
    format := r.URL.Query().Get("format")
    if format == "" {
        format = "xml"
    }
    
    // Prepare command
    var cmd *exec.Cmd
    switch format {
    // TODO HTML
    case "txt":
        cmd = exec.Command("pdftotext", "-enc", "UTF-8", "-eol", "unix", "-layout", "-", "-")
    case "xml":
        cmd = exec.Command("pdftotext", "-enc", "UTF-8", "-eol", "unix", "-bbox-layout", "-", "-")
        w.Header().Set("Content-Type", "application/xhtml+xml")
    default:
        w.WriteHeader(400)
        fmt.Fprintf(w, `Invalid format "%s"`, format)
        return
    }
    
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
        fmt.Printf("%s -> %s %s %s -> 400\n", r.RemoteAddr, r.Proto, r.Method, r.URL)
        return
    }
    
    // Send payload
    w.Write(stdout.Bytes())
    fmt.Printf("%s -> %s %s %s -> 200\n", r.RemoteAddr, r.Proto, r.Method, r.URL)
    
}

// Entry point
func main() {
    fmt.Printf("Using Poppler %d.%d.%d\n", major, minor, revision)
    fmt.Printf("Listening on port 8080\n")
    http.HandleFunc("/version", versionHandler)
    http.HandleFunc("/extract", extractHandler)
    http.ListenAndServe(":8080", nil)
}
