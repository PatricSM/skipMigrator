package migrator

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
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

	// Supabase checklist: only renders if we can extract a project ref from the URL.
	if ref := extractSupabaseRef(opts.OutputDir); ref != "" {
		writeSupabaseChecklist(&sb, ref)
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

// writeSupabaseChecklist appends a post-migration Supabase setup checklist
// with direct deep-links to the project's dashboard panels.
func writeSupabaseChecklist(sb *strings.Builder, ref string) {
	base := fmt.Sprintf("https://supabase.com/dashboard/project/%s", ref)

	sb.WriteString("## Checklist pós-migração: Supabase\n\n")
	sb.WriteString("Como a app importada vai conectar no MESMO projeto Supabase do Lovable, ")
	sb.WriteString("o banco já está pronto — schema, RLS, triggers, edge functions e buckets continuam valendo. ")
	sb.WriteString("**Você só precisa mexer no painel Supabase se for trocar o domínio público do app.**\n\n")

	sb.WriteString("### Não precisa mudar nada\n\n")
	sb.WriteString("- Schema do banco (migrations já rodaram pelo Lovable)\n")
	sb.WriteString("- RLS policies, triggers, functions, enums\n")
	sb.WriteString("- Storage buckets (já existem)\n")
	sb.WriteString("- Edge functions (já deployadas)\n")
	sb.WriteString("- Sessões/JWT dos usuários ativos (continuam válidos)\n\n")

	sb.WriteString("### Verifique APENAS se mudar o domínio público (ex: de `*.lovable.app` para um próprio)\n\n")
	sb.WriteString("| O quê | Onde | Quando importa |\n")
	sb.WriteString("|---|---|---|\n")
	fmt.Fprintf(sb, "| **Site URL** | [Auth > URL Configuration](%s/auth/url-configuration) | E-mails de reset/confirmação apontam pra cá |\n", base)
	fmt.Fprintf(sb, "| **Redirect URLs** (adicionar a nova) | [Auth > URL Configuration](%s/auth/url-configuration) | OAuth, magic link, password reset voltam pra cá |\n", base)
	fmt.Fprintf(sb, "| **OAuth callbacks** (Google/GitHub etc) | [Auth > Providers](%s/auth/providers) | Cada provider tem callback URL que precisa bater com a nova |\n", base)
	fmt.Fprintf(sb, "| **CORS allowed origins** | [Project Settings > API](%s/settings/api) | Só se você já tiver restringido (default permite tudo) |\n", base)
	sb.WriteString("\n")

	sb.WriteString("### Regenerar types.ts (somente se mudar schema no futuro)\n\n")
	sb.WriteString("O `src/integrations/supabase/types.ts` é uma cópia do que o Lovable gerou. Se você alterar tabelas/colunas depois, regenere:\n\n")
	fmt.Fprintf(sb, "```bash\nnpx supabase gen types typescript --project-id %s > src/integrations/supabase/types.ts\n```\n\n", ref)

	fmt.Fprintf(sb, "**Atalho pro painel do projeto:** %s\n\n", base)
}

// ExtractSupabaseRef returns the Supabase project ref (the 20-char subdomain
// of *.supabase.co) found in a migrated project directory. Reads .env.local
// first, then src/integrations/supabase/client.ts. Returns "" if not found.
func ExtractSupabaseRef(outDir string) string {
	return extractSupabaseRef(outDir)
}

// extractSupabaseRef reads the project ref from .env.local (preferred)
// or from src/integrations/supabase/client.ts as fallback. Returns "" if
// no URL like https://<ref>.supabase.co can be located.
func extractSupabaseRef(outDir string) string {
	candidates := []string{
		filepath.Join(outDir, ".env.local"),
		filepath.Join(outDir, "src", "integrations", "supabase", "client.ts"),
	}
	re := regexp.MustCompile(`https?://([a-z0-9]{20})\.supabase\.co`)
	for _, p := range candidates {
		data, err := os.ReadFile(p)
		if err != nil {
			continue
		}
		if m := re.FindSubmatch(data); len(m) > 1 {
			return string(m[1])
		}
	}
	return ""
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
