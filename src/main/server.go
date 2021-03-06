// 
// Copyright (c) 2014 Brian William Wolter, All rights reserved.
// Slang
// 
// Redistribution and use in source and binary forms, with or without modification,
// are permitted provided that the following conditions are met:
// 
//   * Redistributions of source code must retain the above copyright notice, this
//     list of conditions and the following disclaimer.
// 
//   * Redistributions in binary form must reproduce the above copyright notice,
//     this list of conditions and the following disclaimer in the documentation
//     and/or other materials provided with the distribution.
//     
//   * Neither the names of Brian William Wolter, Wolter Group New York, nor the
//     names of its contributors may be used to endorse or promote products derived
//     from this software without specific prior written permission.
//     
// THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS "AS IS" AND
// ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE IMPLIED
// WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE ARE DISCLAIMED.
// IN NO EVENT SHALL THE COPYRIGHT HOLDER OR CONTRIBUTORS BE LIABLE FOR ANY DIRECT,
// INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR CONSEQUENTIAL DAMAGES (INCLUDING,
// BUT NOT LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR SERVICES; LOSS OF USE,
// DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER CAUSED AND ON ANY THEORY OF
// LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY, OR TORT (INCLUDING NEGLIGENCE
// OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE OF THIS SOFTWARE, EVEN IF ADVISED
// OF THE POSSIBILITY OF SUCH DAMAGE.
// 

package main

import (
  "os"
  "fmt"
  "log"
  "path"
  "time"
  "strings"
)

import (
  "ejs"
  "bww/errors"
)

import (
  "net/url"
  "net/http"
  "html"
  "html/template"
)

/**
 * Resource mimetypes
 */
var MIMETYPES = map[string]string {
  ".scss":  "text/css",
  ".css":   "text/css",
  ".ejs":   "application/javascript",
  ".js":    "application/javascript",
  ".ghtml": "text/html",
  ".html":  "text/html",
}

/**
 * Managed resource extensions
 */
var MANAGED_EXTENSIONS = map[string]bool {
  ".css":   true,
  ".scss":  true,
  ".ejs":   true,
  ".js":    true,
  ".ghtml": true,
}

/**
 * A server
 */
type Server struct {
  port    int
  peer    *url.URL
  root    string
  routes  map[string][]string
  proxy   *ReverseProxy
  strict  bool
}

/**
 * Create a server
 */
func NewServer(port int, peer, root string, routes map[string][]string) (*Server, error) {
  var proxy *ReverseProxy = nil
  var peerURL *url.URL = nil
  
  if peer != "" {
    var err error
    if peerURL, err = url.Parse(peer); err != nil {
      return nil, err
    }else{
      proxy = NewSingleHostReverseProxy(peerURL)
    }
  }
  
  return &Server{port, peerURL, root, routes, proxy, false}, nil
}

/**
 * Run the server
 */
func (s *Server) Run() error {
  
  mux := http.NewServeMux()
  
  if s.proxy != nil {
    mux.HandleFunc("/", s.handler)
  }else{
    mux.HandleFunc("/", s.serveRequest)
  }
  
  server := &http.Server{
    Addr: fmt.Sprintf(":%d", s.port),
    Handler: mux,
    ReadTimeout: 30 * time.Second,
    WriteTimeout: 30 * time.Second,
  }
  
  return server.ListenAndServe()
}

/**
 * Handle a request
 */
func (s *Server) handler(writer http.ResponseWriter, request *http.Request) {
  if request.Method == "GET" && CanCompile(nil, request.URL.Path) {
    s.serveRequest(writer, request)
  }else{
    s.proxyRequest(writer, request)
  }
}

/**
 * Proxy a request
 */
func (s *Server) proxyRequest(writer http.ResponseWriter, request *http.Request) {
  
  if s.proxy != nil && SharedOptions().GetFlag(OptionsFlagVerbose) {
    if u, err := url.Parse(request.URL.Path); err == nil {
      log.Printf("%s %s \u2192 %v", request.Method, request.URL.Path, s.peer.ResolveReference(u))
    }
  }
  
  if s.proxy == nil {
    s.serveError(writer, request, http.StatusBadGateway, fmt.Errorf("No proxy is configured for non-managed resource: %s", request.URL.Path))
  }else if err := s.proxy.ServeHTTP(writer, request); err == FileNotFoundError {
    s.serveRequestWithOptions(writer, request, false) // attempt to serve the local version
  }else if err != nil {
    s.serveError(writer, request, http.StatusBadGateway, err)
  }
  
}

/**
 * Route a request
 */
func (s *Server) routeRequest(request *http.Request) ([]string, string, error) {
  var candidates []string
  var mimetype string
  
  absolute := request.URL.Path
  alternates := make([]string, 0)
  
  for k, v := range s.routes {
    if strings.HasPrefix(absolute, k) {
      for _, e := range v {
        alternates = append(alternates, path.Join(e, absolute[len(k):]))
      }
    }
  }
  
  ext := path.Ext(absolute)
  relatives := make([]string, len(alternates))
  bases := make([]string, len(alternates))
  
  for i, e := range alternates {
    r := path.Join(s.root, e[1:])
    relatives[i] = r
    bases[i] = r[:len(r) - len(ext)]
  }
  
  bases = append(bases, path.Join(s.root, absolute[:len(absolute) - len(ext)]))
  
  extWithBases := func(b []string, ext string) []string {
    o := make([]string, len(b))
    for i, e := range b {
      o[i] = e + ext
    }
    return o
  }
  
  switch ext {
    case ".css":
      candidates = append(extWithBases(bases, ".scss"), relatives...)
    case ".js":
      candidates = append(extWithBases(bases, ".ejs"), relatives...)
    case ".html":
      candidates = append(extWithBases(bases, ".ghtml"), relatives...)
    default:
      candidates = relatives
  }
  
  var ok bool
  if mimetype, ok = MIMETYPES[ext]; !ok {
    mimetype = "text/plain"
  }
  
  return candidates, mimetype, nil
}

/**
 * Serve a request
 */
func (s *Server) serveRequest(writer http.ResponseWriter, request *http.Request) {
  s.serveRequestWithOptions(writer, request, true)
}

/**
 * Serve a request
 */
func (s *Server) serveRequestWithOptions(writer http.ResponseWriter, request *http.Request, allowProxy bool) {
  var candidates []string
  var mimetype string
  var file *os.File
  var err error
  
  if candidates, mimetype, err = s.routeRequest(request); err != nil {
    s.serveError(writer, request, http.StatusNotFound, fmt.Errorf("Could not map resource: %s", request.URL.Path))
    return
  }
  
  if SharedOptions().GetFlag(OptionsFlagVerbose) {
    log.Printf("%s %s \u2192 {%s}", request.Method, request.URL.Path, strings.Join(candidates, ", "))
  }
  
  for _, e := range candidates {
    if file, err = os.Open(e); err == nil {
      defer file.Close()
      if !SharedOptions().GetFlag(OptionsFlagQuiet) { log.Printf("%s %s \u2192 %s", request.Method, request.URL.Path, e) }
      writer.Header().Add("Content-Type", mimetype)
      s.compileAndServeFile(writer, request, file)
      return
    }
  }
  
  if !allowProxy || s.strict || s.proxy == nil {
    s.serveError(writer, request, http.StatusNotFound, fmt.Errorf("No such resource: %s", request.URL.Path))
  }else{
    s.proxyRequest(writer, request)
  }
  
}

/**
 * Serve a request
 */
func (s *Server) compileAndServeFile(writer http.ResponseWriter, request *http.Request, file *os.File) {
  context := NewContext()
  
  if fstat, err := file.Stat(); err != nil {
    s.serveError(writer, request, http.StatusBadRequest, fmt.Errorf("Could not stat file: %v", file.Name()))
    return
  }else if fstat.Mode().IsDir() {
    s.serveError(writer, request, http.StatusBadRequest, fmt.Errorf("Resource is not a file: %v", file.Name()))
    return
  }
  
  if compiler, err := NewCompiler(context, file.Name()); err != nil {
    s.serveError(writer, request, http.StatusBadRequest, fmt.Errorf("Resource is not supported: %v", file.Name()))
    return
  }else if err := compiler.Compile(context, file.Name(), "", file, writer); err != nil {
    s.serveError(writer, request, http.StatusInternalServerError, err)
    return
  }
  
}

/**
 * An error
 */
type templateError struct {
  Message   string
  Source    []template.HTML
  Base      int
}

/**
 * Serve an error
 */
func (s *Server) serveError(writer http.ResponseWriter, request *http.Request, status int, problem error) {
  log.Println("ERROR:", problem)
  if t, err := template.ParseFiles(SharedOptions().Resource("html/error.html")); err != nil {
    
    log.Printf("ERROR: Could not compile template: %v\n", err)
    writer.WriteHeader(status)
    writer.Write([]byte(problem.Error()))
    
  }else{
    var issues []*templateError
    var message string
    
    switch e := problem.(type) {
      case *errors.Error:
        message = e.Message()
        problem = e.Cause()
      case *ejs.SourceError:
        message = "Compilation error"
        // same error gets processed below
      default:
        message = e.Error()
        problem = nil
    }
    
    for problem != nil {
      switch e := problem.(type) {
        case *errors.Error:
          issues = append(issues, &templateError{e.Message(), nil, 0})
          problem = e.Cause()
        case *ejs.SourceError:
          lines := e.ExcerptLines("<span class=\"marker\">", "</span>", html.EscapeString, 3)
          excerpt := make([]template.HTML, len(lines))
          for i, l := range lines { excerpt[i] = template.HTML(l) }
          issues = append(issues, &templateError{fmt.Sprintf("%s\n%s", e.Location(), e.Message()), excerpt, e.Line()})
          problem = nil
        default:
          issues = append(issues, &templateError{e.Error(), nil, 0})
          problem = nil
      }
    }
    
    params := map[string]interface{} {
      "Title":    "Error",
      "Header":   fmt.Sprintf("%d: %s", status, http.StatusText(status)),
      "Message":  message,
      "Errors":   issues,
    }
    
    writer.Header().Add("Content-Type", "text/html")
    writer.WriteHeader(status)
    t.Execute(writer, params)
    
  }
}

