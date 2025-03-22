/*
Package mold provides a web page rendering engine that combines layouts and views for HTML output.

The core component of this package is the [Engine], which manages the rendering process.

	dir := os.DirFS("web")

	engine, err := mold.New(dir)
	if err != nil {
	    // handle error
	}

	data := map[string]any{
	    "Title":   "My Page",
	    "Content": "Welcome to my website!",
	}

	handler := func(w http.ResponseWriter, r *http.Request) {
	   engine.Render(w, "view.html", data)
	}

Layouts define the top-level structure, such as headers, footers, and
navigation.

Inside a layout, calling "render" without an argument inserts the view's content into the layout's body.
To render a specific section, pass the section's name as an argument.

	<!DOCTYPE html>
	<html>

	<head>
	    {{render "head"}}
	</head>

	<body>
	    {{render}}
	</body>

	</html>

Views are templates that generate the content that is inserted into the body of layouts.
Typically what you would put in the "<body>" tag of an HTML page.

	<h3>Hello from Mold :)</h3>

The path to the view file is passed to the rendering engine to produce HTML output.

	engine.Render(w, "path/to/view.html", nil)

Sections allow content to be rendered in specific parts of the layout.
They are defined within views with a "define" block.

The default layout is able to render HTML content within the "<head>" tag by utilising the "head" section.

	{{define "scripts"}}
	<script src="//unpkg.com/alpinejs" defer></script>
	{{end}}

Partials are reusable template snippets that allow you to break down complex views into smaller,
manageable components. They are supported in both views and layouts with the "partial" function.

Partials are ideal for sharing common logic across multiple views and layouts.

	{{partial "path/to/partial.html"}}

An optional second argument allows customizing the data passed to the partial.
By default, the view's data context is used.

	{{partial "partials/user_session.html" .User}}
*/
package mold
