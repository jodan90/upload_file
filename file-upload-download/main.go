package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"os"
	"path/filepath"
)

var tmpl = template.Must(template.ParseFiles("templates/upload.html"))

func uploadPage(w http.ResponseWriter, r *http.Request) {
	tmpl.Execute(w, nil)
}

func uploadFile(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	r.ParseMultipartForm(10 << 20)

	file, handler, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "파일을 가져오는 중 오류 발생", http.StatusInternalServerError)
		return
	}
	defer file.Close()

	filePath := filepath.Join("uploads", handler.Filename)

	dst, err := os.Create(filePath)
	if err != nil {
		http.Error(w, "파일 생성 오류", http.StatusInternalServerError)
		return
	}
	defer dst.Close()

	if _, err := io.Copy(dst, file); err != nil {
		http.Error(w, "파일 저장 중 오류 발생", http.StatusInternalServerError)
		return
	}

	response := map[string]string{"message": "파일이 성공적으로 업로드되었습니다!", "filename": handler.Filename}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

func downloadFile(w http.ResponseWriter, r *http.Request) {
	fileName := r.URL.Query().Get("filename")
	if fileName == "" {
		http.Error(w, "Filename is required", http.StatusBadRequest)
		return
	}

	filePath := filepath.Join("uploads", fileName)

	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		http.Error(w, "File not found", http.StatusNotFound)
		return
	}

	file, err := os.Open(filePath)
	if err != nil {
		http.Error(w, "Unable to open file", http.StatusInternalServerError)
		return
	}
	defer file.Close()

	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", fileName))
	w.Header().Set("Content-Type", "application/octet-stream")
	http.ServeFile(w, r, filePath)
}

func main() {
	os.MkdirAll("uploads", os.ModePerm)

	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	http.HandleFunc("/", uploadPage)
	http.HandleFunc("/upload", uploadFile)
	http.HandleFunc("/download", downloadFile)

	fmt.Println("서버가 :8080에서 실행 중입니다...")
	http.ListenAndServe(":8080", nil)
}
