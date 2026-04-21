package cmd

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	"github.com/jquiaios/worklog/internal/db"
	"github.com/jquiaios/worklog/internal/server"
	"github.com/spf13/cobra"
)

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start the web UI at http://localhost:7171",
	RunE: func(cmd *cobra.Command, args []string) error {
		port, _   := cmd.Flags().GetInt("port")
		host, _   := cmd.Flags().GetString("host")
		noOpen, _ := cmd.Flags().GetBool("no-open")

		store, err := db.Open(dbPath())
		if err != nil {
			return fmt.Errorf("opening database: %w", err)
		}
		defer store.Close()

		addr := fmt.Sprintf("%s:%d", host, port)
		url  := fmt.Sprintf("http://localhost:%d", port)

		httpServer := &http.Server{
			Addr:    addr,
			Handler: server.New(store),
		}

		fmt.Fprintf(os.Stderr, "worklog serving at %s\n", url)

		quit := make(chan os.Signal, 1)
		signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
		go func() {
			<-quit
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			httpServer.Shutdown(ctx)
		}()

		if !noOpen {
			go func() {
				time.Sleep(100 * time.Millisecond)
				openBrowser(url)
			}()
		}

		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			return err
		}
		fmt.Fprintln(os.Stderr, "bye")
		return nil
	},
}

func openBrowser(url string) {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("cmd", "/c", "start", url)
	case "darwin":
		cmd = exec.Command("open", url)
	default:
		cmd = exec.Command("xdg-open", url)
	}
	cmd.Start()
}

func init() {
	serveCmd.Flags().IntP("port", "p", 7171, "port to listen on")
	serveCmd.Flags().String("host", "127.0.0.1", "host to bind to (use 0.0.0.0 inside Docker)")
	serveCmd.Flags().Bool("no-open", false, "do not open the browser automatically")
	rootCmd.AddCommand(serveCmd)
}
