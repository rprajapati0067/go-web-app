package main

import (
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
)

var imdb map[int]interface{}
var tpl = template.Must(template.ParseFiles("index.html"))
var MAX_UPLOAD_SIZE = 8e+6
var i int

func indexHandler(w http.ResponseWriter, r *http.Request) {
	tpl.Execute(w, nil)
}

func init() {
	imdb = dbinit()
}

func main() {

	err := godotenv.Load()
	if err != nil {
		log.Println("Error loading .env file")
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}

	fs := http.FileServer(http.Dir("assets"))

	mux := http.NewServeMux()
	mux.Handle("/assets/", http.StripPrefix("/assets/", fs))
	log.Println("Starting web app...")
	mux.HandleFunc("/upload", uploadFileHandler)
	mux.HandleFunc("/", indexHandler)
	http.ListenAndServe(":"+port, mux)

}

func uploadFileHandler(w http.ResponseWriter, r *http.Request) {

	auth := r.FormValue("auth")
	token := os.Getenv("TOKEN")

	if auth != token {
		w.Write([]byte("<h1>Token do not matched</h1>"))
		http.Error(w, "Token do not matched", http.StatusForbidden)
		return
	}

	log.Println("File Upload Endpoint Hit")

	if float64(r.ContentLength) > MAX_UPLOAD_SIZE {
		http.Error(w, "The uploaded image is too big. Please use an image less than 8MB in size", http.StatusBadRequest)
		return
	}

	// FormFile returns the first file for the given key `data`
	file, handler, err := r.FormFile("data")
	if err != nil {
		log.Println("Error Retrieving the File")
		log.Println(err)
		return
	}
	defer file.Close()
	log.Printf("Uploaded File: %+v\n", handler.Filename)
	log.Printf("File Size: %+v\n", handler.Size)
	log.Printf("MIME Header: %+v\n", handler.Header)
	log.Printf("MIME Header1: %+T\n", handler.Header["Content-Type"])
	//map[Content-Disposition:[form-data; name="data"; filename="driving-license.jpg"] Content-Type:[image/jpeg]]
	// Create a temporary file within our temp-images directory that follows
	// a particular naming pattern
	tempFile, err := ioutil.TempFile("temp-images", "upload-*.png")
	if err != nil {
		log.Println(err)
	}
	defer tempFile.Close()

	// read all of the contents of our uploaded file into a
	// byte array
	fileBytes, err := ioutil.ReadAll(file)
	if err != nil {
		log.Println(err)
	}

	// Now detect the type of File and throw error based on validation
	detectedFileType := http.DetectContentType(fileBytes)
	switch detectedFileType {
	case "image/jpeg", "image/jpg":
	case "image/gif", "image/png":
		break
	default:
		http.Error(w, "Please upload image file", http.StatusForbidden)
		return
	}

	// write this byte array to our temporary file
	tempFile.Write(fileBytes)
	// return that we have successfully uploaded our file!

	fileStruct := struct {
		Name        string
		Size        int64
		ContentType string
	}{
		Name:        handler.Filename,
		Size:        handler.Size,
		ContentType: handler.Header["Content-Type"][0],
	}
	log.Println(fileStruct)

	i++

	imdb[i] = fileStruct
	log.Println(imdb)

	w.Write([]byte("<h1>Successfully Uploaded File!</h1>"))
}
