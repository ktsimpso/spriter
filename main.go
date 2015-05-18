package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/ktsimpso/spriter/css"
	"github.com/ktsimpso/spriter/imagestitcher"
)

func main() {

	cssFileName := "css/styles.css"
	imagePaths, err := css.GetPaths(cssFileName)
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println(imagePaths)

	err = os.Chdir(filepath.Dir(cssFileName))
	if err != nil {
		fmt.Println(err)
		return
	}

	m, err := imagestitcher.Combine(imagePaths, "../img/sprite.png")
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println(m)
}
