# Mold

Fork of https://github.com/abiosoft/mold to use different Go version.

Mold builds on [Go templates](https://pkg.go.dev/text/template) to provide a simple and familiar API for rendering web pages.

## Features

Mold offers the following, making it an ideal choice for Go projects.

- **Lightweight**: uses only the Go standard library with no external dependencies.
- **Efficient**: utilises Go's in-built template parser under the hood.
- **Capable**: supports all capabilities of Go templates.
- **Familiar**: employs the well-known [concepts](#concepts) of Layouts, Views and Partials.

## Getting Started

### 1. Create a view file

Create an HTML file named `index.html`.

```html
{{define "head"}}
<link rel="stylesheet" href="https://cdn.simplecss.org/simple.min.css">
{{end}}

<h1>Hello from a <a href="//github.com/deniskrumko/mold">Mold</a> template</h1>
```

### 2. Render

Create a new instance and render the view in an HTTP handler.

```go
//go:embed index.html
var dir embed.FS

var engine, _ = mold.New(dir)

func handle(w http.ResponseWriter, r *http.Request){
    engine.Render(w, "index.html", nil)
}
```

### Examples

Check the [examples](https://github.com/deniskrumko/mold/tree/main/examples) directory for more.

### Documentation

Go package documentation is available at https://pkg.go.dev/github.com/deniskrumko/mold

## Concepts

### Layouts

Layouts provide the overall structure for your web pages.
They define the common elements that are shared across multiple views,
such as headers, footers, navigation menus, stylesheets e.t.c.

Inside a layout, calling `render` without an argument inserts the view's content into the layout's body.
To render a specific section, pass the section's name as an argument.

```html
<!DOCTYPE html>
<html>
<head>
    {{render "head"}}
</head>
<body>
    {{render}}
</body>
</html>
```
The [default layout](https://github.com/deniskrumko/mold/blob/main/layout.html) can be overridden
by creating a custom layout file and specifying it as an option for a new instance.

```go
option := mold.WithLayout("path/to/layout.html")
engine, err := mold.New(fs, option)
```

### Views

Views are templates that generate the content that is inserted into the body of layouts.
Typically what you would put in the `<body>` of an HTML page.

```html
<h3>Hello from Mold :)</h3>
```

The path to the view file is passed to the rendering engine to produce HTML output.

```go
engine.Render(w, "path/to/view.html", nil)
```

### Partials

Partials are reusable template snippets that allow you to break down complex views into smaller, manageable components.
They are supported in both views and layouts with the `partial` function.

Partials are ideal for sharing common logic across multiple views and layouts.

```html
{{partial "path/to/partial.html"}}
```

An optional second argument allows customizing the data passed to the partial.
By default, the view's data context is used.

```html
{{partial "partials/user_session.html" .User}}
```

### Sections

Sections allow content to be rendered in specific parts of the layout.
They are defined within views with a `define` block.

The [default layout](https://github.com/deniskrumko/mold/blob/main/layout.html) is able to render HTML content within
the `<head>` tag by utilising the `head` section.

```html
{{define "scripts"}}
<script src="//unpkg.com/alpinejs" defer></script>
{{end}}
```


## Why not standard Go templates?

Go templates, while simple and powerful, can feel unfamiliar when dealing with multiple template files.

Mold addresses this by providing an intuitive framework for structuring web applications with Go templates.

## License

MIT

## Sponsoring

You can support the author by donating on [Github Sponsors](https://github.com/sponsors/abiosoft)
or [Buy me a coffee](https://www.buymeacoffee.com/abiosoft).
