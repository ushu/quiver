# quiver

A library to load data from the [Quiver] app.

For details see for [Quiver format docs].

## Getting Started

This is a basic Go lib without external dependencies.

Tested with **Go 1.8**, but it should work on most setups.

### Installing

```
$ go get github.com/ushu/quiver
```

### Usage

```go
import (
    "github.com/ushu/quiver"
)
```

and then:

```go
// Load contents:
lib, _ := quiver.ReadLibrary("/path/to/Quiver.qvlibrary", true)

// Then use the loaded data tree:
for _, notebook := range lib.Notebooks {
    for _, note := notebook.Notes {
        // Print the title of the note
        fmt.Println(note.Title)
        
        for _, cell := note.cells {
            // Print the type of cell
            fmt.Println(note.Title)
            
            // ...
        }
    }
}
```

## Additional tooling

This library comes with two binaries:

* `cmd/quiver_to_json` is a small tool that allows loading a full library into a single JSON file
* `cmd/quiver_to_markdown` is a small tool output all the notes as a tree of Markdown files

You can install then right away with the `go` tool:

```sh
$ go install github.com/ushu/quiver/cmd/quiver_to_markdown
$ go install github.com/ushu/quiver/cmd/quiver_to_json
```

## Version & Contributing

This is **v0.2.0** of the tool, and is an **early release** by all standards.

Feel free to contribute anytime !

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details

## TODO

* [ ] Add support for creating a valid Quiver Library from code (this version is mostly for reading)
* [ ] Add some tests

[Quiver]: https://itunes.apple.com/app/id866773894
[Quiver format docs]: https://github.com/HappenApps/Quiver/wiki/Quiver-Data-Format
