package module

import (
	"fmt"
	"log"
	"os"
)

func CreatePackageDir(pack Package) {
	err := os.MkdirAll(fmt.Sprint("./", sourceOutputDirectory, "/", pack.Name()), os.ModePerm)
	if err != nil {
		return
	}
}

func CreateFile(name string, pack Package) {
	path := fmt.Sprint("./", sourceOutputDirectory, "/", pack.Name(), "/", name, ".go")
	f, err := os.Create(path)
	if err != nil {
		log.Fatal(err)
	}
	pack.AddFile(path, f)
}
