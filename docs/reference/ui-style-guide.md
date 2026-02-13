# CLI Style Guide: Dotfile Manager

## 1. Core Philosophy

GDF should feel **predictable but modern**. Dotfile management is a high-stakes task (accidentally overwriting a `.zshrc` is painful). The UI should prioritize **clarity over flashiness**, using color and animation only to reduce cognitive load or provide reassurance.

* **Informative, not noisy:** Use whitespace to group related items.
* **Safety first:** Destructive actions (deleting/overwriting) must look distinct.
* **Incremental disclosure:** Don't show a massive table if a simple status line suffices.

---

## 2. Color Palette (The "Slate & Electric" Palette)

A professional CLI palette avoids "pure" colors (like bright #FF0000) which can be jarring. Use subtle, muted tones for backgrounds and vibrant accents for focus.

| Element            | Hex Code  | Lip Gloss / ANSI | Purpose                                          |
| ------------------ | --------- | ---------------- | ------------------------------------------------ |
| **Primary Accent** | `#7D56F4` | `63` (Purple)    | Primary focus, highlights, and "Charm" branding. |
| **Success**        | `#04B575` | `42` (Green)     | Symlinks created, sync complete, "OK" status.    |
| **Warning**        | `#FFB347` | `215` (Orange)   | File conflicts, manual intervention required.    |
| **Error**          | `#E84855` | `197` (Red)      | Failed operations, invalid paths.                |
| **Dimmed/Subtle**  | `#626262` | `241` (Grey)     | Helper text, breadcrumbs, timestamps.            |
| **Highlight**      | `#00D7FF` | `45` (Cyan)      | Command suggestions, keybinding hints.           |

---

## 3. Component Usage Guide

### A. Navigation & Wizards (Bubble Tea / Gum)

When the user is in an interactive "setup" or "selection" mode.

* **Use Lists (`bubbles/list`):** E.g. For selecting which dotfiles to sync.
* **Use Paginators:** If the list of files exceeds 10 items, use the paginator to avoid scrolling the terminal history buffer.
* **The "Gum" Rule:** Use `gum` (or the underlying `bubbles/textinput`) for simple one-off questions like "What is your GitHub username?". Use a full `Bubble Tea` program for complex workflows like a multi-step installation wizard.

### B. Feedback & Progress (Bubbles)

* **Spinners (`bubbles/spinner`):** Use the `Dot` or `Pulse` spinner when fetching remote repos or running `git clone`.
* *Placement:* Always place the spinner to the left of the text: `⠋ Syncing .config/nvim...`

* **Progress Bars (`bubbles/progress`):** Use only for long-running batch operations (e.g., "Downloading 50 plugins"). If the task takes <2 seconds, stick to a spinner.

### C. Information Display (Lip Gloss)

* **Tables (`bubbles/table`):** Use for `list` commands.
* *Style:* Use a borderless or "thin" border style. Highlight the active row with the **Primary Accent** (#7D56F4).

* **Headers:** Use a bolded, underlined, or colored prefix for section headers.
* `LipGloss.NewStyle().Foreground(accent).Bold(true).Render("STRUCTURE")`



---

## 4. Typography & Layout Standards

### Status Indicators

Consistency is key for scanning logs:

* `[ OK   ]` — Green text.
* `[ FAIL ]` — Red text.
* `[ SKIP ]` — Grey text.
* `[ INFO ]` — Cyan text.

### Interactive Hints

Always provide an "exit hatch" and instructions at the bottom of the screen (the "Help" bubble).

* **Style:** Dimmed grey text.
* **Format:** `enter: select • q: back • ctrl+c: quit`

---

## 5. Decision Matrix: When to use what?

| User Action             | Component          | UI Logic                                                                |
| ----------------------- | ------------------ | ----------------------------------------------------------------------- |
| **Syncing Files**       | `Progress/Spinner` | Real-time feedback during IO operations.                                |
| **Conflict Resolution** | `Hue Selection`    | Use Red for the "Current" version and Green for the "Incoming" version. |
| **Configuration**       | `Forms / Inputs`   | Validation should happen as the user types (use `Validate` functions).  |
| **Browsing Files**      | `Tree View`        | Use indentations and icons (e.g., `󰓗` for symlinks) via `Lip Gloss`.   |

---

## 6. Code Snippet (Lip Gloss Styles)

To maintain consistency in your Go code, create a `ui` package and export these base styles:

```go
package ui

import "github.com/charmbracelet/lipgloss"

var (
    // Colors
    AccentColor  = lipgloss.Color("63")
    SuccessColor = lipgloss.Color("42")
    ErrorColor   = lipgloss.Color("197")
    DimmedColor  = lipgloss.Color("241")

    // Component Styles
    TitleStyle = lipgloss.NewStyle().
        Background(AccentColor).
        Foreground(lipgloss.Color("230")).
        Padding(0, 1).
        Bold(true)

    StatusStyle = lipgloss.NewStyle().
        PaddingLeft(1).
        MarginRight(1).
        Foreground(lipgloss.Color("255"))

    DocStyle = lipgloss.NewStyle().Margin(1, 2)
)

```

## 7. Professional Touches

1. **Empty States:** If a user has no dotfiles tracked, don't show an empty table. Show a "Welcome" banner using `lipgloss.Place` with a "Get Started" tip.
2. **Success Celebration:** When a sync finishes successfully, use a subtle "sparkle" or just a clean green checkmark. Avoid excessive ASCII art unless it's the very first time they run the tool.
3. **Adaptive UI:** Check the terminal width. If the terminal is too narrow, collapse the table into a simple list to avoid text wrapping which breaks CLI layouts.