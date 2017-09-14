/*
The quiver_to_markdown converts a Quiver library into a set of Markdown files.

It allows to backup your notes on Github or any other service that supports markdown.

Usage:

	# To load the Quiver library into a set of Markdown files
	$ quiver_to_markdown /path/to/Quiver.qvlibrary output_path
*/
package main

import (
	"fmt"
	"os"

	"errors"
	"path/filepath"

	"strings"

	"bufio"

	"github.com/ushu/quiver"
)

func main() {
	if len(os.Args) != 3 {
		fmt.Println(os.Args)
		fmt.Println("Usage: quiver_to_markdown QUIVER_LIBRARY OUTPUT_DIRECTORY")
		os.Exit(1)
	}

	// Read full library into memory
	library, err := quiver.ReadLibrary(os.Args[1], true)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	// output to the provided directory
	outPath := os.Args[2]
	err = writeLibrary(outPath, library)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func writeLibrary(outPath string, library *quiver.Library) error {
	err := EnsureDirectory(outPath)
	if err != nil {
		return err
	}

	for _, nb := range library.Notebooks {
		np := filepath.Join(outPath, CleanPathElement(nb.Name))
		err = writeNoteBook(np, nb)
		if err != nil {
			return err
		}
	}

	return nil
}

func writeNoteBook(np string, nb *quiver.Notebook) error {
	err := EnsureDirectory(np)
	if err != nil {
		return err
	}

	for _, note := range nb.Notes {
		err = writeNote(np, note)
		if err != nil {
			return err
		}
	}

	return nil
}

func writeNote(np string, note *quiver.Note) error {
	fn := CleanPathElement(note.Title)
	if len(fn) == 0 {
		return nil // skip
	}
	p := filepath.Join(np, fn+".md")

	// Write the note itself
	err := writeNoteMarkdown(p, note)
	if err != nil {
		return err
	}

	// has resources ?
	rp := filepath.Join(np, "resources")
	if len(note.Resources) > 0 {
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
	} else {
		os.RemoveAll(rp)
	}

	return nil
}

func writeNoteMarkdown(p string, note *quiver.Note) error {
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

		switch {
		case c.IsCode():
			_, err = fmt.Fprintf(out, "```%v\n%v\n```", c.Language, string(c.Data))
		case c.IsLatex():
			_, err = fmt.Fprintf(out, "```latex\n%v\n```", string(c.Data))
		case c.IsMarkdown():
			_, err = fmt.Fprintln(out, string(c.Data))
		case c.IsText():
			_, err = fmt.Fprintln(out, string(c.Data))
		}
		if err != nil {
			return err
		}
	}

	return nil
}

func writeResource(op string, r *quiver.NoteResource) error {
	rf, err := os.Create(op)
	if err != nil {
		return err
	}
	defer rf.Close()

	buf := bufio.NewWriter(rf)
	defer buf.Flush()

	_, err = fmt.Fprintln(buf, string(r.Data))
	if err != nil {
		return err
	}

	return nil
}

func EnsureDirectory(outPath string) error {
	stat, err := os.Stat(outPath)
	if err != nil {
		if !os.IsNotExist(err) {
			return err
		}
		// now we create the dir
		err = os.MkdirAll(outPath, 0755)
		if err != nil {
			msg := fmt.Sprintf("Could not create directory: %v", outPath)
			return errors.New(msg)
		}
	} else if !stat.IsDir() {
		msg := fmt.Sprintf("Shoud be a directory: %v", outPath)
		return errors.New(msg)
	}

	return nil
}

func CleanPathElement(p string) string {
	p = strings.Replace(p, "/", "|", -1)
	p = strings.Replace(p, ":", "-", -1)
	return strings.TrimSpace(p)
}
