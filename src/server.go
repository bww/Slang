// 
// Copyright (c) 2014 Brian William Wolter, All rights reserved.
// Webasm
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
  "strings"
)

import (
  "net/url"
  "net/http"
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
  ".html":  "text/html",
}

/**
 * A server
 */
type Server struct {
  context Context
  port    int
  peer    string
  routes  map[string]string
  proxy   *ReverseProxy
}

/**
 * Create a server
 */
func NewServer(context Context, port int, peer string, routes map[string]string) (*Server, error) {
  var proxy *ReverseProxy = nil
  
  if peer != "" {
    if u, err := url.Parse(peer); err != nil {
      return nil, err
    }else{
      proxy = NewSingleHostReverseProxy(u)
    }
  }
  
  return &Server{context, port, peer, routes, proxy}, nil
}

/**
 * Run the server
 */
func (s *Server) Run() {
  http.HandleFunc("/", s.handler)
  http.ListenAndServe(fmt.Sprintf(":%d", s.port), nil)
}

/**
 * Handle a request
 */
func (s *Server) handler(writer http.ResponseWriter, request *http.Request) {
  if s.proxy == nil {
    s.serveRequest(writer, request)
  }else if strings.ToUpper(request.Method) == "GET" {
    switch path.Ext(request.URL.Path) {
      case ".css", ".scss", ".ejs", ".js":
        s.serveRequest(writer, request)
      default:
        s.proxyRequest(writer, request)
    }
  }else{
    s.proxyRequest(writer, request)
  }
}

/**
 * Proxy a request
 */
func (s *Server) proxyRequest(writer http.ResponseWriter, request *http.Request) {
  if s.proxy == nil {
    s.serveError(writer, request, http.StatusBadGateway, fmt.Errorf("No proxy is configured for non-managed resource: %s", request.URL.Path))
  }else if err := s.proxy.ServeHTTP(writer, request); err != nil {
    s.serveError(writer, request, http.StatusBadGateway, err)
  }
}

/**
 * Route a request
 */
func (s *Server) routeRequest(request *http.Request) ([]string, string, error) {
  candidates := make([]string, 0)
  absolute := request.URL.Path
  var mimetype string
  
  for k, v := range s.routes {
    if strings.HasPrefix(absolute, k) {
      absolute = path.Join(v, absolute[len(k):])
      break
    }
  }
  
  ext := path.Ext(absolute)
  relative := absolute[1:]
  base := relative[:len(relative) - len(ext)]
  
  switch ext {
    case ".css":
      candidates = append(candidates, base +".min.scss", base +".scss", relative)
    case ".js":
      candidates = append(candidates, base +".min.ejs", base +".ejs", relative)
    default:
      candidates = append(candidates, relative)
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
  var candidates []string
  var mimetype string
  var file *os.File
  var err error
  
  if candidates, mimetype, err = s.routeRequest(request); err != nil {
    s.serveError(writer, request, http.StatusNotFound, fmt.Errorf("Could not map resource: %s", request.URL.Path))
    return
  }
  
  for _, e := range candidates {
    fmt.Println(request.URL.Path, "->", e)
    
    if file, err = os.Open(e); err == nil {
      defer file.Close()
      writer.Header().Add("Content-Type", mimetype)
      s.compileAndServeFile(writer, request, file)
      return
    }
    
  }
  
  // make this a flag or something...
  strict := false
  
  if strict {
    s.serveError(writer, request, http.StatusNotFound, fmt.Errorf("No such resource: %s", request.URL.Path))
  }else{
    s.proxyRequest(writer, request)
  }
  
}

/**
 * Serve a request
 */
func (s *Server) compileAndServeFile(writer http.ResponseWriter, request *http.Request, file *os.File) {
  
  if fstat, err := file.Stat(); err != nil {
    s.serveError(writer, request, http.StatusBadRequest, fmt.Errorf("Could not stat file: %v", file.Name()))
    return
  }else if fstat.Mode().IsDir() {
    s.serveError(writer, request, http.StatusBadRequest, fmt.Errorf("Resource is not a file: %v", file.Name()))
    return
  }
  
  if compiler, err := NewCompiler(s.context, file.Name()); err != nil {
    s.serveError(writer, request, http.StatusBadRequest, fmt.Errorf("Resource is not supported: %v", file.Name()))
    return
  }else if err := compiler.Compile(s.context, file.Name(), "", file, writer); err != nil {
    s.serveError(writer, request, http.StatusInternalServerError, fmt.Errorf("Could not compile resource: %v", err))
    return
  }
  
}

/**
 * Serve an error
 */
func (s *Server) serveError(writer http.ResponseWriter, request *http.Request, status int, problem error) {
  log.Println(problem)
  if t, err := template.ParseFiles("resources/html/error.html"); err != nil {
    
    fmt.Printf("Could not compile template: %v\n", err)
    writer.WriteHeader(status)
    writer.Write([]byte(problem.Error()))
    
  }else{
    
    params := map[string]interface{} {
      "Title": "Error",
      "Header": fmt.Sprintf("%d: %s", status, http.StatusText(status)),
      "Detail": problem,
    }
    
    writer.Header().Add("Content-Type", "text/html")
    writer.WriteHeader(status)
    t.Execute(writer, params)
    
  }
}

