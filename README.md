# Mold

Mold builds on [Go templates](https://pkg.go.dev/text/template) to provide a simple and familiar API for rendering web pages.

## Getting Started

### 1. Create a view file

Create an HTML file named `index.html`.

```html
{{define "head"}}
<link rel="stylesheet" href="https://cdn.simplecss.org/simple.min.css">
{{end}}

<h1>Hello from a <a href="//github.com/abiosoft/mold">Mold</a> template</h1>
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

Check the [examples](https://github.com/abiosoft/mold/tree/main/examples) directory for more.

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
    {{ render "head" }}
</head>
<body>
    {{ render }}
</body>
</html>
```
The [default](https://github.com/abiosoft/mold/blob/main/layout.html) layout can be overriden
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

### Sections

Sections allows content to be rendered in specific parts of the layout.
They are defined within views with the `define` block.

The default template includes the `head` section.

```html
{{define "scripts"}}
<script src="//unpkg.com/alpinejs" defer></script>
{{end}}
```

### Partials

Partials are reusable template snippets that allow you to break down complex views into smaller, manageable components.
They are supported in both views and layouts with the `partial` function.

Partials are ideal sharing common logic across multiple views and layouts.

```html
{{ partial "path/to/partial.html" }}
```

An optional second argument allows customizing the data passed to the partial.
By default, the view's data context is used.

```html
{{ partial "partials/user_session.html" .User }}
```

## Why not standard Go templates?

Go templates, while simple and powerful, can be unfamiliar when dealing with multiple template files.

Mold provides an intuitive and conventional higher-level usage of Go templates for dealing with multiple template files.

## License

MIT

## Sponsoring

You can support the author by donating on [Github Sponsors](https://github.com/sponsors/abiosoft)
or [Buy me a coffee](https://www.buymeacoffee.com/abiosoft).
