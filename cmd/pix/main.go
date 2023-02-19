package main

import (
	"encoding/json"
	"io"
	"net/http"
	"os"
	"strconv"
)

// cmd/pix/main.go
// this is a service that will increment views for a given tracking pixel
// the tracking pixel will be a single 1x1 transparent gif
// the tracking pixel will also have a query string parameter that will be the id of the pixel
// the id of the pixel will create a new file in the os filesystem
// the file will be named after the id of the pixel
// the file will be a text file that will contain the number of views for that pixel
// the file will be created if it does not exist
// the file will be incremented by 1 if it does exist
// the file will be created in the os filesystem
// the file will be created in the current working directory plus a folder called "data" plus a folder called "pixels" plus a folder called "id"
// the directory structure will be created if it does not exist
// the service will also expose an endpoint that will return the number of views for a given pixel
// it will also expose an endpoint that will return the number of views for all pixels
// it will return 0 if the pixel does not exist

func main() {

	// create the http server
	// create the router
	// create the routes
	// start the server
	os.MkdirAll("data/pixels", 0755)

	http.HandleFunc("/pix.gif", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "image/gif")
		// allow cross origin requests
		w.Header().Add("Access-Control-Allow-Origin", "*")

		// get the id from the query string
		id := r.URL.Query().Get("id")
		// create the file for the id
		_, err := os.Stat("data/pixels/" + id)
		if os.IsNotExist(err) {
			// create the file
			_, err := os.Create("data/pixels/" + id)
			if err != nil {
				panic(err)
			}
		} else if err != nil {
			panic(err)
		}

		// read the file for the id
		b, err := os.ReadFile("data/pixels/" + id)
		if err != nil {
			panic(err)
		}
		if string(b) == "" {
			b = []byte("0")
		}
		// convert the bytes to an int
		count, err := strconv.Atoi(string(b))
		if err != nil {
			panic(err)
		}

		// increment the int
		count++

		// convert the int to bytes
		b = []byte(strconv.Itoa(count))

		// write the bytes to the file
		err = os.WriteFile("data/pixels/"+id, b, 0644)
		if err != nil {
			panic(err)
		}

		// return the gif
		w.WriteHeader(http.StatusOK)
		// return the bytes of a transparent 1x1 gif
		b, err = os.ReadFile("pix.gif")
		if err != nil {
			panic(err)
		}
		w.Write(b)
	})

	http.HandleFunc("/view", func(w http.ResponseWriter, r *http.Request) {
		// allow cross origin requests
		w.Header().Add("Access-Control-Allow-Origin", "*")

		// set the content type
		w.Header().Add("Content-Type", "text/plain")

		// get the id from the query string
		id := r.URL.Query().Get("id")

		// open the file for the id
		_, err := os.Stat("data/pixels/" + id)
		if os.IsNotExist(err) {
			// return 0
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("0"))
			return
		} else if err != nil {
			panic(err)
		}

		// read the file for the id
		bytes, err := os.ReadFile("data/pixels/" + id)
		if err != nil && io.EOF.Error() == err.Error() {
			// set the bytes to 0
			bytes = []byte("0")
		} else if err != nil {
			panic(err)
		}

		// convert the bytes to an int
		count, err := strconv.Atoi(string(bytes))
		if err != nil {
			panic(err)
		}

		// return the int
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(strconv.Itoa(count)))

	})

	http.HandleFunc("/total", func(w http.ResponseWriter, r *http.Request) {
		// allow cross origin requests
		w.Header().Add("Access-Control-Allow-Origin", "*")

		// set the content type
		w.Header().Add("Content-Type", "text/plain")

		// open all the files in the directory
		files, err := os.ReadDir("data/pixels")
		if err != nil {
			panic(err)
		}

		counts := map[string]int{}
		for _, file := range files {
			// read the file for the id
			bytes, err := os.ReadFile("data/pixels/" + file.Name())
			if err != nil && io.EOF.Error() == err.Error() {
				// set the bytes to 0
				bytes = []byte("0")
			} else if err != nil {
				panic(err)
			}

			// convert the bytes to an int
			count, err := strconv.Atoi(string(bytes))
			if err != nil {
				panic(err)
			}
			// add counts to map
			counts[file.Name()] = count
			// add counts to total
			counts["total"] += count
		}

		data, err := json.Marshal(counts)
		if err != nil {
			panic(err)
		}

		// return the int
		w.WriteHeader(http.StatusOK)
		w.Write(data)

	})

	// handle root
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// allow cross origin requests
		w.Header().Add("Access-Control-Allow-Origin", "*")

		// set the content type
		w.Header().Add("Content-Type", "text/html")

		w.Write([]byte(`
		<!DOCTYPE html>
		<html>
			<head>
				<title>Pixel</title>
			</head>
			<body>
				<h1>Pixel</h1>
				<p>Use this pixel to track views of your content.</p>
				<p>Simply add the following to your HTML:</p>
				<pre>
					&lt;img src="https://pix.kfelter.com/pix.gif?id=your-id" width="1" height="1" /&gt;
				</pre>
				<p>Replace <code>id</code> with a unique identifier for your content.</p>

				<p>Then, to view the number of views for your content, use the following URL:</p>
				<pre>
					https://pix.kfelter.com/view?id=your-id
				</pre>
				<p>Replace <code>your-id</code> with the unique identifier you used in the pixel.</p>

				<p>Finally, to view the total number of views for all content, use the following URL:</p>
				<pre>
					https://pix.kfelter.com/total
				</pre>
			</body>
		</html>
		`))
	})

	err := http.ListenAndServeTLS(":443",
		"/etc/letsencrypt/live/pix.kfelter.com/fullchain.pem",
		"/etc/letsencrypt/live/pix.kfelter.com/privkey.pem",
		nil,
	)
	if err != nil {
		panic(err)
	}

}
