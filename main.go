package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/gorilla/mux"
	"github.com/mitchellh/mapstructure"
	"github.com/spf13/viper"
)

var cfg Config

func init() {
	if err := loadConfig(&cfg, "env/config.yaml", "yaml"); err != nil {
		log.Fatalf("error reading configuration: %v", err)
	}
}

func main() {
	r := mux.NewRouter()

	r.HandleFunc("/", handler)
	r.HandleFunc("/health", healthHandler)
	r.HandleFunc("/readiness", readinessHandler)

	srv := &http.Server{
		Handler:      r,
		Addr:         ":8080",
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	// Start Server
	go func() {
		log.Println("Starting Server")
		if err := srv.ListenAndServe(); err != nil {
			log.Fatal(err)
		}
	}()

	interruptChan := make(chan os.Signal, 1)
	signal.Notify(interruptChan, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	// Block until we receive our signal.
	<-interruptChan

	// create a deadline to wait for.
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	srv.Shutdown(ctx)

	log.Println("Shutting down")
	os.Exit(0)
}

type Config struct {
	Runtime Runtime `mapstructure:"runtime"`
	Name    string  `mapstructure:"name"`
}

type Runtime struct {
	Environment string `mapstructure:"environment"`
}

func handler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte(fmt.Sprintf("Config: %s\n", cfg.Name)))
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func readinessHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func loadConfig(cfgPtr interface{}, address, format string) error {
	{ // first we need to read the viper from file, no need to replace envs, that is next step
		fconfig := viper.New()
		fconfig.SetConfigFile(address)
		fconfig.SetConfigType(format)
		if err := fconfig.ReadInConfig(); err != nil {
			return fmt.Errorf("error loading configuration: %w", err)
		}

		if err := fconfig.UnmarshalExact(cfgPtr); err != nil {
			return fmt.Errorf("error parsing configuration: %w", err)
		}
	}
	// then, we need to re-export it, so we get all the keys
	expMap := make(map[string]interface{})

	err := mapstructure.Decode(cfgPtr, &expMap)
	if err != nil {
		// this has almost no chance of hapenning
		return fmt.Errorf("error decoding mapstructure: %w", err)
	}
	expJson, err := json.Marshal(expMap)
	if err != nil {
		// this has almost no chance of hapenning
		return fmt.Errorf("error re-marshaling json: %w", err)
	}

	// then, we need to re-read it in viper, *with all the keys now*, and replacer and automaticenv
	vconfig := viper.NewWithOptions(
		viper.EnvKeyReplacer(
			strings.NewReplacer(".", "_"),
		),
	)
	vconfig.AutomaticEnv()
	vconfig.SetConfigType("json")
	vconfig.ReadConfig(bytes.NewReader(expJson))

	if err := vconfig.UnmarshalExact(cfgPtr); err != nil {
		// this has almost no chance of hapenning
		return fmt.Errorf("error parsing configuration: %v", err)
	}
	return nil
}
