Slang
======

Slang is a web server for web developers. It deals with managing, compiling, and transforming resources and coordinating between local and remote files. You just edit and refresh.

About
-----

Slang is a specialized web server that compiles and serves static assets (like SCSS, Javascript, etc) as they are requested. It can act as a reverse proxy to another server and allow you to view a project as a combination of local files on your computer (so you can easily edit them) and resources on a remote server (which may be out of your control or a hassle to get up and running locally).

You can use Slang to do some handy things:

* Test a static site locally and have your SCSS, CSS, JS, etc, automatically compiled when you refresh,
* Test google.com, but with some CSS you're editing locally,
* Test HTML generate by whatever app server you use but with SCSS you need updated for each request when you refresh.

When you make requests to the Slang server it determines whether that request is for a resource that Slang manages (such as SCSS, CSS, EJS, JS files) and if so, it compiles and responds with that resource. Requests for resources that Slang does not manage directly (such as HTML) are either reverse-proxied to another server (if so configured) or copied from disk without modification.

Currently Slang can manage these file types, with more to come:

* **SCSS** can be compiled and minified
* **CSS** can be minified
* **Extended Javascript** (see below) can be compiled an minified
* **Javascript** can be minified

Running Slang
--------------

If you are only have static assets, you can start the Slang server by running the following command.

	$ slang run

If you are using Slang in tandem with another server (such as an app server that is providing the HTML) you need to provide the URL that it will reverse-proxy.

	$ slang run -proxy http://localhost:8080/

In either case Slang looks for resources relative to the directory you provide as an argument to `run`, or in the directory you run it in if you don't specify one.


Running Slang: Slightly more advanced edition
---------------------------------------------

### Routes

If your app server maps any assets that Slang manages to a path other than where they exist in the filesystem, you can provide explicit route mappings.

	$ slang run -proxy http://localhost:8080/ -route /assets/css=/stylesheets

In this example, the URL `http://localhost:9090/assets/css/style.css` would be mapped to the file at `./stylesheets/style.css`.

### Using a Config File

Routes, along with most other things, can also be configured in a `slang.conf` to save you some typing every time you start Slang up. To generate a default configuration file that you can customize, use the following command.

	$ slang init

Slang will create a default configuration file called `slang.conf` in the current directory. When Slang starts up it checks for a `slang.conf` file in the current directory and, if it finds one, loads settings from it. You can override settings in your configuration on the command line.


Packaging Your Project
----------------------

The Slang server compiles assets in memory and serves them up but it does not write them to disk. When you're ready to deploy your project you'll need to generate static versions of all your managed assets.

To package your project use the `build` command and point Slang at the root under which your files are located. This may be the same directory you run the Slang server in.

	$ slang build -output ./ship ./assets

Slang will traverse the directory `./assets`, compile any supported assets it encounters, and write the result to a corresponding location under `./ship`. For example, a file named `assets/css/site.scss` will be compiled by Slang and written to `ship/css/site.css`.


What Gets Processed
-------------------

Slang will process all the file types it understands for you. Currently that includes SASS, CSS, EJS, and Javascript files.

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

The Slang server will check for the following resources, in order. The first resource it encounters will be compiled as appropriate and responded with:

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

