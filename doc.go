// Package mold provides a flexible web page rendering engine that
// incorporates layouts and views for structured HTML output.
//
// The core component of this package is the [Engine], which
// manages the rendering process. The engine is configured using [Option].
//
// The [Engine] combines layouts and views to produce complete HTML pages.
// Layouts define the top-level structure, such as headers, footers, and
// navigation, while views provide the dynamic content for specific pages.
//
// [Engine] also supports partial templates, which can be included within both layouts and views.
//
// Example Usage:
//
//	dir := os.DirFS("web")
//
//	options := mold.With(
//	    mold.WithRoot("templates"),
//	    mold.WithLayout("layout.html"),
//	}
//
//	engine, err := mold.New(dir, options)
//	if err != nil {
//	    // handle error
//	}
//
//	data := map[string]any{
//	    "Title":   "My Page",
//	    "Content": "Welcome to my website!",
//	}
//
//	handler := func(w http.ResponseWriter, r *http.Request) {
//	   engine.Render(w, "view.html", data)
//	}
package mold
