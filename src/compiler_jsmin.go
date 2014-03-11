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
  "unsafe"
  "io/ioutil"
)

/*
#include <string.h>
#include <stdlib.h>
#include "../dep/jsmin/webasm_jsmin.h"
*/
import "C"

/**
 * A JSMin compiler
 */
type JSMinCompiler struct {
  // ...
}

/**
 * Output path
 */
func (c JSMinCompiler) OutputPath(context Context, inpath string) (string, error) {
  ext := fullExtension(inpath)
  switch ext {
    case ".min.js":
      return inpath[:len(inpath)-len(ext)] +".js", nil
    default:
      return "", fmt.Errorf("Invalid input file extension: %s", ext)
  }
}

/**
 * Compile JSMin
 */
func (c JSMinCompiler) Compile(context Context, inpath, outpath string, input io.Reader, output io.Writer) error {
  var minified *C.char
  var source *C.char
  
  if c, err := ioutil.ReadAll(input); err != nil {
    return err
  }else if source = C.CString(string(c)); source == nil {
    return fmt.Errorf("Input source is invalid")
  }else{
    defer C.free(unsafe.Pointer(source))
  }
  
  if minified = C.jsmin_minify(source); minified == nil {
    return fmt.Errorf("Could not compile input")
  }else{
    defer C.free(unsafe.Pointer(minified))
  }
  
  if _, err := output.Write(C.GoBytes(unsafe.Pointer(minified), C.int(C.strlen(minified)))); err != nil {
    return err
  }
  
  return nil
}

