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
  "strings"
  "io/ioutil"
  "path"
  "path/filepath"
)

import (
  "github.com/BurntSushi/toml"
  "bitbucket.org/kardianos/osext"
)

const (
  OptionsFlagNone       = 0
  OptionsFlagQuiet      = 1 << 0
  OptionsFlagVerbose    = 1 << 1
  OptionsFlagDebug      = 1 << 2
)

/**
 * Shared options / global config
 */
var __options *Options

/**
 * Options
 */
type Options struct {
  home        string
  Flags       int
  Routes      map[string]string
  Server      ServerOptions
  Stylesheet  StylesheetOptions
  Javascript  JavascriptOptions
  Unmanaged   UnmanagedOptions
}

/**
 * Server options
 */
type ServerOptions struct {
  Port      int                   `toml:"port"`
  Proxy     string                `toml:"proxy"`
  Root      string                `toml:"root"`
}

/**
 * Stylesheet options
 */
type StylesheetOptions struct {
  Minify    bool                  `toml:"minify"`
}

/**
 * Javascript options
 */
type JavascriptOptions struct {
  Minify    bool                  `toml:"minify"`
}

/**
 * Unmanaged options
 */
type UnmanagedOptions struct {
  Copy      bool                  `toml:"copy"`
  Exclude   []string              `toml:"exclude_from_copy"`
}

/**
 * Determine whether a file should be copied
 */
func (u UnmanagedOptions) ShouldCopy(inpath string) bool {
  if u.Copy {
    if u.Exclude == nil || len(u.Exclude) < 1 {
      return true
    }else{
      ext := strings.ToLower(path.Ext(inpath))
      for _, ex := range u.Exclude {
        if ext == strings.ToLower(ex) {
          return false
        }
      }
      return true
    }
  }
  return false
}

/**
 * Config
 */
type config struct {
  Quiet       bool                `toml:"quiet"`
  Verbose     bool                `toml:"verbose"`
  Debug       bool                `toml:"debug"`
  Server      ServerOptions       `toml:"server"`
  Routes      map[string]string   `toml:"routes"`
  Stylesheet  StylesheetOptions   `toml:"stylesheet"`
  Javascript  JavascriptOptions   `toml:"javascript"`
  Unmanaged   UnmanagedOptions    `toml:"unmanaged"`
}

/**
 * Initialize options
 */
func InitOptions(configPath string, inputPaths []string) (*Options) {
  var requireConfig bool
  options := &Options{}
  
  // where are we?
  binary, err := osext.Executable()
  if err != nil { panic(err) }
  home, err := filepath.Abs(filepath.Dir(filepath.Join(binary, "..")))
  if err != nil { panic(err) }
  
  // home directory
  options.home = home
  
  // figure out where our config file should be
  if configPath != "" {
    requireConfig = true
  }else{
    configPath = "./slang.conf"
  }
  
  // load configuration
  if file, err := os.Open(configPath); err != nil {
    
    // if a config file was explicitly provided, it must exist, otherwise, we just
    // ignore a missing configuration file and use defaults
    if requireConfig {
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
    options.Server = conf.Server
    // initialize JS config
    options.Javascript = conf.Javascript
    // initialize CSS config
    options.Stylesheet = conf.Stylesheet
    // initialize unmanaged config
    options.Unmanaged = conf.Unmanaged
    // initialize routes
    options.Routes = conf.Routes
    
    // initialize options
    options.SetFlag(OptionsFlagQuiet,   conf.Quiet)
    options.SetFlag(OptionsFlagVerbose, conf.Verbose && !conf.Quiet)
    options.SetFlag(OptionsFlagDebug,   conf.Debug && !conf.Quiet)
    
  }
  
  // setup shared options
  __options = options
  
  return options
}

/**
 * Obtain the shared options
 */
func SharedOptions() (*Options) {
  if __options != nil {
    return __options
  }else{
    panic(fmt.Errorf("No configuration"));
  }
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
  return path.Join(o.home, "share/slang", relative)
}

/**
 * Set or unset a flag
 */
func (o *Options) GetFlag(flag int) bool {
  return (o.Flags & flag) == flag
}

/**
 * Set or unset a flag
 */
func (o *Options) SetFlag(flag int, set bool) {
  if set {
    o.Flags |= flag
  }else{
    o.Flags &^= flag
  }
}


