Slang
======

A thing that builds websites.

About
-----

*Slang* includes a specialized built-in server that compiles and serves assets as they are requested. You can test your site directly from this server or you can run Slang in tandem with whatever app server you use.

When you make requests to the Slang server it determines whether that request is for a resource that Slang manages (such as SCSS, EJS, JS files) and, if so, it compiles and responds with that resource.

Requests for resources that Slang does not manage directly (such as HTML) are either reverse-proxied to your app server (if so configured) or copied from disk without modification.

Running Slang
--------------

If you are only have static assets, you can start the Slang server by running the following command.

	$ slang -server

If you are using Slang in tandem with an app server you need to provide the URL to the app server it will reverse proxy.

	$ slang -server -proxy http://localhost:8080/

In either case Slang looks for resources relative to the directory you run it in. If your app server maps any assets that Slang manages to a different path than they exist in the filesystem, you can provide explicit route mappings.

	$ slang -server -proxy http://localhost:8080/ -route /assets/css=/stylesheets

In this example, the URL `http://localhost:9090/assets/css/style.css` would be mapped to the file at `./stylesheets/style.css`.

Packaging Projects
------------------

The built-in Slang server compiles assets in memory but does not write them to disk. When you're ready to deploy your project you'll need to generate static versions of all your managed assets.

To package your project, simply point Slang at the root under which your assets are located.

	$ slang ./assets

Slang will traverse the directory and compile any supported assets it encounters. For example, a file named `assets/site.ejs` will be compiled by Slang and written to `assets/site.js`. Take care when naming your files, any file already existing at an output path will be overwritten.

What Gets Processed
-------------------

Slang will process all the file types it understands for you. Currently that includes SASS, EJS, and Javascript files.

Slang does not require you to maintain an explicit configuration file, it knows what to do based on some well-established conventions. As a result, *you must take care to follow these conventions* when naming your files or Slang will not work as expected.

### Output Formats

To determine what to do with a file, Slang looks at the file's extension(s). The extensions `.js` and `.css` are considered to be *output formats*. These extensions are considered to be final and complete; they are not processed at all.

* `file.css` → *no action* → `file.css`
* `file.js` → *no action* → `file.js`

### Input Formats

Extensions such as `.scss` and `.ejs` are compiled from their higher-level formats into their counterpart `.css` and `.js` output formats. For example:

* `file.scss` → *scss compiler* → `file.css`
* `file.ejs` → *ejs compiler* → `file.js`

The secondary extension `.min` is handled somewhat differently from the conventions you may be used to. The `.min` extension indicates that an input file *should be* minimized when it is processed, it **does not** mean that the contents of the file *already is* minimized.

* `file.min.scss` → *scss compiler*  → *minimizer* → `file.css`
* `file.min.css` → *minimizer* → `file.css`
* `file.min.ejs` → *ejs compiler*  → *minimizer* → `file.js`
* `file.min.js` → *minimizer* → `file.js`

This somewhat unusual handling may cause problems when you are using already-minimized Javascript library files. To work around this quirk, use the unminimized development version of libraries in your project and let Slang minimize them for you.

### Reverse Mapping

When running in server mode during development these conventions are reversed. In your HTML you should reference the *output format* version of your assets. Slang maps these back to their counterpart input format in your project. For example, when referencing the following stylesheet:

	<link rel="stylesheet" href="css/style.css" />

The Slang server will check for the following resources, in order. The first resource it encounteres will be compiled as appropriate and responded with:

1. `css/style.min.scss`
2. `css/style.scss`
3. `css/style.min.css`
4. `css/style.css`

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

When Slang encounters an `import` macro in an EJS file it replaces the `import` statement itself with the entire contents of the file it references. Imported files are declared relative to the file that is importing it.

To import a file from a URL you can use the following. Slang automatically detects that this is a URL and fetches it. Only `http` and `https` URLs are supported for the `import` macro.

	#import "http://ajax.googleapis.com/ajax/libs/jquery/1.11.0/jquery.min.js"

