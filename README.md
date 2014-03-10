Webasm
======

The latest and best incarnation of Web Assembler.

About
-----

*Web Assembler* includes a bundled server that compiles and serves assets as they are requested. You can test your site directly from this server or you can run Web Assembler in tandem with whatever app server you use.

When you make requests to the Web Assembler server it determines whether request is for a resource that Web Assembler manages (such as SCSS, CSS, JS files) and, if so, it compiles and responds with that resource. Requests for resources that Web Assembler does not manage directly (such as HTML) are reverse-proxied to your app server.

Running Webasm
--------------

If you are only have static assets, you can start the Web Assembler server by running the following command.

	$ webasm -server

If you are using Web Assembler in tandem with an app server you need to provide the URL to the app server it will reverse proxy.

	$ webasm -server -proxy http://localhost:8080/

In either case Web Assembler looks for resources relative to the directory you run it in. If your app server maps any assets that Web Assembler manages to a different path than they exist in the filesystem, you can provide explicit route mappings.

	$ webasm -server -proxy http://localhost:8080/ -route /assets/css=/stylesheets

In this example, the URL `http://localhost:9090/assets/css/style.css` would be mapped to the file at `./stylesheets/style.css`.

Extended Javascript
-------------------

Extended Javascript (EJS) is a lightweight extension to Javascript that adds support for pre-compilation macros which are interpreted by Web Assember. Currently only one such macro is supported: `import`.

Macros have semantics similar to those used by the C preprocessor. Macros begin with the `#` character as the first character of a line, followed by an identifier, then by content specific to that macro.

### Using Macros

The `import` macro is illustrative, it is used like this.

	#import "another_file.js"
	
The following, however, will **not** be interpreted as a macro since the `#` does not begin on the first character of line of the line:

	    #import "another_file.js"

Within a macro, whitespace is ignored. So this is perfectly valid:

	# import
		"another_file.js"

Be aware that the EJS parser does not consider the Javascript context that macros are declared in. As a result, macros in multi-line comments **will** still be expanded, such as in the following:

	/*
	#import "another_file.js"
	*/

### #import

The `import` macro is used like this:

	#import "another_file.js"

When Web Assembler encounters an `import` macro in an EJS file it replaces the `import` statement itself with the entire contents of the file it references. Imported files are declared relative to the file that is importing it.

To import a file from a URL you can use the following. Web Assembler automatically detects that this is a URL and fetches it. Only `http` and `https` URLs are supported for the `import` macro.

	#import "http://ajax.googleapis.com/ajax/libs/jquery/1.11.0/jquery.min.js"

