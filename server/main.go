package main

import (
	"fmt"
	"log"
	"os/exec"
	"runtime"
	"time"

	"github.com/xavier268/mychess/game"
)

func main() {
	fmt.Printf("Memory model : %dM, loading cache ...\n", game.ZSize/1_000_000)
	s := newServer()
	const url = "http://localhost:8080"
	fmt.Printf("Serveur chess sur %s\n", url)
	go func() {
		time.Sleep(200 * time.Millisecond)
		openFirefox(url)
	}()
	if err := s.run(":8080"); err != nil {
		log.Fatal(err)
	}
	fmt.Println("Serveur arrêté.")
}

func openFirefox(url string) {
	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.Command("cmd", "/c", "start", "firefox", url)
	} else {
		cmd = exec.Command("firefox", url)
	}
	if err := cmd.Start(); err != nil {
		fmt.Println("Impossible d'ouvrir Firefox :", err)
	}
}
