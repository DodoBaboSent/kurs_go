package main

import (
	"strings"
	"syscall/js"

	"github.com/lithammer/fuzzysearch/fuzzy"
)

var out = ""

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func fileSearch() js.Func {
	return js.FuncOf(func(this js.Value, args []js.Value) any {

		println(args[0].String())

		needle := strings.ToLower(args[0].String())

		found := fuzzy.Find(needle, strings.Split(out, " "))

		htmlTemplate := ""

		for i := 0; i < len(found); i++ {
			ind := strings.Index(out, found[i])
			ind_clamped := max(0, ind-125)
			upp_ind_clamped := min(len(out), ind+125)
			htmlTemplate += "<div class=\"px-3 border-dashed border-amber-500 border-2 rounded\">..." + out[ind_clamped:ind] + "<strong>" + out[ind:ind+len(found[i])] + "</strong>" + out[ind+len(found[i]):upp_ind_clamped] + "...</div>\n"
		}

		document := js.Global().Get("document")
		fileOutput := document.Call("getElementById", "searchOutput")
		fileOutput.Set("innerHTML", htmlTemplate)
		fileOutput.Set("style", "height: 400px; border-width: 2px;")
		return nil
	})

}

func SendFile() {
	document := js.Global().Get("document")

	fileInput := document.Call("getElementById", "fileInput")

	fileInput.Set("oninput", js.FuncOf(func(v js.Value, x []js.Value) any {
		fileInput.Get("files").Call("item", 0).Call("arrayBuffer").Call("then", js.FuncOf(func(v js.Value, x []js.Value) any {
			data := js.Global().Get("Uint8Array").New(x[0])
			dst := make([]byte, data.Get("length").Int())
			js.CopyBytesToGo(dst, data)

			out = string(dst)

			fileOutput := document.Call("getElementById", "fileOutput")
			fileOutput.Set("innerText", out)
			fileOutput.Set("style", "height: 200px; border-width: 2px;")

			return nil
		}))

		return nil
	}))
}

func main() {

	ch := make(chan struct{}, 0)
	js.Global().Set("fileSearch", fileSearch())
	SendFile()
	<-ch
}
