package utils

import (
	"github.com/go-gl/gl/v3.3-core/gl"
	"github.com/tsunyoku/danser/framework/graphics/texture"
	"log"
	"os"
	"time"
)

func MakeScreenshot(w, h int, name string, async bool) {
	pixmap := texture.NewPixMapC(w, h, 3)

	gl.PixelStorei(gl.PACK_ALIGNMENT, int32(1))
	gl.ReadPixels(0, 0, int32(w), int32(h), gl.RGB, gl.UNSIGNED_BYTE, pixmap.RawPointer)

	save := func() {
		defer pixmap.Dispose()

		err := os.Mkdir("screenshots", 0755)
		if err != nil && !os.IsExist(err) {
			log.Println("Failed to save the screenshot! Error:", err)
			return
		}

		fileName := name

		if fileName == "" {
			fileName = "danser_" + time.Now().Format("2006-01-02_15-04-05")
		}

		fileName += ".png"

		err = pixmap.WritePng("screenshots/"+fileName, true)
		if err != nil {
			log.Println("Failed to save the screenshot! Error:", err)
			return
		}

		log.Println("Screenshot", fileName, "saved!")
	}

	if async {
		go save()
	} else {
		save()
	}
}
