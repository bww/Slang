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
  "unsafe"
  "io/ioutil"
)

/*
#include <string.h>
#include <stdlib.h>
#include "../../dep/libsass/sass_interface.h"

const char *SASS_EMTPY_STRING = "";
*/
import "C"

const (
  sassOptionNone        = 0
  sassOptionCompress    = 1 << 0
)

/**
 * A SASS compiler
 */
type SassCompiler struct {
  options   int
}

/**
 * Output path
 */
func (c SassCompiler) OutputPath(context *Context, inpath string) (string, error) {
  ext := path.Ext(inpath)
  switch ext {
    case ".scss":
      return inpath[:len(inpath)-len(ext)] +".css", nil
    case ".css":
      return inpath, nil
    default:
      return "", fmt.Errorf("Invalid input file extension: %s", ext)
  }
}

/**
 * Compile SASS
 */
func (c SassCompiler) Compile(context *Context, inpath, outpath string, input io.Reader, output io.Writer) error {
  var sass *C.struct_sass_context
  var options C.struct_sass_options
  
  if sass = C.sass_new_context(); sass == nil {
    return fmt.Errorf("Could not create SASS context")
  }else{
    defer C.sass_free_context(sass)
  }
  
  if c, err := ioutil.ReadAll(input); err != nil {
    return err
  }else if sass.source_string = C.CString(string(c)); sass.source_string == nil {
    return fmt.Errorf("Input source is invalid")
  }else{
    defer C.free(unsafe.Pointer(sass.source_string))
  }
  
  if (c.options & sassOptionCompress) == sassOptionCompress {
    options.output_style = C.SASS_STYLE_COMPRESSED
  }else{
    options.output_style = C.SASS_STYLE_EXPANDED
  }
  
  // use the input path's directory as our include path
  indir := path.Dir(inpath)
  includePath := C.CString(indir)
  defer C.free(unsafe.Pointer(includePath))
  
  options.include_paths = includePath
  options.image_path = C.SASS_EMTPY_STRING
  
  sass.options = options
  C.sass_compile(sass)
  
  if sass.error_status != 0 {
    return fmt.Errorf("Could not compile SASS: %s", C.GoString(sass.error_message))
  }else if sass.output_string == nil {
    return fmt.Errorf("An unknown error occured; SASS produced no output")
  }else if _, err := output.Write(C.GoBytes(unsafe.Pointer(sass.output_string), C.int(C.strlen(sass.output_string)))); err != nil {
    return err
  }
  
  return nil
}

