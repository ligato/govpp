//  Copyright (c) 2020 Cisco and/or its affiliates.
//
//  Licensed under the Apache License, Version 2.0 (the "License");
//  you may not use this file except in compliance with the License.
//  You may obtain a copy of the License at:
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
//  Unless required by applicable law or agreed to in writing, software
//  distributed under the License is distributed on an "AS IS" BASIS,
//  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//  See the License for the specific language governing permissions and
//  limitations under the License.

package binapigen

import (
	"bufio"
	"bytes"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/printer"
	"go/token"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/sirupsen/logrus"

	"git.fd.io/govpp.git/binapigen/vppapi"
)

type Options struct {
	OutputDir     string // output directory for generated files
	ImportPrefix  string // prefix for import paths
	NoVersionInfo bool   // disables generating version info
}

type Generator struct {
	Files       []*File
	FilesByName map[string]*File

	opts       Options
	apifiles   []*vppapi.File
	vppVersion string

	filesToGen []string
	genfiles   []*GenFile

	enumsByName    map[string]*Enum
	aliasesByName  map[string]*Alias
	structsByName  map[string]*Struct
	unionsByName   map[string]*Union
	messagesByName map[string]*Message
}

func New(opts Options, apifiles []*vppapi.File, filesToGen []string) (*Generator, error) {
	gen := &Generator{
		FilesByName:    make(map[string]*File),
		opts:           opts,
		apifiles:       apifiles,
		filesToGen:     filesToGen,
		enumsByName:    map[string]*Enum{},
		aliasesByName:  map[string]*Alias{},
		structsByName:  map[string]*Struct{},
		unionsByName:   map[string]*Union{},
		messagesByName: map[string]*Message{},
	}

	SortFilesByImports(gen.apifiles)

	if gen.opts.ImportPrefix == "" {
		gen.opts.ImportPrefix = resolveImportPath(gen.opts.OutputDir)
		logrus.Debugf("resolved import prefix: %s", gen.opts.ImportPrefix)
	}

	// prepare package names and import paths
	packageNames := make(map[string]GoPackageName)
	importPaths := make(map[string]GoImportPath)

	for _, apifile := range gen.apifiles {
		filename := getFilename(apifile)
		packageNames[filename] = cleanPackageName(apifile.Name)
		importPaths[filename] = GoImportPath(path.Join(gen.opts.ImportPrefix, baseName(apifile.Name))) // GoImportPath(baseName(apifile.Name))

		importedTypes := ListImportedTypes(gen.apifiles, apifile)
		logrus.Debugf("%s has %d imported types: %v", apifile.Name, len(importedTypes), importedTypes)
	}

	logrus.Debugf("adding %d VPP API files to generator", len(gen.apifiles))

	for _, apifile := range gen.apifiles {
		filename := getFilename(apifile)

		if _, ok := gen.FilesByName[apifile.Name]; ok {
			return nil, fmt.Errorf("duplicate file: %q", apifile.Name)
		}

		file, err := newFile(gen, apifile, packageNames[filename], importPaths[filename])
		if err != nil {
			return nil, fmt.Errorf("loading file %s failed: %w", apifile.Name, err)
		}
		gen.Files = append(gen.Files, file)
		gen.FilesByName[apifile.Name] = file

		logrus.Debugf("added file %q (path: %v)", apifile.Name, apifile.Path)
		if len(file.Imports) > 0 {
			logrus.Debugf(" - %d imports: %v", len(file.Imports), file.Imports)
		}
	}

	// mark files for generation
	if len(gen.filesToGen) > 0 {
		logrus.Debugf("Checking %d files to generate: %v", len(gen.filesToGen), gen.filesToGen)
		for _, genfile := range gen.filesToGen {
			file, ok := gen.FilesByName[genfile]
			if !ok {
				return nil, fmt.Errorf("no API file found for: %v", genfile)
			}
			file.Generate = true
			// generate all imported files
			for _, impFile := range file.importedFiles(gen) {
				impFile.Generate = true
			}
		}
	} else {
		logrus.Debugf("Files to generate not specified, marking all %d files for generate", len(gen.Files))
		for _, file := range gen.Files {
			file.Generate = true
		}
	}

	return gen, nil
}
func getFilename(file *vppapi.File) string {
	if file.Path == "" {
		return file.Name
	}
	return file.Path
}

func (g *Generator) Generate() error {
	if len(g.genfiles) == 0 {
		return fmt.Errorf("no files to generate")
	}

	logrus.Infof("Generating %d files", len(g.genfiles))

	for _, genfile := range g.genfiles {
		content, err := genfile.Content()
		if err != nil {
			return err
		}
		if err := writeSourceTo(genfile.filename, content); err != nil {
			return fmt.Errorf("writing source package %s failed: %v", genfile.filename, err)
		}
	}
	return nil
}

type GenFile struct {
	gen           *Generator
	file          *File
	filename      string
	goImportPath  GoImportPath
	buf           bytes.Buffer
	manualImports map[GoImportPath]bool
	packageNames  map[GoImportPath]GoPackageName
}

func (g *Generator) NewGenFile(filename string, importPath GoImportPath) *GenFile {
	f := &GenFile{
		gen:           g,
		filename:      filename,
		goImportPath:  importPath,
		manualImports: make(map[GoImportPath]bool),
		packageNames:  make(map[GoImportPath]GoPackageName),
	}
	g.genfiles = append(g.genfiles, f)
	return f
}

func (g *GenFile) Write(p []byte) (n int, err error) {
	return g.buf.Write(p)
}

func (g *GenFile) Import(importPath GoImportPath) {
	g.manualImports[importPath] = true
}

func (g *GenFile) GoIdent(ident GoIdent) string {
	if ident.GoImportPath == g.goImportPath {
		return ident.GoName
	}
	if packageName, ok := g.packageNames[ident.GoImportPath]; ok {
		return string(packageName) + "." + ident.GoName
	}
	packageName := cleanPackageName(baseName(string(ident.GoImportPath)))
	g.packageNames[ident.GoImportPath] = packageName
	return string(packageName) + "." + ident.GoName
}

func (g *GenFile) P(v ...interface{}) {
	for _, x := range v {
		switch x := x.(type) {
		case GoIdent:
			fmt.Fprint(&g.buf, g.GoIdent(x))
		default:
			fmt.Fprint(&g.buf, x)
		}
	}
	fmt.Fprintln(&g.buf)
}

func (g *GenFile) Content() ([]byte, error) {
	if !strings.HasSuffix(g.filename, ".go") {
		return g.buf.Bytes(), nil
	}
	return g.injectImports(g.buf.Bytes())
}

func (g *GenFile) injectImports(original []byte) ([]byte, error) {
	// Parse source code
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "", original, parser.ParseComments)
	if err != nil {
		var src bytes.Buffer
		s := bufio.NewScanner(bytes.NewReader(original))
		for line := 1; s.Scan(); line++ {
			fmt.Fprintf(&src, "%5d\t%s\n", line, s.Bytes())
		}
		return nil, fmt.Errorf("%v: unparsable Go source: %v\n%v", g.filename, err, src.String())
	}

	// Prepare list of all imports
	var importPaths [][2]string
	for importPath := range g.packageNames {
		pkgName := string(g.packageNames[GoImportPath(importPath)])
		pkgPath := string(importPath)
		importPaths = append(importPaths, [2]string{pkgName, pkgPath})
	}
	for importPath := range g.manualImports {
		if _, ok := g.packageNames[importPath]; !ok {
			pkgPath := string(importPath)
			importPaths = append(importPaths, [2]string{"_", pkgPath})
		}
	}
	// Sort imports by import path
	sort.Slice(importPaths, func(i, j int) bool {
		return importPaths[i][1] < importPaths[j][1]
	})
	// Inject new import block into parsed AST
	if len(importPaths) > 0 {
		pos := file.Package
		tokFile := fset.File(file.Package)
		pkgLine := tokFile.Line(file.Package)
		for _, c := range file.Comments {
			if tokFile.Line(c.Pos()) > pkgLine {
				break
			}
			pos = c.End()
		}
		// Prepare the import block
		impDecl := &ast.GenDecl{Tok: token.IMPORT, TokPos: pos, Lparen: pos, Rparen: pos}
		for _, importPath := range importPaths {
			var name *ast.Ident
			if importPath[0] == "_" || strings.Contains(importPath[1], ".") {
				name = &ast.Ident{Name: importPath[0], NamePos: pos}
			}
			impDecl.Specs = append(impDecl.Specs, &ast.ImportSpec{
				Name:   name,
				Path:   &ast.BasicLit{Kind: token.STRING, Value: strconv.Quote(importPath[1]), ValuePos: pos},
				EndPos: pos,
			})
		}
		file.Decls = append([]ast.Decl{impDecl}, file.Decls...)
	}
	// Reformat source code
	var out bytes.Buffer
	cfg := &printer.Config{
		Mode:     printer.TabIndent | printer.UseSpaces,
		Tabwidth: 8,
	}
	if err = cfg.Fprint(&out, fset, file); err != nil {
		return nil, fmt.Errorf("%v: can not reformat Go source: %v", g.filename, err)
	}
	return out.Bytes(), nil
}

// GoIdent is a Go identifier, consisting of a name and import path.
// The name is a single identifier and may not be a dot-qualified selector.
type GoIdent struct {
	GoName       string
	GoImportPath GoImportPath
}

func (id GoIdent) String() string {
	return fmt.Sprintf("%q.%v", id.GoImportPath, id.GoName)
}

func newGoIdent(f *File, fullName string) GoIdent {
	name := strings.TrimPrefix(fullName, string(f.PackageName)+".")
	return GoIdent{
		GoName:       camelCaseName(name),
		GoImportPath: f.GoImportPath,
	}
}

type GoImportPath string

func (p GoImportPath) String() string {
	return strconv.Quote(string(p))
}

func (p GoImportPath) Ident(s string) GoIdent {
	return GoIdent{GoName: s, GoImportPath: p}
}

type GoPackageName string

func cleanPackageName(name string) GoPackageName {
	return GoPackageName(sanitizedName(name))
}

// baseName returns the last path element of the name, with the last dotted suffix removed.
func baseName(name string) string {
	// First, find the last element
	if i := strings.LastIndex(name, "/"); i >= 0 {
		name = name[i+1:]
	}
	// Now drop the suffix
	if i := strings.LastIndex(name, "."); i >= 0 {
		name = name[:i]
	}
	return name
}

func writeSourceTo(outputFile string, b []byte) error {
	// create output directory
	packageDir := filepath.Dir(outputFile)
	if err := os.MkdirAll(packageDir, 0775); err != nil {
		return fmt.Errorf("creating output dir %s failed: %v", packageDir, err)
	}

	// format generated source code
	gosrc, err := format.Source(b)
	if err != nil {
		_ = ioutil.WriteFile(outputFile, b, 0666)
		return fmt.Errorf("formatting source code failed: %v", err)
	}

	// write generated code to output file
	if err := ioutil.WriteFile(outputFile, gosrc, 0666); err != nil {
		return fmt.Errorf("writing to output file %s failed: %v", outputFile, err)
	}

	lines := bytes.Count(gosrc, []byte("\n"))
	logf("wrote %d lines (%d bytes) to: %q", lines, len(gosrc), outputFile)

	return nil
}

func resolveImportPath(outputDir string) string {
	absPath, err := filepath.Abs(outputDir)
	if err != nil {
		panic(err)
	}
	modRoot := findModuleRoot(absPath)
	if modRoot == "" {
		logrus.Fatalf("module root not found at: %s", absPath)
	}
	modPath := findModulePath(path.Join(modRoot, "go.mod"))
	if modPath == "" {
		logrus.Fatalf("module path not found")
	}
	relDir, err := filepath.Rel(modRoot, absPath)
	if err != nil {
		panic(err)
	}
	return filepath.Join(modPath, relDir)
}

func findModuleRoot(dir string) (root string) {
	if dir == "" {
		panic("dir not set")
	}
	dir = filepath.Clean(dir)
	// Look for enclosing go.mod.
	for {
		if fi, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil && !fi.IsDir() {
			return dir
		}
		d := filepath.Dir(dir)
		if d == dir {
			break
		}
		dir = d
	}
	return ""
}

var (
	modulePathRE = regexp.MustCompile(`module[ \t]+([^ \t\r\n]+)`)
)

func findModulePath(file string) string {
	data, err := ioutil.ReadFile(file)
	if err != nil {
		return ""
	}
	m := modulePathRE.FindSubmatch(data)
	if m == nil {
		return ""
	}
	return string(m[1])
}
