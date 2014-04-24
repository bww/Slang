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
  "fmt"
  "log"
  "io/ioutil"
  "path"
  "path/filepath"
  "reflect"
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

const (
  CONFIG_PATH_DEFAULT   = "./slang.conf"
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
  Routes      map[string][]string
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
  Exclude   []string              `toml:"exclude"`
}

/**
 * Javascript options
 */
type JavascriptOptions struct {
  Minify    bool                  `toml:"minify"`
  Exclude   []string              `toml:"exclude"`
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
    }else  {
      return !shouldExclude(inpath, u.Exclude)
    }
  }
  return false
}

/**
 * Config
 */
type config struct {
  Quiet       *bool                   `toml:"quiet"`
  Verbose     *bool                   `toml:"verbose"`
  Debug       *bool                   `toml:"debug"`
  Server      serverConfig            `toml:"server"`
  Routes      map[string]interface{}  `toml:"routes"`
  Stylesheet  stylesheetConfig        `toml:"stylesheet"`
  Javascript  javascriptConfig        `toml:"javascript"`
  Unmanaged   unmanagedConfig         `toml:"unmanaged"`
}

/**
 * Server config
 */
type serverConfig struct {
  Port      *int                      `toml:"port"`
  Proxy     *string                   `toml:"proxy"`
  Root      *string                   `toml:"root"`
}

/**
 * Stylesheet config
 */
type stylesheetConfig struct {
  Minify    *bool                     `toml:"minify"`
  Exclude   *[]string                 `toml:"exclude"`
}

/**
 * Javascript config
 */
type javascriptConfig struct {
  Minify    *bool                     `toml:"minify"`
  Exclude   *[]string                 `toml:"exclude"`
}

/**
 * Unmanaged config
 */
type unmanagedConfig struct {
  Copy      *bool                     `toml:"copy"`
  Exclude   *[]string                 `toml:"exclude_from_copy"`
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
    configPath = CONFIG_PATH_DEFAULT
  }
  
  // check out our file
  if _, err := os.Stat(configPath); err != nil {
    if !os.IsNotExist(err) {
      fmt.Printf("Could not stat configuration: %v\n", err)
      os.Exit(-1)
    }else if requireConfig {
      fmt.Printf("No such configuration: %v\n", err)
      os.Exit(-1)
    }
  }else{
    if err := options.loadOptions(configPath); err != nil {
      fmt.Println(err)
      os.Exit(-1)
    }
  }
  
  // setup shared options
  __options = options
  
  return options
}

/**
 * Load options
 */
func (o *Options) loadOptions(configPath string) error {
  
  // load configuration
  file, err := os.Open(configPath)
  if err != nil {
    return err
  }else{
    defer file.Close()
  }
  
  conf := &config{}
  
  // load our configuration
  if c, err := ioutil.ReadAll(file); err != nil {
    return fmt.Errorf("Could not read configuration: %v", err)
  }else if _, err := toml.Decode(string(c), conf); err != nil {
    return fmt.Errorf("Configuration is not valid: %v", err)
  }
  
  // initialize options
  if conf.Quiet != nil    { o.SetFlag(OptionsFlagQuiet,   *conf.Quiet) }
  if conf.Verbose != nil  { o.SetFlag(OptionsFlagVerbose, *conf.Verbose && !o.GetFlag(OptionsFlagQuiet)) }
  if conf.Debug != nil    { o.SetFlag(OptionsFlagDebug,   *conf.Debug   && !o.GetFlag(OptionsFlagQuiet)) }
  
  // initialize server config
  if conf.Server.Port != nil  { o.Server.Port = *conf.Server.Port }
  if conf.Server.Proxy != nil { o.Server.Proxy = *conf.Server.Proxy }
  if conf.Server.Root != nil  { o.Server.Root = *conf.Server.Root }
  
  // initialize JS config
  if conf.Javascript.Minify != nil { o.Javascript.Minify = *conf.Javascript.Minify }
  if conf.Javascript.Exclude != nil { o.Javascript.Exclude = append(o.Javascript.Exclude, *conf.Javascript.Exclude...) }
  
  // initialize CSS config
  if conf.Stylesheet.Minify != nil { o.Stylesheet.Minify = *conf.Stylesheet.Minify }
  if conf.Stylesheet.Exclude != nil { o.Stylesheet.Exclude = append(o.Stylesheet.Exclude, *conf.Stylesheet.Exclude...) }
  
  // initialize unmanaged config
  if conf.Unmanaged.Copy != nil { o.Unmanaged.Copy = *conf.Unmanaged.Copy }
  if conf.Unmanaged.Exclude != nil { o.Unmanaged.Exclude = append(o.Unmanaged.Exclude, *conf.Unmanaged.Exclude...) }
  
  // initialize routes
  if o.Routes == nil {
    o.Routes = make(map[string][]string)
  }
  
  // merge routes
  if conf.Routes != nil {
    if routes, err := mapToRoutes(conf.Routes); err != nil {
      return err
    }else{
      for k, v := range routes {
        o.Routes[k] = v
      }
    }
  }
  
  return nil
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
 * Get a flag
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

/**
 * Determine whether the specified resource should be excluded from compilation
 */
func (o *Options) ShouldExclude(resource string) bool {
  switch path.Ext(resource) {
    case ".scss", ".css":
      return shouldExclude(resource, o.Stylesheet.Exclude)
    case ".ejs", ".js":
      return shouldExclude(resource, o.Javascript.Exclude)
    default:
      return false
  }
}

/**
 * Determine whether the specified resource should be excluded from compilation
 */
func shouldExclude(resource string, patterns []string) bool {
  name := path.Base(resource)
  
  for _, p := range patterns {
    if match, err := path.Match(p, name); err != nil {
      log.Printf("ERROR: exclude resource shell pattern is invalid: '%s' %v; ignoring", p, err)
      return false
    }else if match {
      return true
    }
  }
  
  return false
}

/**
 * Convert a map of interfaces to a multimap of string -> []string.
 * Elements in the parameter map must be either string or []string.
 */
func mapToRoutes(m map[string]interface{}) (map[string][]string, error) {
  r := make(map[string][]string)
  
  for k, v := range m {
    switch c := v.(type) {
      case string:
        r[k] = []string{c}
      case []string:
        r[k] = c
      case []interface{}:
        if a, err := arrayToRoutes(c); err != nil {
          return nil, err
        }else{
          r[k] = a
        }
      default:
        return nil, fmt.Errorf("Type is not supported: %v", reflect.TypeOf(v))
    }
  }
  
  return r, nil
}

/**
 * Convert []interface{} to []string. The elements must actually be strings.
 */
func arrayToRoutes(a []interface{}) ([]string, error) {
  r := make([]string, len(a))
  
  for i, e := range a {
    switch c := e.(type) {
      case string:
        r[i] = c
      default:
        return nil, fmt.Errorf("Type is not supported: %v", reflect.TypeOf(e))
    }
  }
  
  return r, nil
}

