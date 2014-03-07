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

