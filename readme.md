
# Poppler webservice container

[Poppler](https://poppler.freedesktop.org/) is a PDF rendering library, which can be used to extract content into machine-readable format.

This container provide a small webservice (~22.6Mo) with basic conversion capabilities.


## Usage

```
docker build -t jojolebarjos/poppler .
docker run -d -p 8080:8080 jojolebarjos/poppler
```

```
HTTP GET /version

{
    "major": 0,
    "minor": 71,
    "revision": 0
}
```

```
HTTP POST /extract?format=xml
file: <my.pdf>

<!DOCTYPE html PUBLIC "-//W3C//DTD XHTML 1.0 Transitional//EN" "http://www.w3.org/TR/xhtml1/DTD/xhtml1-transitional.dtd">
<html xmlns="http://www.w3.org/1999/xhtml">
    <head>
...
```


## ToDo list

 * Support HTTPS
 * Limit attachment size? (should take parameter from env var)
 * Output more formats (e.g. HTML, maybe with images, PS)
 * Maybe use UPX to reduce size
 * Use [C](https://blog.golang.org/c-go-cgo) directly?
 * Better error handling
