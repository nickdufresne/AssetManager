package main

import (
  "github.com/gorilla/mux"
  "github.com/gorilla/sessions"
  "net/http"
  "html"
  "html/template"
  "fmt"
  //"io/ioutil"
  "path"
  "path/filepath"
  "os"
  "io"
  "time"
  "strings"
  "log"
)

type Context struct {
  w http.ResponseWriter
  r *http.Request
}

func (ctx Context) Write(b []byte) (int, error) {
  return ctx.w.Write(b)
}

func (ctx Context) WriteHeader(status int) {
  ctx.w.WriteHeader(status)
}


type handler func(ctx Context) error

func (h handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
  start := time.Now()

  ctx := Context{w,r}

  err := h(ctx)

  if err != nil {
    fmt.Fprintln(w, "Error processing request: ", err)
  }

  elapsed := time.Since(start)

  ct := w.Header().Get("content-type")
  fmt.Printf("%s [%s] content-type: %s\n", r.URL.Path, elapsed, ct)

}


func form(ctx Context) (err error) {
  ctx.Write([]byte("Hello World!"))
  return
}

func formTest(ctx Context) (err error) {
  ctx.Write([]byte("Hello World!"))
  return
}

func formTest2(ctx Context) (err error) {
  ctx.Write([]byte("Hello World!"))
  return
}

func uploadFile(ctx Context) (err error) {
  err = ctx.r.ParseMultipartForm(0)

  if err != nil {
    return
  }

  for name, parts := range ctx.r.MultipartForm.File {
    fmt.Println(name)
    for index, part := range parts {
      fmt.Println("Part: ", index, " => ", part.Filename)
    }
  }

  fmt.Fprintf(ctx.w, "Upload, %q", html.EscapeString(ctx.r.URL.Path))

  return
}

func notFound(ctx Context) error {
  ctx.WriteHeader(http.StatusNotFound)
  _, err := ctx.Write([]byte("Not Found!"))
  return err
}



func index(ctx Context) (err error) {
  session, _ := store.Get(ctx.r, "session-name")
  // Set some session values.
  session.Values["foo"] = "bar"
  session.Values[42] = session.Values[42].(int) + 10
  // Save it.
  session.Save(ctx.r, ctx.w)

  data := struct{
    Value int
  }{ session.Values[42].(int) }

  return ctx.renderTemplate("index", data)
}

func getStaticContentType(path string) string {
  ending := strings.ToLower(filepath.Ext(path))

  //log.Println("ext: ", ending)
  switch ending {
    case ".js":
      return "text/javascript"
    case ".css":
      return "text/css"
    case ".jpg", ".jpeg":
      return "image/jpeg"
    case ".png":
      return "image/png"
    default:
      return "application/octet-stream"
  }
}

func serveStatic(ctx Context) (err error) {
  path := path.Join("./assets", ctx.r.URL.Path)
  
  if _, err := os.Stat(path); os.IsNotExist(err) {
    return notFound(ctx)
  }

  ct := getStaticContentType(path)

  r, err := os.Open(path)

  if err != nil {
    return
  }

  ctx.w.Header().Set("content-type", ct)

  io.Copy(ctx.w,r)

  return

}

func getTemplate(name string) (tmpl *template.Template, found bool, err error) {
  if env == "dev" {
    var filename string
    filename, found = templateMap[name]

    if !found {
      return
    }

    path := path.Join("./templates", filename)
    log.Printf("Parsing template: %s", filename)
    tmpl, err = template.ParseFiles(path)
    return

  }

  // all other environments
  tmpl, found = templates[name]

  return   
  
}

func (ctx Context) renderTemplate(name string, v interface{}) error {
  tmpl, ok, err := getTemplate(name)
  
  if !ok {
    return notFound(ctx)
  }

  if err != nil {
    return err
  }

  ctx.w.Header().Set("content-type", "text/html")
  err = tmpl.Execute(ctx.w, v)
  return err
}

func loadTemplate(name string, file string) {
  path := path.Join("./templates", file)
  tmpl, err := template.ParseFiles(path)
  if err != nil {
    fmt.Printf("Error loading template %s: %s\n", path, err)
    panic(err)
  }

  templates[name] = tmpl
}

func init() {

  log.SetPrefix("[DEV] ")

  templateMap = map[string]string{
    "upload": "upload.html",
    "index": "index.html",
  }

  if env != "dev" {
    for name, file := range templateMap {
      loadTemplate(name, file)
    }
  }
}

var env = "dev"

var templateMap map[string]string
var templates = make(map[string]*template.Template)
var store = sessions.NewCookieStore([]byte("something-very-secret"))

func random(ctx Context) error {
  ctx.w.Write([]byte("Hello World"))
  return nil
}

func main() {
  r := mux.NewRouter()
  r.Handle("/random", handler(random))
  http.Handle("/js/", handler(serveStatic))
  http.Handle("/css/", handler(serveStatic))
  http.Handle("/images/", handler(serveStatic))
  http.Handle("/upload", handler(uploadFile))
  http.Handle("/random", r)
  http.Handle("/", handler(index))
  log.Println("AssetManager started on port 8888")
  err := http.ListenAndServe(":8888", nil)
  if err != nil {
    log.Fatal("ListenAndServe: ", err)
  }
}