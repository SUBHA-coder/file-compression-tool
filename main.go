package main

import (
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/nfnt/resize"
	"github.com/pdfcpu/pdfcpu/pkg/api"
)

const uploadDir = "./uploads"
const compressedDir = "./uploads/compressed"

// Create directories if they do not exist
func init() {
	// Create the uploads and compressed directories if they don't exist
	if err := os.MkdirAll(uploadDir, os.ModePerm); err != nil {
		log.Fatal("Error creating upload directory:", err)
	}
	if err := os.MkdirAll(compressedDir, os.ModePerm); err != nil {
		log.Fatal("Error creating compressed directory:", err)
	}
}

func main() {
	// Serve static files (HTML, CSS, JS)
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	// Handle the file upload and compression
	http.HandleFunc("/compress", compressFileHandler)

	// Serve the HTML file
	http.HandleFunc("/", serveIndex)

	// Start the web server
	log.Println("Starting server on :8080...")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal("Error starting server:", err)
	}
}

// Serve the index.html page
func serveIndex(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "./static/index.html")
}

// Compress file handler - handles file upload and compression
func compressFileHandler(w http.ResponseWriter, r *http.Request) {
	// Limit file size to 10 MB
	r.ParseMultipartForm(10 << 20)

	// Get the uploaded file
	file, _, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "Error uploading file: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer file.Close()

	// Create a temporary file to save the uploaded content
	tmpFile, err := ioutil.TempFile(uploadDir, "uploaded-*.tmp")
	if err != nil {
		http.Error(w, "Error saving uploaded file: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer tmpFile.Close()

	// Copy the uploaded file to the temporary file
	_, err = io.Copy(tmpFile, file)
	if err != nil {
		http.Error(w, "Error writing uploaded file: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Get the file extension and call appropriate compression function
	fileName := tmpFile.Name()
	ext := strings.ToLower(filepath.Ext(fileName))

	// Check file type and compress accordingly
	var compressedFilePath string
	if ext == ".jpg" || ext == ".jpeg" || ext == ".png" {
		compressedFilePath = compressImage(fileName, ext)
	} else if ext == ".pdf" {
		compressedFilePath = compressPDF(fileName)
	} else {
		http.Error(w, "Unsupported file type", http.StatusBadRequest)
		return
	}

	// Respond with the compressed file URL
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, `{"message": "File compressed successfully!", "file": "/uploads/compressed/%s"}`, filepath.Base(compressedFilePath))
}

// Compress an image file (JPG, PNG)
func compressImage(filePath string, ext string) string {
	// Open the image file
	imgFile, err := os.Open(filePath)
	if err != nil {
		log.Println("Error opening image:", err)
		return ""
	}
	defer imgFile.Close()

	// Decode the image
	var img image.Image
	if ext == ".jpg" || ext == ".jpeg" {
		img, err = jpeg.Decode(imgFile)
	} else if ext == ".png" {
		img, err = png.Decode(imgFile)
	}
	if err != nil {
		log.Println("Error decoding image:", err)
		return ""
	}

	// Resize the image to a max width of 800px while maintaining aspect ratio
	resizedImg := resize.Resize(800, 0, img, resize.Lanczos3)

	// Create a new file to save the compressed image
	compressedFilePath := filepath.Join(compressedDir, filepath.Base(filePath))
	outFile, err := os.Create(compressedFilePath)
	if err != nil {
		log.Println("Error creating compressed image:", err)
		return ""
	}
	defer outFile.Close()

	// Save the resized image as a JPG file
	err = jpeg.Encode(outFile, resizedImg, nil)
	if err != nil {
		log.Println("Error encoding compressed image:", err)
		return ""
	}

	return compressedFilePath
}

// Compress a PDF file
func compressPDF(filePath string) string {
	// Create the destination path for the compressed PDF
	compressedFilePath := filepath.Join(compressedDir, filepath.Base(filePath))

	// Use pdfcpu to compress the PDF
	err := api.OptimizeFile(filePath, compressedFilePath, nil)
	if err != nil {
		log.Println("Error compressing PDF:", err)
		return ""
	}

	return compressedFilePath
}
