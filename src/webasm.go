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
  "flag"
)

func main() {
  
  cmdline := flag.NewFlagSet(os.Args[0], flag.ExitOnError)
  fServer := cmdline.Bool   ("server",  false,                    "Run the builtin server")
  fPort   := cmdline.Int    ("port",    9090,                     "The port to run the builtin server on")
  fPeer   := cmdline.String ("proxy",   "",                       "The address to reverse-proxy")
  
  fRoutes := make(AssocParams)
  cmdline.Var(&fRoutes, "route", "Routes, formatted as '<remote>=<local>'")
  cmdline.Parse(os.Args[1:]);
  
  context := Context{}
  
  if(*fServer){
    var server *Server
    var err error
    
    if server, err = NewServer(context, *fPort, *fPeer, fRoutes); err != nil {
      fmt.Println(err)
      return
    }
    
    fmt.Printf("Starting Webasm on port %d\n", *fPort)
    server.Run()
    
  }else{
    for _, f := range cmdline.Args() {
      var compiler Compiler
      var input io.Reader
      var err error
      
      if input, err = os.Open(f); err != nil {
        fmt.Println(err)
        return
      }
      
      if compiler, err = NewCompiler(context, f); err != nil {
        fmt.Println(err)
        return
      }
      
      if err := compiler.Compile(context, f, f +".out", input, os.Stdout); err != nil {
        fmt.Println(err)
        return
      }
      
    }
  }
  
}

