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
  "os"
  "io"
  "fmt"
  "flag"
  "path/filepath"
)

var OPTIONS *Options

func main() {
  
  // process the command line
  cmdline   := flag.NewFlagSet(os.Args[0], flag.ExitOnError)
  fConfig   := cmdline.String ("config",  "",     "Specify a particular configuration to use.")
  fServer   := cmdline.Bool   ("server",  false,  "Run the built-in server.")
  fPort     := cmdline.Int    ("port",    9090,   "The port on which to run the built-in server.")
  fProxy    := cmdline.String ("proxy",   "",     "The base URL the built-in server should reverse-proxy for unmanaged resources.")
  fQuiet    := cmdline.Bool   ("quiet",   false,  "Be quiet. Only print error messages. (Overrides -verbose, -debug)")
  fVerbose  := cmdline.Bool   ("verbose", false,  "Be more verbose.")
  fDebug    := cmdline.Bool   ("debug",   false,  "Be extremely verbose.")
  fInit     := cmdline.Bool   ("init",    false,  "Initialize a Slang configuration file in the current directory.")
  fRoutes   := make(AssocParams)
  cmdline.Var(&fRoutes, "route", "Routing rules, formatted as '<remote>=<local>'; e.g., slang -server -route /css=/styles -route /js=/app/js [...].")
  cmdline.Parse(os.Args[1:]);
  
  // initialize our options
  options := InitOptions(*fConfig)
  
  // server config
  if *fPort != options.Server.Port {
    options.Server.Port = *fPort
  }
  if *fProxy != "" {
    options.Server.Proxy = *fProxy
  }
  
  // routes definitions
  if len(fRoutes) > 0 {
    options.Routes = fRoutes
  }
  
  // apply command line flags
  if *fQuiet    { options.SetFlag(OptionsFlagQuiet,   *fQuiet) }
  if *fVerbose  { options.SetFlag(OptionsFlagVerbose, *fVerbose && !options.GetFlag(OptionsFlagQuiet)) }
  if *fDebug    { options.SetFlag(OptionsFlagDebug,   *fDebug   && !options.GetFlag(OptionsFlagQuiet)) }
  
  // do something useful
  if(*fInit){
    runInit()
  }else if(*fServer){
    runServer(options.Server.Port, options.Server.Proxy, options.Routes)
  }else{
    runCompile(cmdline)
  }
  
}

/**
 * Initialize config
 */
func runInit() {
  configPath := "./slang.conf"
  
  if _, err := os.Stat(configPath); err == nil {
    fmt.Printf("A config file already exists at: %s. Remove it to init a new config file.\n", configPath);
    return
  }else if !os.IsNotExist(err) {
    fmt.Printf("Could not access config file at: %s\n", configPath);
    return
  }
  
  if infile, err := os.Open(SharedOptions().Resource("conf/default.conf")); err != nil {
    fmt.Println(err)
    return
  }else if outfile, err := os.OpenFile(configPath, os.O_WRONLY | os.O_CREATE, 0644); err != nil {
    fmt.Println(err)
    return
  }else if _, err := io.Copy(outfile, infile); err != nil {
    fmt.Println(err)
    return
  }
  
}

/**
 * Service
 */
func runServer(port int, peer string, routes map[string]string) {
  var server *Server
  var err error
  
  if server, err = NewServer(port, peer, routes); err != nil {
    fmt.Println(err)
    return
  }
  
  fmt.Printf("Starting the Slang server on: http://localhost:%d/\n", port)
  
  if err := server.Run(); err != nil {
    fmt.Println(err)
    return
  }
  
}

/**
 * Compile
 */
func runCompile(cmdline *flag.FlagSet) {
  
  args := cmdline.Args()
  
  if len(args) < 1 {
    fmt.Println("No resources provided to compile. Run Slang as one of the following.")
    fmt.Println()
    fmt.Println("Start the built-in server:")
    fmt.Println("  $ slang -server [...]")
    fmt.Println()
    fmt.Println("Compile specific assets:")
    fmt.Println("  $ slang file.scss file.ejs")
    fmt.Println()
    fmt.Println("Traverse a directory and compile all supported assets found in it:")
    fmt.Println("  $ slang assets/")
    fmt.Println()
    fmt.Println("Show command line options:")
    fmt.Println("  $ slang -h")
    fmt.Println()
    return
  }
  
  for _, f := range args {
    var input *os.File
    var fstat os.FileInfo
    var err error
    
    if input, err = os.Open(f); err != nil {
      fmt.Println(err)
      return
    }
    
    defer input.Close()
    
    if fstat, err = input.Stat(); err != nil {
      fmt.Println(err)
      return
    }
    
    if fstat.Mode().IsDir() {
      w := &Walker{}
      if err := filepath.Walk(input.Name(), w.compileResource); err != nil {
        fmt.Println(err)
        return
      }
    }else{
      c := NewContext()
      if err := compileResource(c, input, fstat); err != nil {
        fmt.Println(err)
        return
      }
    }
    
  }
  
}

/**
 * Compile a resource
 */
func compileResource(context *Context, input *os.File, info os.FileInfo) error {
  inpath := input.Name()
  
  if !CanCompile(context, inpath) {
    if !OPTIONS.GetFlag(OptionsFlagQuiet) { fmt.Printf("[ ] %s\n", inpath) }
    return nil
  }else{
    if !OPTIONS.GetFlag(OptionsFlagQuiet) { fmt.Printf("[+] %s\n", inpath) }
  }
  
  if compiler, err := NewCompiler(context, inpath); err != nil {
    return err
  }else if outpath, err := compiler.OutputPath(context, inpath); err != nil {
    return err
  }else if output, err := os.OpenFile(outpath, os.O_WRONLY | os.O_CREATE | os.O_TRUNC, 0644); err != nil {
    return err
  }else if err := compiler.Compile(context, inpath, outpath, input, output); err != nil {
    return err
  }else{
    return nil
  }
  
}

/**
 * Walk context
 */
type Walker struct {
  // ...
}

/**
 * Compile a resource
 */
func (w Walker) compileResource(path string, info os.FileInfo, err error) error {
  
  if err != nil {
    return err
  }
  
  hidden := info.Name() != "." && info.Name()[0] == '.'
  
  if info.Mode().IsDir() {
    if hidden {
      return filepath.SkipDir // skip hidden directories
    }else{
      return nil // just descend
    }
  }
  
  if hidden {
    return nil // skip hidden files
  }
  
  input, err := os.Open(path)
  if err != nil {
    return err
  }else{
    defer input.Close()
  }
  
  return compileResource(NewContext(), input, info)
}


