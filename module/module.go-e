package module

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"text/template"
)

type Module struct {
	Main Package
}

type Package interface {
	GenerateDataStructures()
	GenerateCallables()
	SymbolTable() SymbolTable
	Name() string
	File(string) *os.File
	AddFile(string, *os.File)
}

type pack struct {
	files       map[string]*os.File
	PackageName string
	SymTable    SymbolTable
	subPackages []Package
}

func (p *pack) Name() string {
	return p.PackageName
}

func (p *pack) SymbolTable() SymbolTable {
	return p.SymTable
}

func (p *pack) File(path string) *os.File {
	return p.files[path]
}

func (p *pack) AddFile(path string, file *os.File) {
	p.files[path] = file
}

func (p *pack) GenerateDataStructures() {
	err := generate("data_structures", p, "package {{.Name}}\n\n{{range .SymTable.Table}}\n var {{.Name}} =  {{.Value}}\n{{end}}")
	if err != nil {
		log.Fatal(err)
	}
}

func (p *pack) GenerateCallables() {
	err := generate("callables", p, "\n\n{{range .SymTable.Table}}\n var {{.Name}} =  {{.Value}}\n{{end}}")
	if err != nil {
		log.Fatal(err)
	}
}

func NewModule() Module {
	pack := NewPackage("main", nil)
	CreatePackageDir(pack)
	CreateFile("main", pack)
	mainCode := "package main \n func main() {}"
	_, err := pack.File(fmt.Sprint("./", sourceOutputDirectory, "/", packageMain, "/main.go")).Write([]byte(mainCode))
	if err != nil {
		panic(err)
	}
	return Module{
		Main: NewPackage("main", nil),
	}
}

func NewPackage(name string, parent Package) Package {
	if parent != nil {
		return &pack{PackageName: fmt.Sprint(parent.Name(), "/", name), SymTable: NewSymbolTable(), files: map[string]*os.File{}}
	} else {
		return &pack{PackageName: name, SymTable: NewSymbolTable(), files: map[string]*os.File{}}
	}
}

func generate(filename string, p Package, text string) error {
	tmpl, err := template.New(filename).Parse(text)
	if err != nil {
		panic(err)
	}

	newpath := filepath.Join(".", ".obj")
	err = os.MkdirAll(newpath, os.ModePerm)
	if err != nil {
		return err
	}

	f, err := os.Create(fmt.Sprint("./.obj/", filename, ".go"))
	if err != nil {
		return err
	}

	defer func(f *os.File) {
		err := f.Close()
		if err != nil {

		}
	}(f)

	err = tmpl.Execute(f, p)
	if err != nil {
		return err
	}

	return nil
}

var Mod = NewModule()
