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
  "io"
  "path"
  "strings"
)

const (
  compilerOptionNone        = 0
  compilerOptionVerbose     = 1 << 0
)

/**
 * Compilation context
 */
type Context struct {
  Options   int
  visited   map[string]bool
}

/**
 * Create a compiler context
 */
func NewContext() *Context {
  return &Context{0, make(map[string]bool)}
}

/**
 * Add a visited resource
 */
func (c *Context) AddVisited(resource string) {
  c.visited[resource] = true
}

/**
 * Is a resource visited
 */
func (c *Context) IsVisited(resource string) bool {
  if _, ok := c.visited[resource]; ok {
    return true
  }else{
    return false
  }
}

/**
 * Clear the visited resource map
 */
func (c *Context) ClearVisited() {
  for k := range c.visited {
    delete(c.visited, k)
  }
}

/**
 * A compiler
 */
type Compiler interface {
  OutputPath(context *Context, inpath string) (string, error)
  Compile(context *Context, inpath, outpath string, input io.Reader, output io.Writer) error
}

/**
 * Determine if a resource can be compiled
 */
func CanCompile(context *Context, inpath string) bool {
  switch fullExtension(inpath) {
    case ".min.scss", ".scss", ".min.js", ".ejs", ".min.ejs":
      return true
    default:
      return false
  }
}

/**
 * Create the default compiler for the specified file
 */
func NewCompiler(context *Context, inpath string) (Compiler, error) {
  switch fullExtension(inpath) {
    case ".min.scss", ".min.css":
      return &SassCompiler{sassOptionCompress}, nil
    case ".scss":
      return &SassCompiler{}, nil
    case ".min.ejs":
      return CompilerChain([]Compiler{ &EJSCompiler{}, &JSMinCompiler{} }), nil
    case ".ejs":
      return &EJSCompiler{}, nil
    case ".min.js":
      return &JSMinCompiler{}, nil 
    default:
      return &LiteralCompiler{}, nil
  }
}

/**
 * Obtain every extension for a path
 */
func fullExtension(p string) string {
  base := path.Base(p)
  if i := strings.Index(base, "."); i > 0 {
    return base[i:]
  }else{
    return ""
  }
}

