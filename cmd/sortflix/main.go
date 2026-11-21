package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"sortflix/internal/config"
	"sortflix/internal/organizer"
)

func main() {
	cfg, err := config.ParseFlags()
	if err != nil {
		fmt.Fprintf(os.Stderr, "❌ Configuration error: %v\n\n", err)
		os.Exit(1)
	}

	printHeader(cfg)

	org := organizer.New(cfg)

	fmt.Println("🔍 Scanning library...")
	plan, err := org.BuildPlan()
	if err != nil {
		fmt.Fprintf(os.Stderr, "❌ Error analyzing library: %v\n", err)
		os.Exit(1)
	}

	printPlanDetails(plan, cfg)
	printSummary(plan, cfg)

	// If no changes are needed
	if len(plan.Moves) == 0 && len(plan.FileDeletions) == 0 && len(plan.DirDeletions) == 0 {
		fmt.Println("\n🎉 Your movie library is already clean and organized. Nothing to do!")
		return
	}

	// Dry run mode stops here
	if cfg.DryRun {
		fmt.Println("\nℹ️  DRY-RUN MODE: No files or directories were modified.")
		fmt.Println("👉 To apply changes, run the command again without '-dry-run' or '-n'.")
		return
	}

	// Confirmation prompt if -y / -yes was not supplied
	if !cfg.Yes {
		if !askConfirmation() {
			fmt.Println("\n🚫 Operation cancelled. No changes were made.")
			return
		}
	}

	fmt.Println("\n🚀 Applying changes...")
	if err := org.Execute(plan); err != nil {
		fmt.Fprintf(os.Stderr, "\n❌ Error applying changes: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("\n✅ Done! Library organized and cleaned up successfully.")
}

func printHeader(cfg *config.Config) {
	fmt.Println("================================================================")
	fmt.Println("   🎬 SORTFLIX - Movie Library Organizer & Cleaner")
	fmt.Println("================================================================")
	fmt.Printf("📁 Root folder:    %s\n", cfg.Path)
	fmt.Printf("⚖️  Min video size: %d MB\n", cfg.MinSizeMB)
	if cfg.DryRun {
		fmt.Println("🛡️  Mode:           DRY-RUN (simulation) - Nothing will be modified")
	} else {
		fmt.Println("⚡ Mode:           LIVE EXECUTION")
	}
	fmt.Println("================================================================")
}

func printPlanDetails(plan *organizer.Plan, cfg *config.Config) {
	if len(plan.Moves) > 0 {
		fmt.Printf("\n📦 Movies to relocate into alphabetical folders (%d):\n", len(plan.Moves))
		maxShow := 15
		if cfg.Verbose || len(plan.Moves) <= maxShow {
			for _, m := range plan.Moves {
				fmt.Printf("  • %s  ➜  [%s] %s (%s)\n",
					rel(cfg.Path, m.Source),
					m.Category,
					rel(cfg.Path, m.Destination),
					organizer.FormatBytes(m.Size))
			}
		} else {
			for i := 0; i < maxShow; i++ {
				m := plan.Moves[i]
				fmt.Printf("  • %s  ➜  [%s] %s (%s)\n",
					rel(cfg.Path, m.Source),
					m.Category,
					rel(cfg.Path, m.Destination),
					organizer.FormatBytes(m.Size))
			}
			fmt.Printf("  ... and %d more movies (use -v to show full list)\n", len(plan.Moves)-maxShow)
		}
	}

	if cfg.DeleteJunk && len(plan.FileDeletions) > 0 {
		fmt.Printf("\n🗑️  Junk files to delete (%d, %s):\n",
			len(plan.FileDeletions),
			organizer.FormatBytes(plan.JunkBytesTotal))
		maxShow := 10
		if cfg.Verbose || len(plan.FileDeletions) <= maxShow {
			for _, d := range plan.FileDeletions {
				fmt.Printf("  • %s (%s)\n", rel(cfg.Path, d.Path), organizer.FormatBytes(d.Size))
			}
		} else {
			for i := 0; i < maxShow; i++ {
				d := plan.FileDeletions[i]
				fmt.Printf("  • %s (%s)\n", rel(cfg.Path, d.Path), organizer.FormatBytes(d.Size))
			}
			fmt.Printf("  ... and %d more files (use -v to show full list)\n", len(plan.FileDeletions)-maxShow)
		}
	}

	if cfg.CleanDirs && len(plan.DirDeletions) > 0 {
		fmt.Printf("\n📂 Empty or leftover subdirectories to remove (%d):\n", len(plan.DirDeletions))
		maxShow := 10
		if cfg.Verbose || len(plan.DirDeletions) <= maxShow {
			for _, d := range plan.DirDeletions {
				fmt.Printf("  • %s/\n", rel(cfg.Path, d))
			}
		} else {
			for i := 0; i < maxShow; i++ {
				d := plan.DirDeletions[i]
				fmt.Printf("  • %s/\n", rel(cfg.Path, d))
			}
			fmt.Printf("  ... and %d more folders (use -v to show full list)\n", len(plan.DirDeletions)-maxShow)
		}
	}
}

func printSummary(plan *organizer.Plan, cfg *config.Config) {
	fmt.Println("\n----------------------------------------------------------------")
	fmt.Println("📊 SUMMARY:")
	fmt.Printf("  • Movies found:                  %d\n", plan.MoviesFound)
	fmt.Printf("  • Movies already in place:       %d\n", plan.MoviesInPlace)
	fmt.Printf("  • Movies to relocate:            %d (%s)\n", len(plan.Moves), organizer.FormatBytes(plan.MovieBytesToMove))
	if cfg.DeleteJunk {
		fmt.Printf("  • Junk files to delete:          %d (%s to free)\n", len(plan.FileDeletions), organizer.FormatBytes(plan.JunkBytesTotal))
	}
	if cfg.CleanDirs {
		fmt.Printf("  • Subdirectories to remove:      %d\n", len(plan.DirDeletions))
	}
	fmt.Println("----------------------------------------------------------------")
}

func askConfirmation() bool {
	fmt.Print("\n⚠️  Do you want to apply these changes to your library? [y/N]: ")
	reader := bufio.NewReader(os.Stdin)
	response, err := reader.ReadString('\n')
	if err != nil {
		return false
	}

	response = strings.TrimSpace(strings.ToLower(response))
	return response == "y" || response == "yes" || response == "s" || response == "si" || response == "sí"
}

func rel(base, target string) string {
	relPath, err := filepath.Rel(base, target)
	if err != nil {
		return target
	}
	return relPath
}
