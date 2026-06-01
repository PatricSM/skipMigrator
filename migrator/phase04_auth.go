package migrator

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// phase04Auth replaces Skip's simple auth (use-auth.tsx + ProtectedRoute.tsx)
// with the source's AuthContext + role-based system.
func phase04Auth(opts Options) (PhaseLog, error) {
	start := time.Now()
	log := PhaseLog{Phase: "04-auth"}

	// Remove Skip baseline's simple auth
	_ = os.Remove(filepath.Join(opts.OutputDir, "src", "hooks", "use-auth.tsx"))
	_ = os.Remove(filepath.Join(opts.OutputDir, "src", "components", "ProtectedRoute.tsx"))

	// Copy AuthContext
	srcContext := filepath.Join(opts.SourceDir, "src", "contexts", "AuthContext.tsx")
	if fileExists(srcContext) {
		dst := filepath.Join(opts.OutputDir, "src", "contexts", "AuthContext.tsx")
		if err := copyFile(srcContext, dst); err != nil {
			return log, fmt.Errorf("copying AuthContext: %w", err)
		}
	}

	// Copy useAuth hook
	srcUseAuth := filepath.Join(opts.SourceDir, "src", "hooks", "useAuth.ts")
	if fileExists(srcUseAuth) {
		dst := filepath.Join(opts.OutputDir, "src", "hooks", "useAuth.ts")
		if err := copyFile(srcUseAuth, dst); err != nil {
			return log, fmt.Errorf("copying useAuth: %w", err)
		}
	}

	// Copy entire components/auth directory
	srcAuthDir := filepath.Join(opts.SourceDir, "src", "components", "auth")
	if fileExists(srcAuthDir) {
		dst := filepath.Join(opts.OutputDir, "src", "components", "auth")
		if err := os.MkdirAll(dst, 0o755); err != nil {
			return log, err
		}
		if err := copyDir(srcAuthDir, dst, nil); err != nil {
			return log, fmt.Errorf("copying components/auth: %w", err)
		}
	}

	log.Status = "ok"
	log.Message = "AuthContext + useAuth + components/auth/* installed; baseline simple auth removed"
	log.Duration = time.Since(start).String()
	return log, nil
}
