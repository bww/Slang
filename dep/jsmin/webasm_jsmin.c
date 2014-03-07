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
// Based on the Ruby implementation, found here:
// https://github.com/rf-/jsminc
// 
// Distributed under the license:
// https://github.com/rf-/jsminc/blob/master/LICENSE
// 

#include "webasm_jsmin.h"

#include <string.h>
#include <stdlib.h>

/*
  isAlphanum -- return true if the character is a letter, digit, underscore,
  dollar sign, or non-ASCII character.
 */
static int isAlphanum(char c) {
  return ((c >= 'a' && c <= 'z') || (c >= '0' && c <= '9') ||
          (c >= 'A' && c <= 'Z') || c == '_' || c == '$' || c == '\\' ||
           c > 126);
}

/*
  get -- return the next character from the string. Watch out for lookahead. If
  the character is a control character, translate it to a space or
  linefeed.
 */
static char get(jsmin_context *s) {
  char c = s->lookahead;
  s->lookahead = '\0';
  if (c == '\0') {
    c = *(s->in)++;
  }
  if (c >= ' ' || c == '\n' || c == '\0') {
    return c;
  }
  if (c == '\r') {
    return '\n';
  }
  return ' ';
}

static void write_char(jsmin_context *s, const char c) {
  *(s->out) = c;
	s->out++;
}

/* peek -- get the next character without getting it. */
static char peek(jsmin_context *s) {
  s->lookahead = get(s);
  return s->lookahead;
}

/*
  next -- get the next character, excluding comments. peek() is used to see
  if a '/' is followed by a '/' or '*'.
 */
static char next(jsmin_context *s) {
  char c = get(s);
  if  (c == '/') {
    switch (peek(s)) {
    case '/':
      for (;;) {
        c = get(s);
        if (c <= '\n') {
          return c;
        }
      }
    case '*':
      get(s);
      for (;;) {
        switch (get(s)) {
        case '*':
          if (peek(s) == '/') {
            get(s);
            return ' ';
          }
          break;
        case '\0':
          abort();
          //rb_raise(rb_eStandardError, "unterminated comment");
        }
      }
    default:
      return c;
    }
  }
  return c;
}


/*
  action -- do something! What you do is determined by the argument:
  1   Output A. Copy B to A. Get the next B.
  2   Copy B to A. Get the next B. (Delete A).
  3   Get the next B. (Delete B).
  action treats a string as a single character. Wow!
  action recognizes a regular expression if it is preceded by ( or , or =.
 */
static void action(jsmin_context *s, int d) {
  switch (d) {
  case 1:
    write_char(s, s->A);
  case 2:
    s->A = s->B;
    if (s->A == '\'' || s->A == '"' || s->A == '`') {
      for (;;) {
        write_char(s, s->A);
        s->A = get(s);
        if (s->A == s->B) {
          break;
        }
        if (s->A == '\\') {
          write_char(s, s->A);
          s->A = get(s);
        }
        if (s->A == '\0') {
          abort();
          //rb_raise(rb_eStandardError, "unterminated string literal");
        }
      }
    }
  case 3:
    s->B = next(s);
    if (s->B == '/' && (s->A == '(' || s->A == ',' || s->A == '=' ||
        s->A == ':' || s->A == '[' || s->A == '!' ||
        s->A == '&' || s->A == '|' || s->A == '?' ||
        s->A == '{' || s->A == '}' || s->A == ';' ||
        s->A == '\n')) {
      write_char(s, s->A);
      write_char(s, s->B);
      for (;;) {
        s->A = get(s);
        if (s->A == '[') {
          for (;;) {
            write_char(s, s->A);
            s->A = get(s);
            if (s->A == ']') {
              break;
            }
            if (s->A == '\\') {
              write_char(s, s->A);
              s->A = get(s);
            }
            if (s->A == '\0') {
              abort();
              //rb_raise(rb_eStandardError, "unterminated set in regex literal");
            }
          }
        } else if (s->A == '/') {
          break;
        } else if (s->A =='\\') {
          write_char(s, s->A);
          s->A = get(s);
        }
        if (s->A == '\0') {
          abort();
          //rb_raise(rb_eStandardError, "unterminated regex literal");
        }
        write_char(s, s->A);
      }
      s->B = next(s);
    }
  }
}


/*
  jsmin -- Copy the input to the output, deleting the characters which are
  insignificant to JavaScript. Comments will be removed. Tabs will be
  replaced with spaces. Carriage returns will be replaced with linefeeds.
  Most spaces and linefeeds will be removed.
 */
static void jsmin(jsmin_context *s) {
  s->A = '\n';
	s->lookahead = '\0';
	
  action(s, 3);
  while (s->A != '\0') {
    switch (s->A) {
    case ' ':
      if (isAlphanum(s->B)) {
        action(s, 1);
      } else {
        action(s, 2);
      }
      break;
    case '\n':
      switch (s->B) {
      case '{':
      case '[':
      case '(':
      case '+':
      case '-':
        action(s, 1);
        break;
      case ' ':
        action(s, 3);
        break;
      default:
        if (isAlphanum(s->B)) {
          action(s, 1);
        } else {
          action(s, 2);
        }
      }
      break;
    default:
      switch (s->B) {
      case ' ':
        if (isAlphanum(s->A)) {
          action(s, 1);
          break;
        }
        action(s, 3);
        break;
      case '\n':
        switch (s->A) {
        case '}':
        case ']':
        case ')':
        case '+':
        case '-':
        case '"':
        case '\'':
        case '`':
          action(s, 1);
          break;
        default:
          if (isAlphanum(s->A)) {
            action(s, 1);
          } else {
            action(s, 3);
          }
        }
        break;
      default:
        action(s, 1);
        break;
      }
    }
  }
  write_char(s, '\0');
}

/*
 * Minify source
 */
char * jsmin_minify(const char *source) {
  jsmin_context s;
  
  char *minified;
  if((minified = malloc(strlen(source) + 1)) == NULL){
    return NULL;
  }
  
  s.in  = source;
  s.out = minified;
  jsmin(&s);
  
  return minified;
}

