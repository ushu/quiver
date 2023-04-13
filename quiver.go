/*
Package quiver implements a simple library for parsing quiver libraries, notebooks and notes.

The most straightforward way to use it is to load a library from disk, and then iterate the object tree:

	lib, _ := quiver.ReadLibrary("/path/to/library.quiverlibrary", false)

	// Print the title of all the notes in all the notebooks
	for _, notebooks := range lib.Notebooks {
		for _, note := notebook.Notes {
			fmt.Println(n.Title)
			//...
		}
	}
*/
package quiver

import (
	"bufio"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"mime"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// The version of the quiver package
const Version = "0.3.4"

// The data tree: holds all the data of the library

// Library represents the contents of a Quiver library (.qvlibrary) file.
type Library struct {
	*LibraryMetadata
	// The list of Notebooks found inside the Library.
	Notebooks []*Notebook `json:"notebooks"`
}

// LibraryMetadata represents the contents of a Quiver library metadata (meta.json) file.
type LibraryMetadata struct {
	// The root of the notebook hierarchy
	Children []NotebookHierarchyInfo `json:"children"`
}

// A note in the Quote notebooks hierarchy
type NotebookHierarchyInfo struct {
	// The UUID of the Notebook.
	UUID string `json:"uuid"`
	// The list of its children
	Children []NotebookHierarchyInfo `json:"children"`
}

// Notebook represents the contents of a Quiver notebook (.qvnotebook) directory.
type Notebook struct {
	*NotebookMetadata
	// The list of Notes found inside the Notebook.
	Notes []*Note `json:"notes"`
}

// NotebookMetadata represents the contents of a Quiver notebook (.qvnotebook) directory.
type NotebookMetadata struct {
	// The Name of the Notebook.
	Name string `json:"name"`
	// The UUID of the Notebook.
	UUID string `json:"uuid"`
}

// NoteContent represents the contents of a Quiver note (.qvnote) directory.
type Note struct {
	*NoteMetadata
	*NoteContent
	// The list of all Resources attached to this Note.
	Resources []*NoteResource `json:"resources,omitempty"`
}

// NoteMetadata represents the contents of a Quiver note metadata (meta.json) file.
type NoteMetadata struct {
	// The time the note was created.
	CreatedAt TimeStamp `json:"created_at"`
	// A list of tags attached to the Note.
	Tags []string `json:"tags"`
	// The Title of the Note.
	Title string `json:"title"`
	// The time the note was last updated.
	UpdatedAt TimeStamp `json:"updated_at"`
	// The UUID of the Note.
	UUID string `json:"uuid"`
}

// A timestamp in a Quiver note metadata file (meta.json).
// It holds time info (from time.Time) and marshals as an integer.
type TimeStamp time.Time

// MarshalJSON marshals TimeStamp as an integer (seconds since Epoch).
func (u *TimeStamp) MarshalJSON() ([]byte, error) {
	secs := (*time.Time)(u).Unix()
	return json.Marshal(secs)
}

// MarshalJSON unmarshals TimeStamp from an integer (seconds since Epoch).
func (u *TimeStamp) UnmarshalJSON(data []byte) error {
	var secs int64
	err := json.Unmarshal(data, &secs)
	if err != nil {
		return err
	}

	// copy values
	*u = TimeStamp(time.Unix(secs, 0))

	return nil
}

// NoteContent represents the contents of a Quiver note resource: any file found under the resources/ folder in the note.
type NoteResource struct {
	// The file name.
	Name string `json:"name"`
	// The file relative path.
	Rel string `json:"rel"`
	// The file data as raw bytes.
	// It serializes in JSON as a data URI.
	Data []byte `json:"data"`
}

// MarshalJSON marshals
func (n *NoteResource) MarshalJSON() ([]byte, error) {
	// Build a data uri for the resource
	ext := filepath.Ext(n.Name)
	mimeType := mime.TypeByExtension(ext)
	b64 := base64.RawURLEncoding.EncodeToString(n.Data)
	url := fmt.Sprintf("data:%v,%v", mimeType, b64)

	// And then encode the uri as a JSON string
	aux := struct {
		Name string
		Data string
	}{
		n.Name,
		url,
	}
	return json.Marshal(aux)
}

// UnmarshalJSON unmarshals NoteResource from data:// url
func (u *NoteResource) UnmarshalJSON(data []byte) error {
	var aux struct {
		Name string
		URL  string
	}
	err := json.Unmarshal(data, aux)
	if err != nil {
		return err
	}

	// Name OK
	u.Name = aux.Name

	// Split data url
	if !strings.HasPrefix(aux.URL, "data:") {
		return fmt.Errorf("invalid data URL %v", aux.URL)
	}
	s := strings.SplitN(aux.URL, ",", 2)
	if len(s) != 2 {
		return fmt.Errorf("data URL %v", aux.URL)
	}

	// Decode the base64-encoded data
	resData, err := base64.RawURLEncoding.DecodeString(s[1])
	if err != nil {
		return err
	}

	// Data found !
	u.Data = resData

	return nil
}

// NoteContent represents the contents of a Quiver not content (content.json) file.
//
// Beware: this structure does note contain the Title of the cell, since it is already held in the
// NoteMetadata file.
type NoteContent struct {
	// The list of all cells in the note.
	Cells []*Cell `json:"cells"`
}

// The type of a cell inside of a Quiver Note
type CellType string

// The recognized types of Quiver cells
const (
	CodeCell     CellType = "code"
	TextCell     CellType = "text"
	MarkdownCell CellType = "markdown"
	LatexCell    CellType = "latex"
	DiagramCell  CellType = "diagram"
)

// A cell inside a Quiver note.
type Cell struct {
	// The type of the cell.
	Type CellType `json:"type"`
	// The language of the cell: only relevant for CodeCell cells.
	Language string `json:"language,omitempty"`
	// The type of diagram: only relevant for DiagramCell cells.
	DiagramType string `json:"diagramType,omitempty"`
	// The data for the cell, aka. all the actual content.
	Data string `json:"data"`
}

// IsCode returns true if the Cell is of Type CodeCell.
func (c *Cell) IsCode() bool {
	return c.Type == CodeCell
}

// IsMarkdown returns true if the Cell is of Type MarkdownCell.
func (c *Cell) IsMarkdown() bool {
	return c.Type == MarkdownCell
}

// IsText returns true if the Cell is of Type TextCell.
func (c *Cell) IsText() bool {
	return c.Type == TextCell
}

// IsLatex returns true if the Cell is of Type LatexCell.
func (c *Cell) IsLatex() bool {
	return c.Type == LatexCell
}

// IsDiagram returns true if the Cell is of Type DiagramCell.
func (c *Cell) IsDiagram() bool {
	return c.Type == DiagramCell
}

// IsLibrary checks that the element at the given path is indeed a Quiver library, and
// returns true if found or false with an error otherwise.
func IsLibrary(path string) (bool, error) {
	// it should exist and be a library
	stat, err := os.Stat(path)
	if err != nil {
		return false, err
	}
	if !stat.IsDir() {
		return false, errors.New("the Quiver Library should be a dictionary")
	}
	// and end with .qvlibrary
	if !strings.HasSuffix(stat.Name(), ".qvlibrary") {
		return false, errors.New("the Quiver Library should have .qvlibrary extension")
	}

	return true, nil
}

// ReadLibrary loads the Quiver library at the given path.
// The loadResources parameter tells the function if note resources should be loaded too.
func ReadLibrary(path string, loadResources bool) (*Library, error) {
	_, err := IsLibrary(path)
	if err != nil {
		return nil, err
	}

	// list the files in the library (aka. the notes)
	files, err := ioutil.ReadDir(path)
	if err != nil {
		return nil, err
	}

	var metadata *LibraryMetadata
	notebooks := make([]*Notebook, 0, len(files))
	for _, f := range files {
		p := filepath.Join(path, f.Name())

		// ignore root meta.json
		if f.Name() == "meta.json" {
			metadata, err = ReadLibraryMetadata(p)
			if err != nil {
				return nil, err
			}
		} else {
			// all other elements should be notebooks
			n, err := ReadNotebook(p, loadResources)
			if err != nil {
				return nil, err
			}
			notebooks = append(notebooks, n)
		}
	}

	return &Library{metadata, notebooks}, nil
}

// WalkNotebooksHierarchy returns all the notebooks in order, allowing to "explore" the internal hierarchy of the
// Quiver library.
func (m *Library) WalkNotebooksHierarchy(f func(n *Notebook, parents []*Notebook) error) error {
	notebooks := make(map[string]*Notebook, len(m.Notebooks))
	for _, n := range m.Notebooks {
		notebooks[n.UUID] = n
	}

	parents := make([]string, 0)
	for _, n := range m.LibraryMetadata.Children {
		err := walkNotebooksHierarchy(n, parents, func(c string, parents []string) error {
			pp := make([]*Notebook, len(parents))
			for i, p := range parents {
				pp[i] = notebooks[p]
			}
			return f(notebooks[c], pp)
		})
		if err != nil {
			return err
		}
	}

	return nil
}

func walkNotebooksHierarchy(n NotebookHierarchyInfo, parents []string, f func(n string, parents []string) error) error {
	err := f(n.UUID, parents)
	if err != nil {
		return err
	}

	if len(n.Children) > 0 {
		p := append(parents[:], n.UUID)
		for _, c := range n.Children {
			err = walkNotebooksHierarchy(c, p, f)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// ReadLibraryMetadata loads the library "meta.json" at the given path.
func ReadLibraryMetadata(path string) (*LibraryMetadata, error) {
	// find and read metadata file
	mf, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer mf.Close()

	// Read metadata
	buf := bufio.NewReader(mf)
	return ParseLibraryMetadata(buf)
}

// IsNoteBook checks that the element at the given path is indeed a Quiver notebook, and
// returns true if found or false with an error otherwise.
func IsNotebook(path string) (bool, error) {
	// it should exist and be a directory
	stat, err := os.Stat(path)
	if err != nil {
		return false, err
	}
	if !stat.IsDir() {
		return false, errors.New("a Quiver Notebook should be a directory")
	}
	// and end with .qvnotebook
	if !strings.HasSuffix(stat.Name(), ".qvnotebook") {
		return false, errors.New("a Quiver Notebook should have .qvnotebook extension")
	}

	return true, nil
}

// ReadNotebook loads the Quiver notebook in the given path.
// The loadResources parameter tells the function if note resources should be loaded too.
func ReadNotebook(path string, loadResources bool) (*Notebook, error) {
	_, err := IsNotebook(path)
	if err != nil {
		return nil, err
	}

	// list the files in the notebook (aka. the notes)
	files, err := ioutil.ReadDir(path)
	if err != nil {
		return nil, err
	}

	var metadata *NotebookMetadata
	var numNotes = 0
	if len(files) > 1 {
		numNotes = len(files) - 1
	}
	notes := make([]*Note, numNotes)
	for i, f := range files {
		p := filepath.Join(path, f.Name())
		if f.Name() == "meta.json" {
			metadata, err = ReadNotebookMetadata(p)
			if err != nil {
				return nil, err
			}
		} else {
			n, err := ReadNote(p, loadResources)
			if err != nil {
				return nil, err
			}
			notes[i] = n
		}
	}

	return &Notebook{metadata, notes}, nil
}

// IsNote checks that the element at the given path is indeed a Quiver note, and
// returns true if found or false with an error otherwise.
func IsNote(path string) (bool, error) {
	// it should exist and be a directory
	stat, err := os.Stat(path)
	if err != nil {
		return false, err
	}
	if !stat.IsDir() {
		return false, errors.New("a Quiver Note should be a directory")
	}
	// and end with .qvnote
	if !strings.HasSuffix(stat.Name(), ".qvnote") {
		return false, errors.New("a Quiver Note should have .qvnote extension")
	}

	return true, nil
}

// ReadNote loads the Quiver note in the given path.
// The loadResources parameter tells the function if note resources should be loaded too.
func ReadNote(path string, loadResources bool) (*Note, error) {
	_, err := IsNote(path)
	if err != nil {
		return nil, err
	}

	// Read the metadata file
	mp := filepath.Join(path, "meta.json")
	m, err := ReadNoteMetadata(mp)
	if err != nil {
		return nil, err
	}

	// Read the content file
	cp := filepath.Join(path, "content.json")
	c, err := ReadNoteContent(cp)
	if err != nil {
		return nil, err
	}

	var res []*NoteResource
	if loadResources {
		rp := filepath.Join(path, "resources")
		res, err = ReadNoteResources(rp, "")
		// we check for error but ignore not existing dir
		if err != nil && !os.IsNotExist(err) {
			return nil, err
		}
	}

	return &Note{m, c, res}, nil
}

// ReadNoteResource loads the resource (any file actually) into a NoteResource instance.
func ReadNoteResources(path string, rel string) ([]*NoteResource, error) {
	stat, err := os.Stat(path)
	if err != nil {
		return nil, err
	}
	if !stat.IsDir() {
		return nil, errors.New("quiver Note Resources should be held in a directory")
	}

	files, err := ioutil.ReadDir(path)
	if err != nil {
		return nil, err
	}

	res := []*NoteResource{}
	for _, file := range files {
		name := file.Name()
		fp := filepath.Join(path, name)

		stat, err = os.Stat(fp)
		if err != nil {
			return nil, err
		}
		if stat.IsDir() {
			rp := filepath.Join(rel, name)
			r, err := ReadNoteResources(fp, rp)
			if err != nil {
				return nil, err
			}
			res = append(res, r...)
		} else {
			// Read the file completely in memory
			f, err := os.Open(fp)
			if err != nil {
				return nil, err
			}
			buf, err := ioutil.ReadAll(f)
			f.Close()
			if err != nil {
				return nil, err
			}

			// And add the note to the list
			res = append(res, &NoteResource{name, rel, buf})
		}
	}

	return res, nil
}

// ReadNoteResource loads the note "meta.json" at the given path.
func ReadNoteMetadata(path string) (*NoteMetadata, error) {
	// find and read metadata file
	mf, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer mf.Close()

	// Read metadata
	buf := bufio.NewReader(mf)
	return ParseNoteMetadata(buf)
}

// ReadNoteContent loads the note "content.json" at the given path.
func ReadNoteContent(path string) (*NoteContent, error) {
	// find and read content file
	cf, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer cf.Close()

	// Read Content
	buf := bufio.NewReader(cf)
	return ParseContent(buf)
}

// ReadNotebookMetadata loads the notebook "meta.json" at the given path.
func ReadNotebookMetadata(path string) (*NotebookMetadata, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	buf := bufio.NewReader(f)
	return ParseNotebookMetadata(buf)
}

// ParseLibraryMetadata loads the JSON from the given stream into a LibraryMetadata.
func ParseLibraryMetadata(r io.Reader) (*LibraryMetadata, error) {
	d := json.NewDecoder(r)
	m := new(LibraryMetadata)
	err := d.Decode(m)
	if err != nil {
		return nil, err
	}
	return m, nil
}

// ParseNotebookMetadata loads the JSON from the given stream into a NotebookMetadata.
func ParseNotebookMetadata(r io.Reader) (*NotebookMetadata, error) {
	d := json.NewDecoder(r)
	m := new(NotebookMetadata)
	err := d.Decode(m)
	if err != nil {
		return nil, err
	}
	return m, nil
}

// ParseNoteMetadata loads the JSON from the given stream into a NoteMetadata.
func ParseNoteMetadata(r io.Reader) (*NoteMetadata, error) {
	d := json.NewDecoder(r)
	m := new(NoteMetadata)
	err := d.Decode(m)
	if err != nil {
		return nil, err
	}
	return m, nil
}

// ParseNoteMetadata loads the JSON from the given stream into a NoteContent.
func ParseContent(r io.Reader) (*NoteContent, error) {
	d := json.NewDecoder(r)
	n := new(NoteContent)
	err := d.Decode(n)
	if err != nil {
		return nil, err
	}
	return n, err
}
