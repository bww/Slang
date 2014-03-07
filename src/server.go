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
  "path"
)

import (
  "net/url"
  "net/http"
  "net/http/httputil"
)

/**
 * A server
 */
type Server struct {
  Port    int
  Route   string
  proxy   *httputil.ReverseProxy
}

/**
 * Create a server
 */
func NewServer(port int, route string) (*Server, error) {
  var proxy *httputil.ReverseProxy = nil
  
  if route != "" {
    if u, err := url.Parse(fmt.Sprintf("http://%s/", route)); err != nil {
      return nil, err
    }else{
      proxy = httputil.NewSingleHostReverseProxy(u)
    }
  }
  
  return &Server{port, route, proxy}, nil
}

/**
 * Run the server
 */
func (s *Server) Run() {
  http.HandleFunc("/", s.handler)
  http.ListenAndServe(fmt.Sprintf(":%d", s.Port), nil)
}

/**
 * Handle a request
 */
func (s *Server) handler(writer http.ResponseWriter, request *http.Request) {
  switch path.Ext(request.URL.Path) {
    case ".css", ".scss", ".js":
      s.serveRequest(writer, request)
    default:
      s.proxyRequest(writer, request)
  }
}

/**
 * Proxy a request
 */
func (s *Server) proxyRequest(writer http.ResponseWriter, request *http.Request) {
  s.proxy.ServeHTTP(writer, request)
}

/**
 * Serve a request
 */
func (s *Server) serveRequest(writer http.ResponseWriter, request *http.Request) {
  switch path.Ext(request.URL.Path) {
    case ".css":
      s.serveCounterpart(writer, request)
    case ".scss":
      s.serveProcessed(writer, request)
    case ".js":
      s.serveCounterpart(writer, request)
    case ".ejs":
      s.serveProcessed(writer, request)
    default:
      s.proxyRequest(writer, request)
  }
}

/**
 * Serve a request
 */
func (s *Server) serveCounterpart(writer http.ResponseWriter, request *http.Request) {
  
}

/**
 * Serve a request
 */
func (s *Server) serveProcessed(writer http.ResponseWriter, request *http.Request) {
  var file *os.File
  var err error
  
  fmt.Println("->", request.URL.Path)
  
  context := Context{}
  
  if len(request.URL.Path) < 1 {
    fmt.Printf("Invalid path: %s\n", request.URL.Path)
    s.serveError(writer, request, 400, fmt.Errorf("Invalid path: %s\n", request.URL.Path));
    return
  }
  
  relative := request.URL.Path
  
  if relative[0] == '/' {
    relative = relative[1:]
  }
  
  if file, err = os.Open(relative); err != nil {
    fmt.Printf("Could not open file: %v\n", err)
    s.serveError(writer, request, 404, err);
    return
  }
  
  if compiler, err := NewCompiler(context, request.URL.Path); err != nil {
    fmt.Printf("Resource is not supported: %v\n", err)
    s.serveError(writer, request, 400, fmt.Errorf("Resource is not supported: %s", relative));
    return
  }else if err := compiler.Compile(context, "", "", file, writer); err != nil {
    fmt.Printf("Could not compile resource: %v\n", err)
    s.serveError(writer, request, 500, fmt.Errorf("Could not compile resource: %v", err));
    return
  }
  
}

/**
 * Serve an error
 */
func (s *Server) serveError(writer http.ResponseWriter, request *http.Request, status int, err error) {
  writer.WriteHeader(status)
  writer.Write([]byte(err.Error()))
}

