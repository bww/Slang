// 
// Copyright (c) 2014 Brian William Wolter, All rights reserved.
// Service Minder
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
#include <stdlib.h>
#include "../dep/libsass/sass_interface.h"

const char *SASS_EMTPY_STRING = "";
*/
import "C"

/**
 * A SASS compiler
 */
type SassCompiler struct {
  // ...
}

/**
 * Compile SASS
 */
func (c SassCompiler) Compile(context Context, inpath, outpath string, input io.Reader, output io.Writer) error {
  var sass *C.struct_sass_context
  var options C.struct_sass_options
  
  fmt.Println("-->", inpath)
  
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
  
  options.include_paths = C.SASS_EMTPY_STRING
  options.image_path = C.SASS_EMTPY_STRING
  sass.options = options
  
  C.sass_compile(sass)
  
  if sass.error_status != 0 {
    return fmt.Errorf("Could not compile SASS: %s", C.GoString(sass.error_message))
  }else if sass.output_string != nil {
    fmt.Println(C.GoString(sass.output_string))
  }else{
    return fmt.Errorf("An unknown error occured; SASS produced no output")
  }
  
  return nil
}

