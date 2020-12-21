package gitd

import (
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
)

func uploadPackHandler(w http.ResponseWriter, r *http.Request) {
	execPath := "."

	cmd := exec.Command("git", "upload-pack", "--stateless-rpc", execPath)
	stdin, stdout, stderr, ok := GetChildPipes(cmd, w)
	if !ok {
		return
	}

	if err := cmd.Start(); err != nil {
		log.Println(err)
		http.Error(w, "Error while spawning", http.StatusInternalServerError)
		return
	}

	reqBody, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Println("Error while reading request body:", err)
		http.Error(w, "Error while reading request body", http.StatusInternalServerError)
		return
	}
	stdin.Write(reqBody)
    os.Stderr.Write(reqBody)

	contentType := "application/x-git-upload-pack-result"
	SetHeader(w, contentType)

	go io.Copy(w, stdout)
	go io.Copy(w, stderr)

	if err := cmd.Wait(); err != nil {
		log.Println("Error while waiting:", err)
		return
	}
}
