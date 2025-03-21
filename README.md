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

Create a new Mold layout and render the view in an HTTP handler.

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
by creating a custom layout file and specifying it in the config for a new instance.

```go
config := mold.Config{
    Layout: "path/to/layout.html",
}
layout, err := mold.NewWithConfig(config)
```

### Views

Views are templates that generate the content that is inserted into the body of layouts.

Sections can also be defined within views, allowing content to be rendered in specific parts of the layout.
The `head` section in the default layout is a named section.

```html
{{define "scripts"}}
<script src="//unpkg.com/alpinejs" defer></script>
{{end}}
```

### Partials

Partials are reusable template snippets that allow you to break down complex views into smaller, manageable components.
Ideal for sharing common logic across multiple views and layouts.

```html
{{ partial "path/to/partial.html" }}
```

An optional second argument allows customizing the data passed to the partial.
By default, the view's data context is used.

```html
{{ partial "partials/user_session.html" .User }}
```

## What is wrong with Go templates?

Nothing! It is good at what it does.

However, Go templates, while simple and powerful, can be unfamiliar when dealing with multiple files.
Mold provides a more intuitive and familiar higher-level usage, without reinventing the wheel.

## License

MIT

## Sponsoring

You can support the author by donating on [Github Sponsors](https://github.com/sponsors/abiosoft)
or [Buy me a coffee](https://www.buymeacoffee.com/abiosoft).
