package server

import (
	"fmt"
	// "encoding/json"
	"net/http"
	// "github.com/adamstruck/ebsmount/ebsmount"
)

func Run(port string) error {
	// mounter, err  := ebsmount.NewEC2Mounter()
	// if err != nil {
	// 	return err
	// }

	http.HandleFunc("/mount", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "mount")
	})

	http.HandleFunc("/unmount", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "unmount")
	})

	return http.ListenAndServe(":"+port, nil)
}
