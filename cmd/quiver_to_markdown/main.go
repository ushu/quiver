/*
The quiver_to_markdown converts a Quiver library into a set of Markdown files.

It allows to backup your notes on Github or any other service that supports markdown.

Usage:

	$ quiver_to_markdown /path/to/Quiver.qvlibrary output_path
*/
package main

import (
	"fmt"
	"os"

	"path/filepath"

	"strings"

	"bufio"

	"io/ioutil"

	"regexp"

	"flag"

	"github.com/ushu/quiver"
	"path"
	"github.com/pkg/errors"
)

// PathElementReplacer
//
// Theorically, all characters are acceptable (https://en.wikipedia.org/wiki/HFS_Plus) in path elements,
// in practise they can cause strange issues in Finder...
var PathElementReplacer = strings.NewReplacer(
	"/", "|",
	":", "-",
)

// Rewrite language name from Quiver Code Cell conventions to Github Markdown ones
var languageEquivalents = map[string]string{
	"c_cpp": "c++",
}

// Index of notes by UUID -> new path
type NotesIndex map[string]string

var noteURLRegexp = regexp.MustCompile(`quiver-note-url/([0-9A-F]{8}-[0-9A-F]{4}-[0-9A-F]{4}-[0-9A-F]{4}-[0-9A-F]{12})`)

var flagVersion bool

func init() {
	flag.BoolVar(&flagVersion, "v", false, "print version")
}

func main() {
	flag.Parse()

	if flagVersion {
		fmt.Printf("v%v\n", quiver.Version)
		os.Exit(0)
	}

	if flag.NArg() != 2 {
		fmt.Println("Usage: quiver_to_markdown [-v] QUIVER_LIBRARY OUTPUT_DIRECTORY")
		flag.PrintDefaults()
		os.Exit(1)
	}

	// Read full library into memory
	inPath := flag.Arg(0)
	library, err := quiver.ReadLibrary(inPath, true)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	outPath := flag.Arg(1)

	var index NotesIndex = make(map[string]string)
	err = library.WalkNotebooksHierarchy(func(nb *quiver.Notebook, parents []*quiver.Notebook) error {
		// build the notebook path
		pe := make([]string, 0)
		pe = append(pe, outPath)
		for _, p := range parents {
			pe = append(pe, CleanPathElement(p.Name))
		}
		// then rhe notebook and the file name
		pe = append(pe, CleanPathElement(nb.Name))
		nbp := filepath.Join(pe...)

		for _, n := range nb.Notes {
			if _, ok := index[n.UUID]; ok {
				return errors.Errorf("There found two notes with UUID \"%s\", aborting...", n.UUID)
			}
			index[n.UUID] = filepath.Join(nbp, CleanPathElement(n.Title)+".md")
		}

		return nil
	})
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	// output to the provided directory
	err = writeLibrary(outPath, library, index)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	fmt.Printf("Done converting %q to %q\n", inPath, outPath)
}

func writeLibrary(outPath string, library *quiver.Library, index NotesIndex) error {
	err := ResetDirectory(outPath, false)
	if err != nil {
		return err
	}

	return library.WalkNotebooksHierarchy(func(nb *quiver.Notebook, parents []*quiver.Notebook) error {
		// build the notebook path
		pe := make([]string, 0)
		pe = append(pe, outPath)
		for _, p := range parents {
			pe = append(pe, CleanPathElement(p.Name))
		}
		// then rhe notebook and the file name
		pe = append(pe, CleanPathElement(nb.Name))
		nbp := filepath.Join(pe...)

		return writeNoteBook(nbp, nb, index)
	})
}

func writeNoteBook(np string, nb *quiver.Notebook, index NotesIndex) error {
	err := ResetDirectory(np, true)
	if err != nil {
		return err
	}

	for _, note := range nb.Notes {
		p := index[note.UUID]
		err := writeNote(p, note, index)
		if err != nil {
			return err
		}
	}

	return nil
}

func writeNote(p string, note *quiver.Note, index NotesIndex) error {
	// Write the note itself
	err := writeNoteMarkdown(p, note, index)
	if err != nil {
		return err
	}

	// has resources ?
	if len(note.Resources) > 0 {
		rp := filepath.Join(path.Dir(p), "resources")
		err = EnsureDirectory(rp)
		if err != nil {
			return err
		}

		for _, r := range note.Resources {
			op := filepath.Join(rp, r.Name)
			err = writeResource(op, r)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func writeNoteMarkdown(p string, note *quiver.Note, index NotesIndex) error {
	f, err := os.Create(p)
	if err != nil {
		return err
	}
	defer f.Close()

	// the output stream
	out := bufio.NewWriter(f)
	defer out.Flush()

	for i, c := range note.Cells {
		if i != 0 {
			_, err = fmt.Fprintln(out)
			if err != nil {
				return err
			}
		}

		// content to write: we replace all the data links to relative links
		data := string(c.Data)
		data = strings.Replace(data, "quiver-image-url/", "resources/", -1)

		if index != nil {
			data = noteURLRegexp.ReplaceAllStringFunc(data, func(m string) string {
				UUID := strings.TrimPrefix(m, "quiver-note-url/")
				return "../" + index[UUID]
			})
		}

		switch {
		case c.IsCode():
			// load language and (optionally) converts it to its Github Markdown equivalent
			l := c.Language
			if eq, ok := languageEquivalents[l]; ok {
				l = eq
			}
			_, err = fmt.Fprintf(out, "```%v\n%v\n```", l, data)
		case c.IsLatex():
			_, err = fmt.Fprintf(out, "```latex\n%v\n```", data)
		case c.IsMarkdown():
			_, err = fmt.Fprintln(out, data)
		case c.IsText():
			_, err = fmt.Fprintln(out, data)
		case c.IsDiagram():
			tool := "Sequence diagram, see https://bramp.github.io/js-sequence-diagrams"
			if c.DiagramType == "flow" {
				tool = "Flowchart diagram, see http://flowchart.js.org"
			}
			_, err = fmt.Fprintf(out, "```javascript\n// %v\n%v\n```", tool, data)
		}
		if err != nil {
			return err
		}
	}

	return nil
}

func writeResource(op string, r *quiver.NoteResource) error {
	return ioutil.WriteFile(op, r.Data, 0755)
}

func ResetDirectory(outPath string, removeRoot bool) error {
	stat, err := os.Stat(outPath)
	if err != nil {
		if !os.IsNotExist(err) {
			return err
		}
	} else if stat.IsDir() {
		if removeRoot {
			err = os.RemoveAll(outPath)
			if err != nil {
				return err
			}
		} else {
			// Remove all subdirs, except hidden ones, but keep the root dir (outPath)
			// this is mainly to allow storing the results in a git repo
			files, err := ioutil.ReadDir(outPath)
			if err != nil {
				return err
			}
			for _, f := range files {
				// skip standard files and hidden dirs
				if !f.IsDir() || strings.HasPrefix(f.Name(), ".") {
					continue
				}
				os.RemoveAll(filepath.Join(outPath, f.Name()))
			}
		}
	}

	return EnsureDirectory(outPath)
}

func EnsureDirectory(outPath string) error {
	err := os.MkdirAll(outPath, 0755)
	if err != nil && !os.IsExist(err) {
		return err
	}
	return nil
}

func CleanPathElement(p string) string {
	p = strings.TrimSpace(p)
	p = PathElementReplacer.Replace(p)
	return p
}
