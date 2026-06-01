package migrator

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// feature dirs we always copy verbatim from source/src/components/
var featureComponentDirs = []string{
	"admin", "documents", "landing", "layout", "locador",
	"notifications", "property", "rental", "tenant", "tracking",
}

// phase06Components copies feature component subdirs and merges components/ui.
// Pixel-perfect mode overrides all ui/*.tsx with source versions + adapts calendar v8→v9.
// Fast mode keeps Skip's ui/ and only copies files missing from baseline.
func phase06Components(opts Options) (PhaseLog, error) {
	start := time.Now()
	log := PhaseLog{Phase: "06-components"}

	// Remove Skip baseline's monolithic Layout.tsx — source has its own per-flow layouts
	_ = os.Remove(filepath.Join(opts.OutputDir, "src", "components", "Layout.tsx"))

	// Copy feature dirs
	copiedDirs := 0
	for _, d := range featureComponentDirs {
		src := filepath.Join(opts.SourceDir, "src", "components", d)
		if !fileExists(src) {
			continue
		}
		dst := filepath.Join(opts.OutputDir, "src", "components", d)
		if err := os.MkdirAll(dst, 0o755); err != nil {
			return log, err
		}
		if err := copyDir(src, dst, nil); err != nil {
			return log, fmt.Errorf("copying components/%s: %w", d, err)
		}
		copiedDirs++
	}

	// NavLink.tsx (commonly in source root of components/)
	srcNav := filepath.Join(opts.SourceDir, "src", "components", "NavLink.tsx")
	if fileExists(srcNav) {
		_ = copyFile(srcNav, filepath.Join(opts.OutputDir, "src", "components", "NavLink.tsx"))
	}

	// components/ui merge strategy
	srcUI := filepath.Join(opts.SourceDir, "src", "components", "ui")
	dstUI := filepath.Join(opts.OutputDir, "src", "components", "ui")
	uiAdded, uiOverridden := 0, 0
	if fileExists(srcUI) {
		entries, err := os.ReadDir(srcUI)
		if err == nil {
			for _, e := range entries {
				if e.IsDir() {
					continue
				}
				name := e.Name()
				srcFile := filepath.Join(srcUI, name)
				dstFile := filepath.Join(dstUI, name)
				if opts.PixelPerfect {
					// Override everything
					if err := copyFile(srcFile, dstFile); err != nil {
						return log, fmt.Errorf("override ui/%s: %w", name, err)
					}
					if fileExists(dstFile) {
						uiOverridden++
					} else {
						uiAdded++
					}
				} else {
					// Only copy if missing
					if fileExists(dstFile) {
						continue
					}
					if err := copyFile(srcFile, dstFile); err != nil {
						return log, fmt.Errorf("add ui/%s: %w", name, err)
					}
					uiAdded++
				}
			}
		}
	}

	// If pixel-perfect, the source's calendar.tsx uses react-day-picker v8 API; adapt to v9
	if opts.PixelPerfect {
		calPath := filepath.Join(dstUI, "calendar.tsx")
		if fileExists(calPath) {
			if err := os.WriteFile(calPath, []byte(calendarV9Source), 0o644); err != nil {
				return log, fmt.Errorf("writing calendar.tsx v9 adapter: %w", err)
			}
		}
	}

	log.Status = "ok"
	log.Message = fmt.Sprintf("feature dirs: %d copied; ui: %d added, %d overridden (pixel-perfect=%v)", copiedDirs, uiAdded, uiOverridden, opts.PixelPerfect)
	log.Duration = time.Since(start).String()
	return log, nil
}

// calendarV9Source is the react-day-picker v9-compatible Calendar component
// preserving 02dafc89/Lovable visual styling.
const calendarV9Source = `import * as React from "react";
import { ChevronLeft, ChevronRight } from "lucide-react";
import { DayPicker } from "react-day-picker";

import { cn } from "@/lib/utils";
import { buttonVariants } from "@/components/ui/button";

export type CalendarProps = React.ComponentProps<typeof DayPicker>;

function Calendar({ className, classNames, showOutsideDays = true, ...props }: CalendarProps) {
  return (
    <DayPicker
      showOutsideDays={showOutsideDays}
      className={cn("p-3", className)}
      classNames={{
        months: "flex flex-col sm:flex-row space-y-4 sm:space-x-4 sm:space-y-0",
        month: "space-y-4",
        month_caption: "flex justify-center pt-1 relative items-center",
        caption_label: "text-sm font-medium",
        nav: "space-x-1 flex items-center absolute inset-x-0 top-1 justify-between px-1",
        button_previous: cn(
          buttonVariants({ variant: "outline" }),
          "h-7 w-7 bg-transparent p-0 opacity-50 hover:opacity-100",
        ),
        button_next: cn(
          buttonVariants({ variant: "outline" }),
          "h-7 w-7 bg-transparent p-0 opacity-50 hover:opacity-100",
        ),
        month_grid: "w-full border-collapse space-y-1",
        weekdays: "flex",
        weekday: "text-muted-foreground rounded-md w-9 font-normal text-[0.8rem]",
        week: "flex w-full mt-2",
        day: "h-9 w-9 text-center text-sm p-0 relative [&:has([aria-selected].day-range-end)]:rounded-r-md [&:has([aria-selected].day-outside)]:bg-accent/50 [&:has([aria-selected])]:bg-accent first:[&:has([aria-selected])]:rounded-l-md last:[&:has([aria-selected])]:rounded-r-md focus-within:relative focus-within:z-20",
        day_button: cn(buttonVariants({ variant: "ghost" }), "h-9 w-9 p-0 font-normal aria-selected:opacity-100"),
        range_end: "day-range-end",
        selected:
          "bg-primary text-primary-foreground hover:bg-primary hover:text-primary-foreground focus:bg-primary focus:text-primary-foreground",
        today: "bg-accent text-accent-foreground",
        outside:
          "day-outside text-muted-foreground opacity-50 aria-selected:bg-accent/50 aria-selected:text-muted-foreground aria-selected:opacity-30",
        disabled: "text-muted-foreground opacity-50",
        range_middle: "aria-selected:bg-accent aria-selected:text-accent-foreground",
        hidden: "invisible",
        ...classNames,
      }}
      components={{
        Chevron: ({ orientation, ...iconProps }) =>
          orientation === "left" ? (
            <ChevronLeft className="h-4 w-4" {...iconProps} />
          ) : (
            <ChevronRight className="h-4 w-4" {...iconProps} />
          ),
      }}
      {...props}
    />
  );
}
Calendar.displayName = "Calendar";

export { Calendar };
`
