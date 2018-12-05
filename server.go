package main

import (
    "bytes"
    "fmt"
    "io"
    "io/ioutil"
    "net/http"
    "os"
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

// TODO refactor everything

// Extract as text document
func extractText(w http.ResponseWriter, r *http.Request, format string) {
    
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

// Extract as image collection
func extractImage(w http.ResponseWriter, r *http.Request, format string) {
    
    // Create temporary folder
    tmpFolder, _ := ioutil.TempDir("", "")
    defer os.RemoveAll(tmpFolder)
    
    // Get attachment
    attachmentFile, _, attachmentError := r.FormFile("file")
    if attachmentError != nil {
        w.WriteHeader(400)
        fmt.Fprintf(w, `No "file" attachment`)
        fmt.Printf("%s -> %s %s %s -> 400\n", r.RemoteAddr, r.Proto, r.Method, r.URL)
        return
    }
    defer attachmentFile.Close()
    
    // Store attachment
    inputPath := tmpFolder + "/input.pdf"
    inputFile, _ := os.Create(inputPath)
    defer inputFile.Close()
    io.Copy(inputFile, attachmentFile)
    inputFile.Sync()
    
    // Create output folder
    outputFolder := tmpFolder + "/output"
    os.Mkdir(outputFolder, 0700)
    
    // Prepare command
    outputPrefix := outputFolder + "/page"
    // TODO -jpegopt <string>        : jpeg options, with format <opt1>=<val1>[,<optN>=<valN>]*
    // TODO -r <fp>                  : resolution, in PPI (default is 150)
    // TODO -antialias <string>      : set cairo antialias option
    cmd := exec.Command("pdftocairo", "-jpeg", inputPath, outputPrefix)
    
    // Run command
    cmdOutput, cmdError := cmd.CombinedOutput()
    if cmdError != nil {
        w.WriteHeader(400)
        w.Write(cmdOutput)
        fmt.Println(cmdError)
        fmt.Printf("%s -> %s %s %s -> 400\n", r.RemoteAddr, r.Proto, r.Method, r.URL)
        return
    }
    
    // Collect results into a single archive
    tarPath := tmpFolder + "/output.tar.gz"
    tarCmd := exec.Command("tar", "-C", tmpFolder, "-zcvf", tarPath, "output")
    tarCmd.Run()
    
    // Send result
    tarFile, _ := os.Open(tarPath)
    defer tarFile.Close()
    w.Header().Set("Content-Type", "application/gzip")
    io.Copy(w, tarFile)
    fmt.Printf("%s -> %s %s %s -> 200\n", r.RemoteAddr, r.Proto, r.Method, r.URL)

}

// Extract PDF content
func extractHandler(w http.ResponseWriter, r *http.Request) {
    
    // TODO Top-level try catch
    
    // Select appropriate handler
    format := r.URL.Query().Get("format")
    switch format {
    case "txt":
        extractText(w, r, "txt")
    case "xml", "":
        extractText(w, r, "xml")
    case "jpg":
        extractImage(w, r, "jpg")
    case "png":
        extractImage(w, r, "png")
    default:
        w.WriteHeader(400)
        fmt.Fprintf(w, `Invalid format "%s"`, format)
        fmt.Printf("%s -> %s %s %s -> 400\n", r.RemoteAddr, r.Proto, r.Method, r.URL)
    }
    
}

// Entry point
func main() {
    fmt.Printf("Using Poppler %d.%d.%d\n", major, minor, revision)
    fmt.Printf("Listening on port 8080\n")
    http.HandleFunc("/version", versionHandler)
    http.HandleFunc("/extract", extractHandler)
    http.ListenAndServe(":8080", nil)
}
