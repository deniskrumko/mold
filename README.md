# Mold

Mold builds on [Go templates](https://pkg.go.dev/text/template) to provide a simple and familiar API for rendering web pages.

## Getting Started

### 1. Create a template file

Create an HTML file named `index.html`.

```html
{{define "head"}}
<link rel="stylesheet" href="https://cdn.simplecss.org/simple.min.css">
{{end}}

<h1>Hello from a <a href="//github.com/abiosoft/mold">Mold</a> template</h1>
```

### 2. Render

Create and render the layout in an HTTP handler.

```go
//go:embed index.html
var dir embed.FS

var layout, _ = mold.New(dir)

func handle(w http.ResponseWriter, r *http.Request){
    layout.Render(w, "index.html", nil)
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
To render a specific named section, pass the section's name as an argument.

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
by creating a custom layout file and specifying it in the config for a new instance.

```go
config := mold.Config{
    Layout: "path/to/layout.html",
}
layout, err := mold.NewWithConfig(config)
```

### Views

Views are templates that generate the content that is inserted into the body of layouts.
Views support named sections, allowing content to be rendered in specific parts of the layout.

The `head` section is in the default layout merely as a convention, a section can be given any name.

```html
{{define "scripts"}}
<script src="//unpkg.com/alpinejs" defer></script>
{{end}}
```

### Partials

Partials are reusable template snippets that allow you to break down complex views into smaller, manageable components.
They can be used to encapsulate and reuse common logic across multiple views and layouts.

```html
{{ partial "path/to/partial.html" }}
```

An optional second argument allows customizing the data passed to the partial.
By default, the view's data context is used.

```html
{{ partial "partials/user_session.html" .User }}
```

## Why?

Go templates, while simple and powerful, can be unfamiliar when dealing with multiple files.
Mold provides a more intuitive and familiar higher-level usage, without reinventing the wheel.

## License

MIT
