// Package main is the entry point for the GitHub MCP Server.
// It initializes the server and starts listening for MCP protocol requests.
package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/github/github-mcp-server/pkg/server"
	"github.com/spf13/cobra"
)

var (
	// Version is set at build time via ldflags.
	Version = "dev"
	// Commit is set at build time via ldflags.
	Commit = "none"
)

func main() {
	if err := rootCmd().Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func rootCmd() *cobra.Command {
	var (
		token   string
		host    string
		logFile string
	)

	cmd := &cobra.Command{
		Use:   "github-mcp-server",
		Short: "GitHub MCP Server — exposes GitHub APIs as MCP tools",
		Long: `github-mcp-server implements the Model Context Protocol (MCP)
and exposes GitHub REST and GraphQL APIs as callable tools for
AI assistants and agents.`,
		Version:      fmt.Sprintf("%s (%s)", Version, Commit),
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runServer(cmd.Context(), token, host, logFile)
		},
	}

	cmd.Flags().StringVarP(&token, "token", "t", "",
		"GitHub personal access token (overrides GITHUB_TOKEN env var)")
	cmd.Flags().StringVar(&host, "host", "https://api.github.com",
		"GitHub API base URL (useful for GitHub Enterprise Server)")
	cmd.Flags().StringVar(&logFile, "log-file", "",
		"Path to write structured JSON logs (defaults to stderr)")

	return cmd
}

// runServer resolves configuration, builds the MCP server, and blocks until
// the process receives SIGINT or SIGTERM.
func runServer(ctx context.Context, token, host, logFile string) error {
	// Prefer explicit flag value; fall back to environment variable.
	if token == "" {
		token = os.Getenv("GITHUB_TOKEN")
	}
	if token == "" {
		return fmt.Errorf("a GitHub token is required: set --token or GITHUB_TOKEN")
	}

	cfg := server.Config{
		Token:   token,
		Host:    host,
		LogFile: logFile,
		Version: Version,
	}

	srv, err := server.New(cfg)
	if err != nil {
		return fmt.Errorf("failed to create server: %w", err)
	}

	// Handle graceful shutdown on interrupt signals.
	ctx, stop := signal.NotifyContext(ctx, syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	fmt.Fprintf(os.Stderr, "github-mcp-server %s starting (stdio transport)\n", Version)

	if err := srv.Serve(ctx); err != nil {
		return fmt.Errorf("server error: %w", err)
	}

	return nil
}
