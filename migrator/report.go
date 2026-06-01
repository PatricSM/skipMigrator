package migrator

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// MigratorVersion is stamped into the report; bump on releases.
const MigratorVersion = "0.1.0"

// writeReport writes MIGRATION_REPORT.md into opts.OutputDir summarizing
// every phase, the build status, and recommended next steps.
func writeReport(opts Options, result *Result, finalErr error) error {
	if result == nil {
		return nil
	}

	srcName, srcPkgVersion := readSourcePackageMeta(opts.SourceDir)

	var sb strings.Builder
	sb.WriteString("# Skip Migrator — Relatório de Migração\n\n")
	fmt.Fprintf(&sb, "**Data:** %s\n", time.Now().UTC().Format("2006-01-02 15:04 UTC"))
	fmt.Fprintf(&sb, "**Skip Migrator:** v%s\n", MigratorVersion)
	if srcName != "" {
		fmt.Fprintf(&sb, "**Projeto origem:** `%s` (versão %s)\n", srcName, srcPkgVersion)
	}
	fmt.Fprintf(&sb, "**Modo pixel-perfect:** %v\n", opts.PixelPerfect)
	fmt.Fprintf(&sb, "**Build validado:** %v\n", opts.Validate)
	sb.WriteString("\n---\n\n")

	if finalErr != nil {
		sb.WriteString("## ❌ Resultado: FALHOU\n\n")
		fmt.Fprintf(&sb, "Erro fatal: `%v`\n\n", finalErr)
	} else if opts.Validate {
		sb.WriteString("## ✅ Resultado: SUCESSO (build validado)\n\n")
		sb.WriteString("`pnpm install`, `tsc --noEmit` e `pnpm build` completaram sem erros.\n\n")
	} else {
		sb.WriteString("## ✅ Resultado: TRANSFORMAÇÕES APLICADAS (build não validado)\n\n")
		sb.WriteString("Migração concluída sem rodar build. Rode `pnpm install && pnpm build` localmente para validar.\n\n")
	}

	sb.WriteString("## Fases executadas\n\n")
	sb.WriteString("| # | Fase | Status | Duração | Mensagem |\n")
	sb.WriteString("|---|---|---|---|---|\n")
	for i, p := range result.PhaseLogs {
		fmt.Fprintf(&sb, "| %02d | %s | %s | %s | %s |\n",
			i+1, p.Phase, statusIcon(p.Status), p.Duration, escapeTable(p.Message))
	}
	sb.WriteString("\n")

	// Per-phase details (only phases that have something interesting)
	anyDetails := false
	for _, p := range result.PhaseLogs {
		if len(p.Details) > 0 {
			anyDetails = true
			break
		}
	}
	if anyDetails {
		sb.WriteString("## Transformações de código aplicadas\n\n")
		for _, p := range result.PhaseLogs {
			if len(p.Details) == 0 {
				continue
			}
			fmt.Fprintf(&sb, "### %s\n\n", p.Phase)
			for _, d := range p.Details {
				fmt.Fprintf(&sb, "- %s\n", d)
			}
			sb.WriteString("\n")
		}
	}

	if result.BuildLog != "" {
		sb.WriteString("## Cauda do build log\n\n")
		sb.WriteString("```\n")
		sb.WriteString(tail(result.BuildLog, 2000))
		sb.WriteString("\n```\n\n")
	}

	sb.WriteString("## O que NÃO foi validado\n\n")
	sb.WriteString("- **Lógica de negócio em runtime**: o build passou mas o app pode crashar ao abrir.\n")
	sb.WriteString("- **Rotas autenticadas e edge functions**: copiadas mas não exercitadas.\n")
	sb.WriteString("- **Comportamento de libs em runtime**: APIs com mesma assinatura mas semântica diferente entre versões.\n\n")

	sb.WriteString("## Próximos passos recomendados\n\n")
	sb.WriteString("1. Extraia o ZIP em uma pasta limpa.\n")
	sb.WriteString("2. Ajuste `.env.local` se as credenciais Supabase mudaram.\n")
	sb.WriteString("3. `pnpm install`\n")
	sb.WriteString("4. `pnpm dev` — abra http://localhost:8080 e teste os fluxos críticos.\n")
	sb.WriteString("5. Se algo quebrar, consulte a tabela de transformações acima para entender o que mudou.\n")
	sb.WriteString("6. Antes de fazer deploy, rode `pnpm build` localmente e teste com `pnpm preview`.\n\n")

	sb.WriteString("---\n_Gerado automaticamente pelo Skip Migrator._\n")

	outPath := filepath.Join(opts.OutputDir, "MIGRATION_REPORT.md")
	return os.WriteFile(outPath, []byte(sb.String()), 0o644)
}

func statusIcon(status string) string {
	switch status {
	case "ok":
		return "✅ ok"
	case "warn":
		return "⚠️ warn"
	case "skipped":
		return "⏭ skipped"
	case "error":
		return "❌ error"
	default:
		return status
	}
}

func escapeTable(s string) string {
	return strings.ReplaceAll(s, "|", `\|`)
}

func tail(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return "[…truncated…]\n" + s[len(s)-max:]
}

// readSourcePackageMeta extracts name + version from package.json for display.
// Returns empty strings if anything fails — the report still renders.
func readSourcePackageMeta(srcDir string) (name, version string) {
	data, err := os.ReadFile(filepath.Join(srcDir, "package.json"))
	if err != nil {
		return "", ""
	}
	var pkg struct {
		Name    string `json:"name"`
		Version string `json:"version"`
	}
	if err := json.Unmarshal(data, &pkg); err != nil {
		return "", ""
	}
	return pkg.Name, pkg.Version
}
