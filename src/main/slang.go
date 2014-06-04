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
  "strings"
  "path/filepath"
)

import (
  "io/ioutil"
  "encoding/json"
)

const (
  COMMAND_INIT    = "init"
  COMMAND_RUN     = "run"
  COMMAND_BUILD   = "build"
  COMMAND_HELP    = "help"
)

func main() {
  
  if len(os.Args) < 2 {
    runHelp(nil, false)
    return
  }
  
  // note the command
  command     := os.Args[1]
  
  // process the command line
  cmdline     := flag.NewFlagSet(os.Args[0], flag.ExitOnError)
  fConfig     := cmdline.String ("conf",        "",             "Specify a particular configuration to use.")
  fVariables  := cmdline.String ("vars",        "",             "Specify a variable set to use, defined in a JSON file.")
  
  fPort       := cmdline.Int    ("port",        9090,           "The port on which to run the built-in server.")
  fProxy      := cmdline.String ("proxy",       "",             "The base URL the built-in server should reverse-proxy for unmanaged resources.")
  fRoutes     := make(AssocParams)
  cmdline.Var(&fRoutes, "route", "Routing rules, formatted as '<remote>=<local>'; e.g., slang -server -route /css=/styles -route /js=/app/js [...].")
  
  fOutput     := cmdline.String ("output",      "./slang.out",  "Specify the path to write compiled resources to.")
  fCopy       := cmdline.Bool   ("copy",        false,          "Copy unmanaged resources to output when compiling.")
  
  fMinify     := cmdline.Bool   ("minify",      false,          "Minify resources that can be minified.")
  fMinifyCSS  := cmdline.Bool   ("css:minify",  false,          "Minify stylesheets resources.")
  fMinifyJS   := cmdline.Bool   ("js:minify",   false,          "Minify Javascript resources.")
  fShip       := cmdline.Bool   ("ship",        false,          "Turn on all presets for shipping a project.")
  
  fQuiet      := cmdline.Bool   ("quiet",       false,          "Be quiet. Only print error messages. (Overrides -verbose, -debug)")
  fVerbose    := cmdline.Bool   ("verbose",     false,          "Be more verbose.")
  fDebug      := cmdline.Bool   ("debug",       false,          "Be extremely verbose.")
  
  cmdline.Parse(os.Args[2:])
  
  // initialize our options
  options := InitOptions(*fConfig, cmdline.Args())
  
  // do init if requested and exit
  if command == COMMAND_INIT {
    runInit(options); return
  }
  
  // variables
  if *fVariables != "" {
    variables := make(map[string]interface{})
    if serial, err := ioutil.ReadFile(*fVariables); err != nil {
      fmt.Printf("Could not read variables: %v\n", err)
      return
    }else if err := json.Unmarshal(serial, &variables); err != nil{
      fmt.Printf("Variables are not valid: %v\n", err)
      return
    }
    options.Variables = variables
  }
  
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
  
  // compilation options
  if *fShip || *fMinify || *fMinifyCSS { options.Stylesheet.Minify = true }
  if *fShip || *fMinify || *fMinifyJS  { options.Javascript.Minify = true }
  
  // unmanaged resource options
  if *fCopy { options.Unmanaged.Copy = true }
  
  // apply command line flags
  if *fQuiet    { options.SetFlag(OptionsFlagQuiet,   *fQuiet) }
  if *fVerbose  { options.SetFlag(OptionsFlagVerbose, *fVerbose && !options.GetFlag(OptionsFlagQuiet)) }
  if *fDebug    { options.SetFlag(OptionsFlagDebug,   *fDebug   && !options.GetFlag(OptionsFlagQuiet)) }
  
  // do something useful
  if command == COMMAND_RUN {
    runServer(options, cmdline.Args())
  }else if command == COMMAND_BUILD {
    runCompile(options, *fOutput, cmdline.Args())
  }else if command == COMMAND_HELP {
    runHelp(cmdline, true)
  }else{
    runHelp(nil, false)
  }
  
}

/**
 * Display help info
 */
func runHelp(cmdline *flag.FlagSet, detail bool) {
  
  if !detail {
    fmt.Println("Usage: slang (run|build|init) [options]");
    fmt.Println(" Help: slang help");
  }else{
    fmt.Println("Usage: slang (run|build|init) [options]");
    fmt.Println()
    fmt.Println("Initialize an optional slang.conf file:")
    fmt.Println("  $ slang init")
    fmt.Println()
    fmt.Println("Start the built-in server:")
    fmt.Println("  $ slang run [./docroot]")
    fmt.Println()
    fmt.Println("Traverse a directory and compile all supported assets found in it:")
    fmt.Println("  $ slang build -output ./build ./assets")
    fmt.Println()
  }
  
  if cmdline != nil {
    fmt.Println("Options:")
    cmdline.PrintDefaults()
    fmt.Println()
  }
  
}

/**
 * Initialize config
 */
func runInit(options *Options) {
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
  
  fmt.Printf("Default configuration created at: %s\n", configPath);
  
}

/**
 * Service
 */
func runServer(options *Options, args []string) {
  var server *Server
  var root string
  var err error
  
  if len(args) > 0 {
    root = args[0]
  }else if options.Server.Root != "" {
    root = options.Server.Root
  }else{
    root = "."
  }
  
  if server, err = NewServer(options.Server.Port, options.Server.Proxy, root, options.Routes); err != nil {
    fmt.Println(err)
    return
  }
  
  if options.Server.Proxy != "" {
    fmt.Printf("Starting the Slang server: http://localhost:%d/ <-> %s\n", options.Server.Port, options.Server.Proxy)
  }else{
    fmt.Printf("Starting the Slang server: http://localhost:%d/\n", options.Server.Port)
  }
  
  if err := server.Run(); err != nil {
    fmt.Println(err)
    return
  }
  
}

/**
 * Compile
 */
func runCompile(options *Options, outbase string, args []string) {
  
  if err := os.MkdirAll(outbase, 0755); err != nil {
    fmt.Println(err)
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
      w := &Walker{f, outbase}
      if err := filepath.Walk(input.Name(), w.compileResource); err != nil {
        fmt.Println(err)
        return
      }
    }else{
      w := &Walker{filepath.Dir(f), outbase}
      if err := filepath.Walk(input.Name(), w.compileResource); err != nil {
        fmt.Println(err)
        return
      }
    }
    
  }
  
}

/**
 * Process a resource
 */
func processResource(context *Context, info os.FileInfo, inpath, outpath string, input *os.File, output io.Writer) error {
  if CanCompile(context, inpath) {
    if !SharedOptions().ShouldExclude(inpath) {
      if !SharedOptions().GetFlag(OptionsFlagQuiet) { fmt.Printf("[+] %s\n", inpath) }
      return compileResource(context, info, inpath, outpath, input, output)
    }else{
      if !SharedOptions().GetFlag(OptionsFlagQuiet) { fmt.Printf("[ ] %s\n", inpath) }
      return nil
    }
  }else if SharedOptions().Unmanaged.ShouldCopy(inpath) {
    if !SharedOptions().GetFlag(OptionsFlagQuiet) { fmt.Printf("[~] %s\n", inpath) }
    return copyResource(context, info, inpath, outpath, input, output)
  }else{
    if !SharedOptions().GetFlag(OptionsFlagQuiet) { fmt.Printf("[ ] %s\n", inpath) }
    return nil
  }
}

/**
 * Compile a resource
 */
func compileResource(context *Context, info os.FileInfo, inpath, outpath string, input *os.File, output io.Writer) error {
  var compiler Compiler
  var err error
  
  compiler, err = NewCompiler(context, inpath)
  if err != nil {
    return err
  }
  
  outpath, err = compiler.OutputPath(context, outpath)
  if err != nil {
    return err
  }
  
  // if we aren't provided an explicit output stream, open the output file and use that
  if output == nil {
    outfile, err := os.OpenFile(outpath, os.O_WRONLY | os.O_CREATE | os.O_TRUNC, 0644)
    if err != nil {
      return err
    }else{
      defer outfile.Close()
    }
    output = outfile
  }
  
  err = compiler.Compile(context, inpath, outpath, input, output)
  if err != nil {
    return err
  }
  
  return nil
}

/**
 * Copy a resource
 */
func copyResource(context *Context, info os.FileInfo, inpath, outpath string, input *os.File, output io.Writer) error {
  
  // if we aren't provided an explicit output stream, open the output file and use that
  if output == nil {
    outfile, err := os.OpenFile(outpath, os.O_WRONLY | os.O_CREATE | os.O_TRUNC, 0644)
    if err != nil {
      return err
    }else{
      defer outfile.Close()
    }
    output = outfile
  }
  
  // copy our resource over
  if _, err := io.Copy(output, input); err != nil {
    return err
  }
  
  return nil
}

/**
 * Walk context
 */
type Walker struct {
  inbase    string
  outbase   string
}

/**
 * Relocate a resource from the input base to the output base
 */
func (w Walker) relocateResource(path string) (string, error) {
  
  absin, err := filepath.Abs(w.inbase)
  if err != nil {
    return "", fmt.Errorf("Could not make input base absolute: %s", path)
  }
  
  absout, err := filepath.Abs(w.outbase)
  if err != nil {
    return "", fmt.Errorf("Could not make output base absolute: %s", path)
  }
  
  abspath, err := filepath.Abs(path)
  if err != nil {
    return "", fmt.Errorf("Could not make input path absolute: %s", path)
  }
  
  if len(abspath) <= len(absin) {
    return "", fmt.Errorf("Input path is not under input base: %s", path)
  }else if !strings.HasPrefix(abspath, absin) {
    return "", fmt.Errorf("Input path is not under input base: %s", path)
  }
  
  return filepath.Join(absout, abspath[len(absin)+1:]), nil
}

/**
 * Compile a resource
 */
func (w Walker) compileResource(path string, info os.FileInfo, err error) error {
  hidden := info.Name() != "." && info.Name()[0] == '.'
  
  if err != nil {
    return err
  }else if path == w.inbase {
    return nil // just descend into the input base path
  }
  
  outpath, err := w.relocateResource(path)
  if err != nil {
    return err
  }
  
  if info.Mode().IsDir() {
    if hidden {
      return filepath.SkipDir // skip hidden directories
    }else if strings.HasPrefix(path, w.outbase) {
      return filepath.SkipDir // skip the tree under the output root
    }else if err := os.Mkdir(outpath, 0755); err != nil && !os.IsExist(err) {
      return err // could not create output directory
    }else{
      return nil // just descend
    }
  }else{
    if hidden {
      return nil // skip hidden files
    }
  }
  
  input, err := os.Open(path)
  if err != nil {
    return err
  }else{
    defer input.Close()
  }
  
  return processResource(NewContext(), info, input.Name(), outpath, input, nil)
}


