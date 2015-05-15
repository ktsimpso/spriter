package main

import (
	"fmt"

	"github.com/ktsimpso/spriter/imagestitcher"
)

func main() {
	imagePaths := []string{
		"img/alien.jpg",
		"img/apple.png",
		"img/drop.png",
		"img/layout_rows.png",
		"img/settings.png",
		"img/star.jpg",
	}

	_, err := imagestitcher.Combine(imagePaths, "img/sprite.png")
	if err != nil {
		fmt.Println(err)
	}
}
