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
  "io"
  "fmt"
	"path"
  "regexp"
  "io/ioutil"
  "path/filepath"
)

import "ejs"

import (
  "net/http"
)

/**
 * An "extended Javascript" (EJS) compiler
 */
type EJSCompiler struct {
  // ...
}

/**
 * Output path
 */
func (c EJSCompiler) OutputPath(context *Context, inpath string) (string, error) {
  ext := path.Ext(inpath)
  switch ext {
    case ".ejs":
      return inpath[:len(inpath)-len(ext)] +".js", nil
    default:
      return "", fmt.Errorf("Invalid input file extension: %s", ext)
  }
}

/**
 * Compile EJS
 */
func (c EJSCompiler) Compile(context *Context, inpath, outpath string, input io.Reader, output io.Writer) error {
  var source []byte
  var err error
  
  if source, err = ioutil.ReadAll(input); err != nil {
    return err
  }
  
  scanner := ejs.NewScanner(inpath, string(source))
  outer:
  for {
    
    var toks []ejs.Token
    if toks, err = scanner.Token(); err != nil {
      return err
    }
    
    for _, tok := range toks {
      switch tok.Type {
        
        case ejs.TokenTypeEOF:
          break outer
          
        case ejs.TokenTypeVerbatim:
          if _, err := output.Write([]byte(tok.Text)); err != nil {
            return err
          }
        
        case ejs.TokenTypeImport:
          if err := c.emitImport(context, inpath, outpath, output, tok.Text); err != nil {
            return err
          }
        
      }
      
    }
    
  }
  
  return nil
}

/**
 * Emit an import
 */
func (c EJSCompiler) emitImport(context *Context, inpath, outpath string, output io.Writer, resource string) error {
  var absolute string
  
  base := path.Dir(inpath)
  isurl, err := regexp.MatchString("^https?://", resource)
  if err != nil {
    return err
  }else if isurl {
    absolute = resource
  }else if abs, err := filepath.Abs(path.Join(base, resource)); err != nil {
    return err
  }else{
    absolute = abs
  }
  
  if context.IsVisited(absolute) {
    fmt.Println("SKIPPING ALREADY-VISITED RESOURCE:", absolute)
    return nil
  }else{
    context.AddVisited(absolute)
  }
  
  if _, err := output.Write([]byte(fmt.Sprintf("/* #import %+q */\n", resource))); err != nil {
    return err
  }
  
  if isurl {
    return c.emitImportURL(context, inpath, outpath, output, resource)
  }else{
    return c.emitImportFile(context, inpath, outpath, output, resource)
  }
  
}

/**
 * Emit an import
 */
func (c EJSCompiler) emitImportURL(context *Context, inpath, outpath string, output io.Writer, resource string) error {
  
  resp, err := http.Get(resource)
  if err != nil {
    return err
  }
  
  defer resp.Body.Close()
  
  if _, err := io.Copy(output, resp.Body); err != nil {
    return err
  }
  
  return nil
}

/**
 * Emit an import
 */
func (c EJSCompiler) emitImportFile(context *Context, inpath, outpath string, output io.Writer, resource string) error {
  base := path.Dir(inpath)
  if file, err := os.Open(path.Join(base, resource)); err != nil {
    return fmt.Errorf("Could not import file (via %s): %s", inpath, err)
  }else if compiler, err := NewCompiler(context, file.Name()); err != nil {
    return err
  }else{
    return compiler.Compile(context, file.Name(), "", file, output);
  }
}

