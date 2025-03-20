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

## Other Features

### Layouts

A custom layout can be specified to override the [default](https://github.com/abiosoft/mold/blob/main/layout.html).

`render` without arguments renders the entire template body.
Alternatively, you can specify a named section to render only that portion.

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

An instance can be created with config specifying the path to the layout file.

```go
config := mold.Config{
    Layout: "path/to/layout.html",
}
layout, err := mold.NewWithConfig(config)
```

### Partials

Reusable template snippets can be rendered within templates with `partial`.

```html
{{ partial "path/to/partial.html" }}
```

## Why?

Go templates, while simple and powerful, can be unfamiliar when dealing with multiple files.
Mold provides a more intuitive and familiar higher-level usage, without reinventing the wheel.

## License

MIT
