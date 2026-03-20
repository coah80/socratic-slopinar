package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os/exec"
	"path/filepath"
	"runtime"

	"github.com/coah80/socratic-slopinar/internal/api"
)

func main() {
	dev := flag.Bool("dev", false, "dev mode - proxy to frontend dev server instead of embedded files")
	port := flag.String("port", "8080", "port to listen on")
	flag.Parse()

	addr := "localhost:" + *port

	var frontend http.Handler
	if !*dev {
		exePath, err := filepath.Abs(".")
		if err != nil {
			log.Fatalf("failed to get working directory: %v", err)
		}
		buildDir := filepath.Join(exePath, "frontend", "build")
		frontend = http.FileServer(http.Dir(buildDir))
	}

	router := api.NewRouter(frontend)

	fmt.Printf("Socratic Slopinar running at http://%s\n", addr)
	openBrowser("http://" + addr)

	if err := http.ListenAndServe(addr, router); err != nil {
		log.Fatalf("server error: %v", err)
	}
}

func openBrowser(url string) {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", url)
	case "linux":
		cmd = exec.Command("xdg-open", url)
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
	default:
		return
	}
	_ = cmd.Start()
}
