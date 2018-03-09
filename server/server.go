package server

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/adamstruck/ebsmount/ebsmount"
)

type UnmountRequest struct {
	VolumeID   string
	MountPoint string
}

func (m *UnmountRequest) Validate() error {
	validationErrs := []string{}
	if m.VolumeID == "" {
		validationErrs = append(validationErrs, "VolumeID not set")
	}
	if m.MountPoint == "" {
		validationErrs = append(validationErrs, "MountPoint not set")
	}
	if len(validationErrs) > 0 {
		return fmt.Errorf(strings.Join(validationErrs, "\n"))
	}
	return nil
}

func Run(port string) error {
	mounter, err := ebsmount.NewEC2Mounter()
	if err != nil {
		return err
	}

	mux := http.NewServeMux()

	mux.HandleFunc("/mount", func(w http.ResponseWriter, r *http.Request) {
		var mountReq ebsmount.MountRequest
		err := json.NewDecoder(r.Body).Decode(&mountReq)
		if err == io.EOF {
			http.Error(w, "Request body empty", 400)
			return
		}
		if err != nil {
			http.Error(w, err.Error(), 400)
			return
		}
		err = mountReq.Validate()
		if err != nil {
			err = fmt.Errorf("Request validation failed:\n%s", err)
			http.Error(w, err.Error(), 400)
			return
		}
		resp, err := mounter.CreateAndMount(&mountReq)
		if err != nil {
			log.Println(err.Error())
			http.Error(w, err.Error(), 500)
			return
		}
		json.NewEncoder(w).Encode(resp)
		return
	})

	mux.HandleFunc("/unmount", func(w http.ResponseWriter, r *http.Request) {
		var unmountReq UnmountRequest
		err := json.NewDecoder(r.Body).Decode(&unmountReq)
		if err == io.EOF {
			http.Error(w, "Request body empty", 400)
			return
		}
		if err != nil {
			http.Error(w, err.Error(), 400)
			return
		}
		err = unmountReq.Validate()
		if err != nil {
			err = fmt.Errorf("Request validation failed:\n%s", err)
			http.Error(w, err.Error(), 400)
			return
		}
		err = mounter.DetachAndDelete(unmountReq.VolumeID, unmountReq.MountPoint)
		if err != nil {
			log.Println(err.Error())
			http.Error(w, err.Error(), 500)
			return
		}
		return
	})

	log.Println("listening on port", port)
	return http.ListenAndServe(":"+port, mux)
}
