package main

import (
	"strings"
	"syscall/js"

	"github.com/lithammer/fuzzysearch/fuzzy"
)

var htmlString = `<h1>TEST TEST</h1>`

func GetHtml() js.Func {
	return js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		return htmlString
	})
}

var out = ""

// func bitap(text string, pattern string, k int) int {
// 	result := -1

// 	m := len(pattern)

// 	size := (k + 1) * int(unsafe.Sizeof(k))

// 	var R []int
// 	var patternMask [1024]int = *new([1024]int)
// 	var i, d int

// 	if len(pattern) == 0 {
// 		return 0
// 	}
// 	if m > 31 {
// 		return -1
// 	}

// 	R = make([]int, size)
// 	for i := 0; i <= k; i++ {
// 		R[i] = ^i
// 	}

// 	for i = 0; i <= 127; i++ {
// 		patternMask[i] = ^0
// 	}

// 	for i = 0; i < m; i++ {
// 		patternMask[pattern[i]] &= ^(1 << i)
// 	}

// 	for i = 0; i < len(text); i++ {
// 		var oldRd1 = R[0]

// 		R[0] |= patternMask[text[i]]
// 		R[0] <<= 1

// 		for d = 1; d <= k; d++ {
// 			var tmp int = R[d]

// 			R[d] = (oldRd1 & (R[d] | patternMask[text[i]])) << 1

// 			oldRd1 = tmp
// 		}

// 		if 0 == (R[k] & (1 << m)) {
// 			result = (i - m) + 1
// 			break
// 		}
// 	}
// 	return result
// }

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
			// println(upp_ind_clamped, ind_clamped)
			htmlTemplate += "<div class=\"px-3 border-dashed border-amber-500 border-2 rounded\">..." + out[ind_clamped:ind] + "<strong>" + out[ind:ind+len(found[i])] + "</strong>" + out[ind+len(found[i]):upp_ind_clamped] + "...</div>\n"
		}

		document := js.Global().Get("document")
		fileOutput := document.Call("getElementById", "searchOutput")
		// index := bitap(strings.ToLower(out), strings.ToLower(args[0].String()), 3)
		fileOutput.Set("innerHTML", htmlTemplate)
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
			// if len(out) > 100 {
			// 	out = out[:100] + "..."
			// }

			fileOutput := document.Call("getElementById", "fileOutput")
			fileOutput.Set("innerText", out)

			return nil
		}))

		return nil
	}))
}

func main() {

	ch := make(chan struct{}, 0)
	js.Global().Set("getHtml", GetHtml())
	js.Global().Set("fileSearch", fileSearch())
	SendFile()
	<-ch
}
