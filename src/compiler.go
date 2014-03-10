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
  "io"
  "fmt"
  "path"
  "bytes"
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
}

/**
 * A compiler
 */
type Compiler interface {
  Compile(context Context, inpath, outpath string, input io.Reader, output io.Writer) error
}

/**
 * A compiler chain
 */
type CompilerChain []Compiler

/**
 * Compile for a chain
 */
func (c CompilerChain) Compile(context Context, inpath, outpath string, input io.Reader, output io.Writer) error {
  var w *bytes.Buffer
  
  for i, e := range c {
    var r io.Reader
    
    if i == 0 {
      r = input
    }else{
      r = w
    }
    
    w = bytes.NewBuffer(make([]byte, 0))
    
    if err := e.Compile(context, inpath, outpath, r, w); err != nil {
      return err
    }
    
  }
  
  if _, err := w.WriteTo(output); err != nil {
    return err
  }
  
  return nil
}

/**
 * Create the default compiler for the specified file
 */
func NewCompiler(context Context, inpath string) (Compiler, error) {
  var ext string
  
  base := path.Base(inpath)
  
  if i := strings.Index(base, "."); i > 0 {
    ext = base[i:]
  }else{
    return nil, fmt.Errorf("Input file has no extension: %s", base)
  }
  
  switch ext {
    case ".scss":
      return &SassCompiler{}, nil
    case ".min.ejs":
      return CompilerChain([]Compiler{ &EJSCompiler{}, &JSMinCompiler{} }), nil
    case ".ejs":
      return &EJSCompiler{}, nil
    case ".min.js":
      return &JSMinCompiler{}, nil 
    case ".js":
      return &LiteralCompiler{}, nil
    default:
      return &LiteralCompiler{}, nil
  }
  
}

