package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
)

type devProcess struct {
	cmd  *exec.Cmd
	mu   sync.Mutex
	done chan struct{}
}

func (p *devProcess) start(bin string) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	cmd := exec.Command(bin)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Env = os.Environ()

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("start: %w", err)
	}
	p.cmd = cmd
	p.done = make(chan struct{})

	go func() {
		cmd.Wait()
		close(p.done)
	}()

	return nil
}

func (p *devProcess) stop() {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.cmd == nil {
		return
	}

	if p.cmd.Process != nil {
		p.cmd.Process.Kill()
	}

	select {
	case <-p.done:
	case <-time.After(5 * time.Second):
	}

	p.cmd = nil
}

func runDev() {
	bin := "/tmp/cosmoria-dev"

	log.Println("building initial binary...")
	if err := buildDev(bin); err != nil {
		log.Fatalf("build: %v", err)
	}

	proc := &devProcess{}
	if err := proc.start(bin); err != nil {
		log.Fatalf("start: %v", err)
	}

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatalf("watcher: %v", err)
	}
	defer watcher.Close()

	if err := watchDirs(".", watcher); err != nil {
		log.Fatalf("watch dirs: %v", err)
	}

	debounce := time.NewTimer(0)
	if !debounce.Stop() {
		<-debounce.C
	}

	for {
		select {
		case ev := <-watcher.Events:
			if !strings.HasSuffix(ev.Name, ".go") {
				continue
			}
			debounce.Reset(150 * time.Millisecond)

		case <-debounce.C:
			log.Println("🔄 change detected, rebuilding...")
			proc.stop()

			if err := buildDev(bin); err != nil {
				log.Printf("❌ build failed: %v", err)
				if err := proc.start(bin); err != nil {
					log.Printf("restart failed: %v", err)
				}
				continue
			}

			if err := proc.start(bin); err != nil {
				log.Fatalf("restart: %v", err)
			}
			log.Println("✅ relay")

		case err := <-watcher.Errors:
			log.Printf("watch error: %v", err)
		}
	}
}

func buildDev(bin string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "go", "build", "-o", bin, "./cmd/cosmoria/")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("go build: %w", err)
	}
	return nil
}

func watchDirs(root string, watcher *fsnotify.Watcher) error {
	return filepath.Walk(root, func(path string, fi os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !fi.IsDir() {
			return nil
		}
		name := fi.Name()
		if name == ".git" || name == "vendor" || strings.HasPrefix(name, ".") {
			return filepath.SkipDir
		}
		return watcher.Add(path)
	})
}
