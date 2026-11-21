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
		fmt.Fprintf(os.Stderr, "❌ Error de configuración: %v\n\n", err)
		os.Exit(1)
	}

	printHeader(cfg)

	org := organizer.New(cfg)

	fmt.Println("🔍 Escaneando biblioteca...")
	plan, err := org.BuildPlan()
	if err != nil {
		fmt.Fprintf(os.Stderr, "❌ Error al analizar la biblioteca: %v\n", err)
		os.Exit(1)
	}

	printPlanDetails(plan, cfg)
	printSummary(plan, cfg)

	// If no changes are needed
	if len(plan.Moves) == 0 && len(plan.FileDeletions) == 0 && len(plan.DirDeletions) == 0 {
		fmt.Println("\n🎉 Tu biblioteca ya está completamente organizada y limpia. ¡No hay nada que hacer!")
		return
	}

	// Dry run mode stops here
	if cfg.DryRun {
		fmt.Println("\nℹ️  MODO SIMULACIÓN (dry-run): No se ha modificado ningún archivo ni carpeta.")
		fmt.Println("👉 Para aplicar los cambios reales, vuelve a ejecutar el comando sin '-dry-run' o '-n'.")
		return
	}

	// Confirmation prompt if -y / -yes was not supplied
	if !cfg.Yes {
		if !askConfirmation() {
			fmt.Println("\n🚫 Operación cancelada. No se realizaron cambios.")
			return
		}
	}

	fmt.Println("\n🚀 Aplicando cambios...")
	if err := org.Execute(plan); err != nil {
		fmt.Fprintf(os.Stderr, "\n❌ Error al ejecutar los cambios: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("\n✅ ¡Listo! Biblioteca organizada y limpia con éxito.")
}

func printHeader(cfg *config.Config) {
	fmt.Println("================================================================")
	fmt.Println("   🎬 SORTFLIX - Organizador & Limpiador de Películas")
	fmt.Println("================================================================")
	fmt.Printf("📁 Carpeta raíz:   %s\n", cfg.Path)
	fmt.Printf("⚖️  Tamaño mínimo:  %d MB\n", cfg.MinSizeMB)
	if cfg.DryRun {
		fmt.Println("🛡️  Modo:           SIMULACIÓN (dry-run) - No se alterará nada")
	} else {
		fmt.Println("⚡ Modo:           EJECUCIÓN REAL")
	}
	fmt.Println("================================================================")
}

func printPlanDetails(plan *organizer.Plan, cfg *config.Config) {
	if len(plan.Moves) > 0 {
		fmt.Printf("\n📦 Películas a mover a carpetas alfabéticas (%d):\n", len(plan.Moves))
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
			fmt.Printf("  ... y %d películas más (usa -v para ver la lista completa)\n", len(plan.Moves)-maxShow)
		}
	}

	if cfg.DeleteJunk && len(plan.FileDeletions) > 0 {
		fmt.Printf("\n🗑️  Archivos basura a eliminar (%d, %s):\n",
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
			fmt.Printf("  ... y %d archivos más (usa -v para ver la lista completa)\n", len(plan.FileDeletions)-maxShow)
		}
	}

	if cfg.CleanDirs && len(plan.DirDeletions) > 0 {
		fmt.Printf("\n📂 Subcarpetas vacías o residuales a eliminar (%d):\n", len(plan.DirDeletions))
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
			fmt.Printf("  ... y %d carpetas más (usa -v para ver la lista completa)\n", len(plan.DirDeletions)-maxShow)
		}
	}
}

func printSummary(plan *organizer.Plan, cfg *config.Config) {
	fmt.Println("\n----------------------------------------------------------------")
	fmt.Println("📊 RESUMEN:")
	fmt.Printf("  • Películas encontradas:         %d\n", plan.MoviesFound)
	fmt.Printf("  • Películas ya en su sitio:      %d\n", plan.MoviesInPlace)
	fmt.Printf("  • Películas a reubicar:          %d (%s)\n", len(plan.Moves), organizer.FormatBytes(plan.MovieBytesToMove))
	if cfg.DeleteJunk {
		fmt.Printf("  • Archivos basura a eliminar:    %d (%s a liberar)\n", len(plan.FileDeletions), organizer.FormatBytes(plan.JunkBytesTotal))
	}
	if cfg.CleanDirs {
		fmt.Printf("  • Subcarpetas a suprimir:        %d\n", len(plan.DirDeletions))
	}
	fmt.Println("----------------------------------------------------------------")
}

func askConfirmation() bool {
	fmt.Print("\n⚠️  ¿Deseas aplicar estos cambios en tu biblioteca? [s/N]: ")
	reader := bufio.NewReader(os.Stdin)
	response, err := reader.ReadString('\n')
	if err != nil {
		return false
	}

	response = strings.TrimSpace(strings.ToLower(response))
	return response == "s" || response == "si" || response == "sí" || response == "y" || response == "yes"
}

func rel(base, target string) string {
	relPath, err := filepath.Rel(base, target)
	if err != nil {
		return target
	}
	return relPath
}
