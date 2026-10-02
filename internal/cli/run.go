package cli

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/spf13/cobra"
)

func run(cmd *cobra.Command, args []string) error {
	target, err := resolveTarget(args[0])
	if err != nil {
		return err
	}

	cfg.URL = target
	if cfg.OutFile == "" {
		cfg.OutFile = defaultOutput(target)
	}
	if cfg.Workers < 1 {
		return fmt.Errorf("--workers must be at least 1")
	}
	if cfg.MaxPages < 1 {
		return fmt.Errorf("--max-pages must be at least 1")
	}

	out := cmd.OutOrStdout()
	fmt.Fprintf(out, "target    %s\n", cfg.URL)
	fmt.Fprintf(out, "output    %s\n", cfg.OutFile)
	fmt.Fprintf(out, "crawl     %d pages max, depth %s, %d workers, delay %s\n",
		cfg.MaxPages, depthLabel(cfg.MaxDepth), cfg.Workers, cfg.Delay)
	if cfg.IgnoreRobots {
		fmt.Fprintf(out, "robots     ignored\n")
	}

	return nil
}

func resolveTarget(raw string) (string, error) {
	if !strings.Contains(raw, "://") {
		raw = "https://" + raw
	}

	u, err := url.Parse(raw)
	if err != nil {
		return "", fmt.Errorf("invalid url %q: %w", raw, err)
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return "", fmt.Errorf("unsupported scheme %q, use http or https", u.Scheme)
	}
	if u.Host == "" {
		return "", fmt.Errorf("url %q has no host", raw)
	}

	u.Fragment = ""
	return u.String(), nil
}

func defaultOutput(raw string) string {
	u, err := url.Parse(raw)
	if err != nil || u.Hostname() == "" {
		return "docs.md"
	}
	return u.Hostname() + "-docs.md"
}

func depthLabel(depth int) string {
	if depth <= 0 {
		return "unlimited"
	}
	return fmt.Sprintf("%d", depth)
}
