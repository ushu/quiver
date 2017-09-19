# quiver_to_markdown

Converts a [Quiver] library to a set of markdown files.

### Installing

```sh
$ go get -u github.com/ushu/quiver/cmd/quiver_to_markdown
```

### Usage

```sh
# Convert an existing Quiver library to Markdown
$ quiver_to_markdown /path/to/Quiver.qvlibrary /output/path

# Print version
$ quiver_to_markdown -v
```

Then all the notes are availation in `/output/path` as Markdown files.

## License

This project is licensed under the MIT License - see the [LICENSE](../../LICENSE) file for details

## TODO

* [ ] Add support for note links

[Quiver]: https://itunes.apple.com/app/id866773894
