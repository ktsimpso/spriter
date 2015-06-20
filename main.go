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
	t, err := css.GetParseTree(cssFileName)
	if err != nil {
		fmt.Println(err)
		return
	}

	imagePaths := css.GetPaths(t)
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

	css.AddSpriteToCss(t, "../img/sprite.png", m)

	err = css.WriteToFile(t, "sprited.css")
	if err != nil {
		fmt.Println(err)
	}
}
