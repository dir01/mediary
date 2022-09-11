package http

import (
	"encoding/json"
	"net/http"
	"os"
)

func handleDocs() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		// read README.md
		bytes, err := os.ReadFile("README.md")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		_, err = json.Marshal(string(bytes))
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		payload := `<html>
		<head>
			<title>mediary</title>
			<meta charset="utf-8">
			<script src="https://unpkg.com/showdown/dist/showdown.min.js"></script>
		</head>
		<body>
			<div id="root"></div>
			<textarea id="markdown" style="display:none">` + string(bytes) + `
			</textarea>
			<script>
				window.__payload__ = document.getElementById('markdown').value
				console.log(window.__payload__);
				document.getElementById('root').innerHTML = new showdown.Converter().makeHtml(window.__payload__)
			</script>
		</body>
</html>
`
		respond(w, http.StatusOK, payload)
	}
}
