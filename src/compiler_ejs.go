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
	"path"
  "regexp"
  "strconv"
  "io/ioutil"
	"unicode/utf8"
)

import (
  "net/http"
)

/**
 * An "extended Javascript" (EJS) compiler
 */
type EJSCompiler struct {
  // ...
}

/**
 * Compile EJS
 */
func (c EJSCompiler) Compile(context Context, inpath, outpath string, input io.Reader, output io.Writer) error {
  var source []byte
  var err error
  
  if source, err = ioutil.ReadAll(input); err != nil {
    return err
  }
  
  scanner := NewScanner(string(source))
  outer:
  for {
    
    var toks []Token
    if toks, err = scanner.Token(); err != nil {
      return err
    }
    
    for _, tok := range toks {
      switch tok.Type {
        
        case tokenTypeEOF:
          break outer
          
        case tokenTypeVerbatim:
          if _, err := output.Write([]byte(tok.Text)); err != nil {
            return err
          }
        
        case tokenTypeImport:
          if err := c.emitImport(context, inpath, outpath, output, tok.Text); err != nil {
            return err
          }
        
      }
      
    }
    
  }
  
  return nil
}

/**
 * Emit an import
 */
func (c EJSCompiler) emitImport(context Context, inpath, outpath string, output io.Writer, resource string) error {
  if m, err := regexp.MatchString("^https?://", resource); err != nil {
    return err
  }else if m {
    return c.emitImportURL(context, inpath, outpath, output, resource)
  }else{
    return c.emitImportFile(context, inpath, outpath, output, resource)
  }
}

/**
 * Emit an import
 */
func (c EJSCompiler) emitImportURL(context Context, inpath, outpath string, output io.Writer, resource string) error {
  
  resp, err := http.Get(resource)
  if err != nil {
    return err
  }
  
  defer resp.Body.Close()
  if _, err := io.Copy(output, resp.Body); err != nil {
    return err
  }
  
  return nil
}

/**
 * Emit an import
 */
func (c EJSCompiler) emitImportFile(context Context, inpath, outpath string, output io.Writer, resource string) error {
  base := path.Dir(inpath)
  if file, err := os.Open(path.Join(base, resource)); err != nil {
    return err
  }else if _, err := io.Copy(output, file); err != nil {
    return err
  }else{
    return nil
  }
}

/**
 * Token types
 */
const (
  tokenTypeEOF          =  0
  tokenTypeVerbatim     =  1
  tokenTypeImport       =  2
  tokenTypeError        = -1
)

const eof = -1
const delimiter = '#'

/**
 * A token
 */
type Token struct {
  Type  int
  Text  string
}

/**
 * EJS scanner
 */
type Scanner struct {
  source    string
  length    int
  index     int
  width     int
}

/**
 * Create a scanner
 */
func NewScanner(source string) *Scanner {
  return &Scanner{source, len(source), 0, 0}
}

/**
 * Product a token
 */
func (s *Scanner) Token() ([]Token, error) {
  start := s.index
  
  for {
    r := s.next()
    switch r {
      
      case eof:
        if s.index - start > 0 {
          return []Token{ Token{tokenTypeVerbatim, s.source[start:s.index]}, Token{tokenTypeEOF, "EOF"} }, nil
        }else{
          return []Token{ Token{tokenTypeEOF, "EOF"} }, nil
        }
        
      case delimiter:
        n := s.index - 1
        if t, err := s.directiveToken(); err != nil {
          return nil, err
        }else if n - start > 0 {
          return append([]Token{ Token{tokenTypeVerbatim, s.source[start:n]} }, t...), nil
        }else{
          return t, nil
        }
      
    }
  }
  
  return nil, fmt.Errorf("Unexpected end of input")
}

/**
 * Produce a directive token
 */
func (s *Scanner) directiveToken() ([]Token, error) {
  var id string
  
  if id = s.scanIdentifier(); len(id) < 1 {
    return nil, fmt.Errorf("Expected identifier after delimiter start '#'")
  }
  
  switch id {
    case "import":
      return s.importToken()
    default:
      return nil, fmt.Errorf("No such directive '%s'", id)
  }
  
}

/**
 * Produce an import directive token
 */
func (s *Scanner) importToken() ([]Token, error) {
  if resource, err := s.scanQuotedString(); err != nil {
    return nil, err
  }else{
    return []Token{ Token{tokenTypeImport, resource} }, nil
  }
}

/**
 * Increment the cursor position
 */
func (s *Scanner) inc(width int) int {
	s.width  = width
	s.index += width
	return s.index
}

/**
 * Produce the next rune from input without incrementing our position
 */
func (s *Scanner) peek() (rune, int) {
	if int(s.index) >= s.length { return eof, 0 }
	return utf8.DecodeRuneInString(s.source[s.index:])
}

/**
 * Consume the next rune from input
 */
func (s *Scanner) next() rune {
	r, w := s.peek()
	s.width  = w
	s.index += w
	return r
}

/**
 * Consume the next identifier and return it
 */
func (s *Scanner) scanIdentifier() string {
  var id string
  s.skipWhite()
  
  for {
    r := s.next()
    if (r >= 'A' && r <= 'Z') || (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '_'  {
      id += string(r)
    }else{
      return id
    }
  }
  
}

/**
 * Scan the next escape sequence
 */
func (s *Scanner) scanEscape(quote rune) (rune, error) {
	ch := s.next() // read character after '/'
	switch ch {
    case 'a':
      return '\a', nil
    case 'b':
      return '\b', nil
    case 'f':
      return '\f', nil
    case 'n':
      return '\n', nil
    case 'r':
      return '\r', nil
    case 't':
      return '\t', nil
    case 'v':
      return '\v', nil
    case '\\':
      return '\\', nil
    case quote:
      return quote, nil
    case '0', '1', '2', '3', '4', '5', '6', '7':
      if v, err := s.scanDecimal(ch, 8, 3); err != nil {
        return 0, err
      }else{
        return rune(v), nil
      }
    case 'x':
      if v, err := s.scanDecimal(ch, 16, 2); err != nil {
        return 0, err
      }else{
        return rune(v), nil
      }
    case 'u':
      if v, err := s.scanDecimal(ch, 16, 4); err != nil {
        return 0, err
      }else{
        return rune(v), nil
      }
    case 'U':
      if v, err := s.scanDecimal(ch, 16, 8); err != nil {
        return 0, err
      }else{
        return rune(v), nil
      }
    default:
      return 0, fmt.Errorf("Invalid escape sequence")
	}
	return ch, nil
}

/**
 * Scan decimal number
 */
func (s *Scanner) scanDecimal(ch rune, base, n int) (int64, error) {
  var num string
  
	for n > 0 && digitValue(ch) < base {
		num += string(ch)
		ch = s.next(); n--
	}
	
	if n > 0 {
		return 0, fmt.Errorf("illegal char escape")
	}
	
	return strconv.ParseInt(num, base, 64)
}

/**
 * Scan the next quoted string
 */
func (s *Scanner) scanQuotedString() (string, error) {
  ch := s.next()
  switch ch {
    case '"':
      return s.scanString('"')
    default:
      return "", fmt.Errorf("Invalid quote character")
  }
}

/**
 * Scan the next quoted string
 */
func (s *Scanner) scanString(quote rune) (string, error) {
  var str string
	for {
	  ch := s.next() // read character after quote
	  
	  if ch == quote {
	    return str, nil
		}else if ch == '\n' || ch < 0 {
			return "", fmt.Errorf("String is not terminated")
		}
		
		if ch == '\\' {
			if ch, err := s.scanEscape(quote); err != nil {
			  return "", err
			}else{
			  str += string(ch)
			}
		}else{
      str += string(ch)
		}
		
	}
	return str, nil
}

/**
 * Consume runes until a non-whitespace rune is encountered and return it
 */
func (s *Scanner) skipWhite() {
  for {
    r, w := s.peek()
    if r <= ' ' {
      s.inc(w)
    }else{
      return
    }
  }
}

/**
 * Consume runes until a non-whitespace rune is encountered and return it
 */
func (s *Scanner) nextSkipWhite() rune {
  s.skipWhite()
  return s.next()
}

/**
 * Digit value
 */
func digitValue(ch rune) int {
	switch {
    case '0' <= ch && ch <= '9':
      return int(ch - '0')
    case 'a' <= ch && ch <= 'f':
      return int(ch - 'a' + 10)
    case 'A' <= ch && ch <= 'F':
      return int(ch - 'A' + 10)
	}
	return 16 // larger than any legal digit val
}

