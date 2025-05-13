package gateway

import "net/http"

// POST /upload?dest
func Upload(w http.ResponseWriter, r *http.Request) {

}

// GET /download?src
func Download(w http.ResponseWriter, r *http.Request) {

}

// POST /directory?dest
func CreateDirectory(w http.ResponseWriter, r *http.Request) {

}

// DELETE /delete?src
func Delete(w http.ResponseWriter, r *http.Request) {

}

// GET /list?src
func List(w http.ResponseWriter, r *http.Request) {

}

// PUT /move?dest,src
func Move(w http.ResponseWriter, r *http.Request) {

}

// PUT /update?src
func Update(w http.ResponseWriter, r *http.Request) {

}

// PUT /copy?dest,src
func Copy(w http.ResponseWriter, r *http.Request) {

}
