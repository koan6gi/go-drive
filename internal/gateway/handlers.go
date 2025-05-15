package gateway

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/koan6gi/go-drive/internal/repository"
	repErr "github.com/koan6gi/go-drive/internal/repository/errors"
)

const (
	maxFileSize = 100 << 20
)

// POST /upload
func Upload(w http.ResponseWriter, r *http.Request) {
	err := r.ParseMultipartForm(maxFileSize)
	if err != nil {
		http.Error(w, fmt.Sprintf("%s: incorrect form", http.StatusText(http.StatusBadRequest)), http.StatusBadRequest)
		return
	}

	formFile, handler, err := r.FormFile("file")
	if err != nil {
		http.Error(w, fmt.Sprintf("%s: can't get a file", http.StatusText(http.StatusBadRequest)), http.StatusBadRequest)
	}
	defer formFile.Close()

	filePath := r.URL.Query().Get("path") + "/" + handler.Filename

	repository.FileStorage.Lock()
	defer repository.FileStorage.Unlock()

	newFile, err := repository.FileStorage.CreateFile(filePath)
	if err != nil {
		switch e := err.(type) {
		case *repErr.PathError:
			http.Error(w, fmt.Sprintf("%s: %s", http.StatusText(http.StatusBadRequest), e.Error()), http.StatusBadRequest)
		default:
			http.Error(w, fmt.Sprintf("%s: %s", http.StatusText(http.StatusInternalServerError), e.Error()), http.StatusInternalServerError)
		}
		return
	}
	defer newFile.Close()

	if _, err := io.Copy(newFile, formFile); err != nil {
		http.Error(w, fmt.Sprintf("%s: %s", http.StatusText(http.StatusInternalServerError), err.Error()), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "file upload success")
}

// GET /download
func Download(w http.ResponseWriter, r *http.Request) {
	filePath := r.URL.Query().Get("path")

	repository.FileStorage.Lock()
	defer repository.FileStorage.Unlock()

	file, err := repository.FileStorage.GetFile(filePath)
	if err != nil {
		switch e := err.(type) {
		case *repErr.PathError:
			http.Error(w, fmt.Sprintf("%s: %s", http.StatusText(http.StatusBadRequest), e.Error()), http.StatusBadRequest)
		default:
			http.Error(w, fmt.Sprintf("%s: %s", http.StatusText(http.StatusInternalServerError), e.Error()), http.StatusInternalServerError)
		}
		return
	}
	defer file.Close()

	fileInfo, err := file.Stat()
	if err != nil {
		http.Error(w, fmt.Sprintf("%s: %s", http.StatusText(http.StatusInternalServerError), err.Error()), http.StatusInternalServerError)
	}

	w.Header().Set("Content-Disposition", "attachment; filename="+fileInfo.Name())
	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Accept-Ranges", "bytes")

	http.ServeContent(w, r, fileInfo.Name(), fileInfo.ModTime(), file)
}

// POST /directory
func CreateDirectory(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Query().Get("path")

	repository.FileStorage.Lock()
	defer repository.FileStorage.Unlock()

	err := repository.FileStorage.CreateDirectory(path)
	if err != nil {
		switch e := err.(type) {
		case *repErr.PathError:
			http.Error(w, fmt.Sprintf("%s: %s", http.StatusText(http.StatusBadRequest), e.Error()), http.StatusBadRequest)
		default:
			http.Error(w, fmt.Sprintf("%s: %s", http.StatusText(http.StatusInternalServerError), e.Error()), http.StatusInternalServerError)
		}
		return
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "create directory success")
}

// DELETE /delete
func Delete(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Query().Get("path")

	repository.FileStorage.Lock()
	defer repository.FileStorage.Unlock()

	err := repository.FileStorage.Delete(path)
	if err != nil {
		switch e := err.(type) {
		case *repErr.PathError:
			http.Error(w, fmt.Sprintf("%s: %s", http.StatusText(http.StatusBadRequest), e.Error()), http.StatusBadRequest)
		default:
			http.Error(w, fmt.Sprintf("%s: %s", http.StatusText(http.StatusInternalServerError), e.Error()), http.StatusInternalServerError)
		}
		return
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "delete success")
}

// GET /list
func List(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Query().Get("path")

	repository.FileStorage.Lock()
	defer repository.FileStorage.Unlock()

	list, err := repository.FileStorage.List(path)
	if err != nil {
		switch e := err.(type) {
		case *repErr.PathError:
			http.Error(w, fmt.Sprintf("%s: %s", http.StatusText(http.StatusBadRequest), e.Error()), http.StatusBadRequest)
		default:
			http.Error(w, fmt.Sprintf("%s: %s", http.StatusText(http.StatusInternalServerError), e.Error()), http.StatusInternalServerError)
		}
		return
	}

	data, err := json.Marshal(list)
	if err != nil {
		http.Error(w, fmt.Sprintf("%s: %s", http.StatusText(http.StatusInternalServerError), err.Error()), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	_, err = w.Write(data)
	if err != nil {
		http.Error(w, fmt.Sprintf("%s: %s", http.StatusText(http.StatusInternalServerError), err.Error()), http.StatusInternalServerError)
		return
	}
}

// PUT /move
func Move(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	dest := query.Get("dest")
	src := query.Get("src")

	repository.FileStorage.Lock()
	defer repository.FileStorage.Unlock()

	err := repository.FileStorage.Move(dest, src)
	if err != nil {
		switch e := err.(type) {
		case *repErr.PathError:
			http.Error(w, fmt.Sprintf("%s: %s", http.StatusText(http.StatusBadRequest), e.Error()), http.StatusBadRequest)
		default:
			http.Error(w, fmt.Sprintf("%s: %s", http.StatusText(http.StatusInternalServerError), e.Error()), http.StatusInternalServerError)
		}
		return
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "move success")
}

// PUT /update
func Update(w http.ResponseWriter, r *http.Request) {
	err := r.ParseMultipartForm(maxFileSize)
	if err != nil {
		http.Error(w, fmt.Sprintf("%s: incorrect form", http.StatusText(http.StatusBadRequest)), http.StatusBadRequest)
		return
	}

	formFile, handler, err := r.FormFile("file")
	if err != nil {
		http.Error(w, fmt.Sprintf("%s: can't get a file", http.StatusText(http.StatusBadRequest)), http.StatusBadRequest)
	}
	defer formFile.Close()

	filePath := r.URL.Query().Get("path") + "/" + handler.Filename

	repository.FileStorage.Lock()
	defer repository.FileStorage.Unlock()

	newFile, err := repository.FileStorage.GetFile(filePath)
	if err != nil {
		switch e := err.(type) {
		case *repErr.PathError:
			http.Error(w, fmt.Sprintf("%s: %s", http.StatusText(http.StatusBadRequest), e.Error()), http.StatusBadRequest)
		default:
			http.Error(w, fmt.Sprintf("%s: %s", http.StatusText(http.StatusInternalServerError), e.Error()), http.StatusInternalServerError)
		}
		return
	}
	defer newFile.Close()

	if _, err := io.Copy(newFile, formFile); err != nil {
		http.Error(w, fmt.Sprintf("%s: %s", http.StatusText(http.StatusInternalServerError), err.Error()), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "file update success")
}

// PUT /copy
func Copy(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	dest := query.Get("dest")
	src := query.Get("src")

	repository.FileStorage.Lock()
	defer repository.FileStorage.Unlock()

	err := repository.FileStorage.Copy(dest, src)
	if err != nil {
		switch e := err.(type) {
		case *repErr.PathError:
			http.Error(w, fmt.Sprintf("%s: %s", http.StatusText(http.StatusBadRequest), e.Error()), http.StatusBadRequest)
		default:
			http.Error(w, fmt.Sprintf("%s: %s", http.StatusText(http.StatusInternalServerError), e.Error()), http.StatusInternalServerError)
		}
		return
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "copy success")
}
