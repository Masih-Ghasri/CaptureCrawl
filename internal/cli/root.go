package cli

import (
	"time"

	"github.com/Masih-Ghasri/CaptureCrawl/internal/model"
	"github.com/spf13/cobra"
)

var cfg = &model.Config{}

var rootCmd = &cobra.Command{
	Use:           "capturecrawl <url>",
	Short:         "Crawl a website and export its pages as Markdown documentation",
	Args:          cobra.ExactArgs(1),
	SilenceUsage:  true,
	RunE:          run,
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	flags := rootCmd.Flags()
	flags.StringVarP(&cfg.OutFile, "out", "o", "", "output file (default <host>-docs.md)")
	flags.IntVar(&cfg.MaxPages, "max-pages", 500, "maximum number of pages to crawl")
	flags.IntVar(&cfg.MaxDepth, "max-depth", 0, "maximum link depth from the start page (0 = unlimited)")
	flags.IntVar(&cfg.Workers, "workers", 4, "number of concurrent browser tabs")
	flags.DurationVar(&cfg.Delay, "delay", 500*time.Millisecond, "delay between requests to the same host")
	flags.DurationVar(&cfg.PageTimeout, "timeout", 15*time.Second, "render timeout per page")
	flags.BoolVar(&cfg.IgnoreRobots, "ignore-robots", false, "ignore robots.txt restrictions")
	flags.BoolVar(&cfg.DumpJSON, "dump-json", false, "also write raw crawl data next to the output file")
}
