import "../../build/vendor/wasm_exec.js"
import "../../build/vendor/htmx.min.js"

const goWasm = new Go()


WebAssembly.instantiateStreaming(fetch("/assets/main.wasm"), goWasm.importObject).then((result) => {
    goWasm.run(result.instance)

    document.getElementById("fileInput").addEventListener("input", function(){
        document.getElementById("fileInput").files[0].arrayBuffer().then(function(x){
            const data = new Uint8Array(x[0])

            // sendFile(data)
        })
    })

    document.getElementById("fileSearch").addEventListener("change", function(){
        const text = document.getElementById("fileSearch").value

        fileSearch(text)
    })
})