package gateway

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/koan6gi/go-drive/internal/repository"
	repErr "github.com/koan6gi/go-drive/internal/repository/errors"
)

const (
	maxFileSize = 100 << 20
)

// Upload godoc
// @Summary Upload file
// @Description Upload a file to the specified path
// @Tags Files
// @Accept multipart/form-data
// @Produce plain
// @Param file formData file true "File to upload"
// @Param path query string true "Destination path"
// @Success 200 {string} string "file upload success"
// @Failure 400 {string} string "Bad Request"
// @Failure 500 {string} string "Internal Server Error"
// @Router /upload [post]
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
	if strings.HasPrefix(filePath, "//") {
		filePath = filePath[1:]
	}

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

// Download godoc
// @Summary Download file
// @Description Download file from specified path
// @Tags Files
// @Produce octet-stream
// @Param path query string true "File path to download"
// @Success 200 {file} binary "File content"
// @Failure 400 {string} string "Bad Request"
// @Failure 500 {string} string "Internal Server Error"
// @Router /download [get]
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

// CreateDirectory godoc
// @Summary Create directory
// @Description Create new directory at specified path
// @Tags Directories
// @Produce plain
// @Param path query string true "Directory path to create"
// @Success 200 {string} string "create directory success"
// @Failure 400 {string} string "Bad Request"
// @Failure 500 {string} string "Internal Server Error"
// @Router /directory [post]
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

// Delete godoc
// @Summary Delete file/directory
// @Description Delete file or directory at specified path
// @Tags Files
// @Produce plain
// @Param path query string true "Path to delete"
// @Success 200 {string} string "delete success"
// @Failure 400 {string} string "Bad Request"
// @Failure 500 {string} string "Internal Server Error"
// @Router /delete [delete]
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

// List godoc
// @Summary List directory contents
// @Description Get list of files/directories in specified path
// @Tags Directories
// @Produce json
// @Param path query string true "Directory path to list"
// @Success 200 {array} string "List of files/directories"
// @Failure 400 {string} string "Bad Request"
// @Failure 500 {string} string "Internal Server Error"
// @Router /list [get]
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

// Move godoc
// @Summary Move file/directory
// @Description Move file or directory from source to destination
// @Tags Files
// @Produce plain
// @Param src query string true "Source path"
// @Param dest query string true "Destination path"
// @Success 200 {string} string "move success"
// @Failure 400 {string} string "Bad Request"
// @Failure 500 {string} string "Internal Server Error"
// @Router /move [put]
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

// Update godoc
// @Summary Update file
// @Description Update existing file content
// @Tags Files
// @Accept multipart/form-data
// @Produce plain
// @Param file formData file true "New file content"
// @Param path query string true "File path to update"
// @Success 200 {string} string "file update success"
// @Failure 400 {string} string "Bad Request"
// @Failure 500 {string} string "Internal Server Error"
// @Router /update [put]
func Update(w http.ResponseWriter, r *http.Request) {
	err := r.ParseMultipartForm(maxFileSize)
	if err != nil {
		http.Error(w, fmt.Sprintf("%s: incorrect form", http.StatusText(http.StatusBadRequest)), http.StatusBadRequest)
		return
	}

	formFile, _, err := r.FormFile("file")
	if err != nil {
		http.Error(w, fmt.Sprintf("%s: can't get a file", http.StatusText(http.StatusBadRequest)), http.StatusBadRequest)
	}
	defer formFile.Close()

	filePath := r.URL.Query().Get("path")

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

// Copy godoc
// @Summary Copy file/directory
// @Description Copy file or directory from source to destination
// @Tags Files
// @Produce plain
// @Param src query string true "Source path"
// @Param dest query string true "Destination path"
// @Success 200 {string} string "copy success"
// @Failure 400 {string} string "Bad Request"
// @Failure 500 {string} string "Internal Server Error"
// @Router /copy [put]
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
