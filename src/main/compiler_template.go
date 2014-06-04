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
  "fmt"
  "path"
  "io/ioutil"
)

import (
  "html/template"
)

/**
 * A template compiler
 */
type TemplateCompiler struct {
  // ...
}

/**
 * Output path
 */
func (c TemplateCompiler) OutputPath(context *Context, inpath string) (string, error) {
  ext := path.Ext(inpath)
  switch ext {
    case ".ghtml":
      return ".html", nil
    case ".html":
      return ".html", nil
    default:
      return "", fmt.Errorf("Invalid input file extension: %s", ext)
  }
}

/**
 * Compile JSMin
 */
func (c TemplateCompiler) Compile(context *Context, inpath, outpath string, input io.Reader, output io.Writer) error {
  
  serial, err := ioutil.ReadAll(input)
  if err != nil {
    return fmt.Errorf("Could not read template: %v\n", err)
  }
  
  base := template.New(inpath)
  
  t, err := base.Parse(string(serial))
  if err != nil {
    return fmt.Errorf("Could not parse template: %v\n", err)
  }
  
  
  
  if err := t.Execute(output, context.Variables); err != nil {
    return fmt.Errorf("Could not execute template: %v\n", err)
  }
  
  return nil
}

