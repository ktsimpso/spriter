package imagestitcher

import (
	"image"
	"image/draw"
	_ "image/gif"
	_ "image/jpeg"
	"image/png"
	"os"
)

func Combine(imagePaths []string, destination string) (map[string]image.Rectangle, error) {
	m := make(map[string]image.Rectangle)
	images := []image.Image{}
	bounds := image.Rect(0, 0, 0, 0)

	for _, imagePath := range imagePaths {
		imageFile, err := os.Open(imagePath)
		if err != nil {
			return nil, err
		}
		defer imageFile.Close()

		i, _, err := image.Decode(imageFile)
		if err != nil {
			return nil, err
		}
		images = append(images, i)
		//TODO: Make this smarter see http://www.codeproject.com/Articles/210979/Fast-optimizing-rectangle-packing-algorithm-for-bu
		bounds.Max = image.Point{
			max(bounds.Max.X, i.Bounds().Max.X),
			bounds.Max.Y + i.Bounds().Max.Y,
		}
	}

	destinationImage := image.NewRGBA(bounds)
	currentMax := image.Point{0, 0}

	for index, sourceImage := range images {
		drawRectangle := image.Rectangle{
			currentMax,
			currentMax.Add(sourceImage.Bounds().Size()),
		}
		draw.Draw(destinationImage, drawRectangle, sourceImage, sourceImage.Bounds().Min, draw.Src)
		currentMax.Y += sourceImage.Bounds().Max.Y
		m[imagePaths[index]] = drawRectangle
	}

	destinationFile, err := os.Create(destination)
	if err != nil {
		return nil, err
	}
	defer destinationFile.Close()

	err = png.Encode(destinationFile, destinationImage)
	if err != nil {
		return nil, err
	}

	return m, nil
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
