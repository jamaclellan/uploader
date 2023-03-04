package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"uploader"

	"gopkg.in/yaml.v3"
)

var (
	configPath   = flag.String("cfg", "./uploader.yaml", "Uploader config path.")
	registerName = flag.String("register", "", "Username to register.")
	host         = flag.String("addr", "[::1]", "Address to listen on.")
	port         = flag.Int("port", 8080, "Port to listen on.")
)

func main() {
	flag.Parse()

	cfg := loadConfig(*configPath)
	if *registerName != "" {
		registerUser(cfg, *registerName)
		return
	}
	runServer(cfg)
}

func runServer(cfg *uploader.Config) {
	up := uploaderFromCfg(cfg)
	defer up.Close()
	server := &http.Server{
		ReadHeaderTimeout: 30 * time.Second,
		IdleTimeout:       30 * time.Second,
		Handler:           up,
		Addr:              fmt.Sprintf("%s:%d", *host, *port),
	}
	if err := server.ListenAndServe(); err != nil {
		log.Fatalf("error during runtime of server: %q", err)
	}
}

func loadConfig(path string) *uploader.Config {
	contents, err := os.ReadFile(path)
	if err != nil {
		log.Fatalf("Failed reading config file: %s", err)
	}
	cfg := &uploader.Config{}
	if err = yaml.Unmarshal(contents, cfg); err != nil {
		log.Fatalf("Failed parsing config ifle: %s", err)
	}
	return cfg
}

func uploaderFromCfg(cfg *uploader.Config) *uploader.Uploader {
	up, err := uploader.NewUploaderFromConfig(cfg)
	if err != nil {
		log.Fatalf("Failed to start server: %s", err)
	}
	return up
}

func registerUser(cfg *uploader.Config, name string) {
	up := uploaderFromCfg(cfg)
	defer up.Close()
	user, err := up.Auth.UserRegister(name)
	if err != nil {
		log.Fatalf("Failed to register user: %s", err)
	}
	log.Println(up.UploadScript(user))
}
