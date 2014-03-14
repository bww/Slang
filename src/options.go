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

import "fmt"

import (
  "os"
  "flag"
  "path"
  "path/filepath"
  "io/ioutil"
)

import (
  "github.com/BurntSushi/toml"
)

const (
  
  optionsFlagNone       = 0
  optionsFlagQuiet      = 1 << 0
  optionsFlagVerbose    = 1 << 1
  optionsFlagDebug      = 1 << 2
  
  optionsModeServer     = 0
  optionsModeCompile    = 1
  optionsModeInit       = 2
  
)

var __options *Options

/**
 * Options
 */
type Options struct {
  home      string
  flags     int
  mode      int
  routes    map[string]string
  server    ServerOptions
}

/**
 * Server options
 */
type ServerOptions struct {
  Port      int                   `toml:"port"`
  Proxy     string                `toml:"proxy"`
}

/**
 * Config
 */
type config struct {
  Quiet     bool                  `toml:"quiet"`
  Verbose   bool                  `toml:"verbose"`
  Debug     bool                  `toml:"debug"`
  Server    ServerOptions         `toml:"server"`
  Routes    map[string]string     `toml:"routes"`
}

/**
 * Obtain the shared options
 */
func SharedOptions() (*Options) {
  if __options != nil {
    return __options
  }else if __options = initOptions(); __options != nil {
    return __options
  }else{
    panic(fmt.Errorf("Could not create configuration"));
  }
}

/**
 * Initialize options
 */
func initOptions() (*Options) {
  options := &Options{}
  
  // where are we?
  home, err := filepath.Abs(filepath.Dir(os.Args[0]))
  if err != nil { panic(err) }
  
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
  
  // figure out where our config file should be
  var configPath string
  if *fConfig != "" {
    configPath = *fConfig
  }else{
    configPath = "./slang.conf"
  }
  
  // load configuration
  if file, err := os.Open(configPath); err != nil {
    
    // if a config file was explicitly provided, it must exist, otherwise, we just
    // ignore a missing configuration file and use defaults
    if *fConfig != "" {
      fmt.Printf("No such configuration: %v\n", err)
      os.Exit(-1)
    }
    
  }else{
    defer file.Close()
    conf := &config{}
    
    // load our configuration
    if c, err := ioutil.ReadAll(file); err != nil {
      fmt.Printf("Could not read configuration: %v\n", err)
      os.Exit(-1)
    }else if _, err := toml.Decode(string(c), conf); err != nil {
      fmt.Printf("Configuration is not valid: %v\n", err)
      os.Exit(-1)
    }
    
    // initialize server config
    options.server = conf.Server
    // initialize routes
    options.routes = conf.Routes
    
    // initialize options
    options.SetFlag(optionsFlagQuiet,   conf.Quiet)
    options.SetFlag(optionsFlagVerbose, conf.Verbose && !conf.Quiet)
    options.SetFlag(optionsFlagDebug,   conf.Debug && !conf.Quiet)
    
  }
  
  // home directory
  options.home = home
  
  // run mode
  if *fInit {
    options.mode = optionsModeInit
  }else if *fServer {
    options.mode = optionsModeServer
  }else{
    options.mode = optionsModeCompile
  }
  
  // server config
  if *fPort != options.server.Port {
    options.server.Port = *fPort
  }
  if *fProxy != "" {
    options.server.Proxy = *fProxy
  }
  
  // routes definitions
  if len(fRoutes) > 0 {
    options.routes = fRoutes
  }
  
  // apply command line flags
  if *fQuiet    { options.SetFlag(optionsFlagQuiet,   *fQuiet) }
  if *fVerbose  { options.SetFlag(optionsFlagVerbose, *fVerbose && !options.GetFlag(optionsFlagQuiet)) }
  if *fDebug    { options.SetFlag(optionsFlagDebug,   *fDebug && !options.GetFlag(optionsFlagQuiet)) }
  
  __options = options
  
  return options
}

/**
 * Obtain our home path
 */
func (o *Options) Home() string {
  return o.home
}

/**
 * Obtain a resource path
 */
func (o *Options) Resource(relative string) string {
  return path.Join(o.home, "resources", relative)
}

/**
 * Obtain the run mode
 */
func (o *Options) Mode() int {
  return o.mode
}

/**
 * Obtain routes
 */
func (o *Options) Routes() map[string]string {
  return o.routes
}

/**
 * Obtain server config
 */
func (o *Options) Server() ServerOptions {
  return o.server
}

/**
 * Set or unset a flag
 */
func (o *Options) GetFlag(flag int) bool {
  return (o.flags & flag) == flag
}

/**
 * Set or unset a flag
 */
func (o *Options) SetFlag(flag int, set bool) {
  if set {
    o.flags |= flag
  }else{
    o.flags &^= flag
  }
}


