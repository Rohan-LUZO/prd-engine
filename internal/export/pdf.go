package export

import (
	"io"
	"strconv"
	"strings"
	"time"

	"codeberg.org/go-pdf/fpdf"
)

// Document is the input for PDF export (all modules, latest versions, with audit info).
type Document struct {
	Title       string
	GeneratedAt time.Time
	Modules     []Module
}

// Module is one module's snapshot for export.
type Module struct {
	ID        string
	Title     string
	Order     int
	Version   int
	Content   string
	UpdatedBy string
	UpdatedAt time.Time
	CreatedBy string
	CreatedAt time.Time
}

// Layout constants for A4 PDF (millimetres). bodyW is content width after margins.
const (
	margin   = 20
	pageW    = 210  // A4 width
	bodyW    = pageW - 2*margin
	lineH    = 5
	fontSize = 10
)

// BuildPDF writes a single PDF to w: cover with title, each module's content,
// and a final "Document history" page (who updated what, versioning).
func BuildPDF(doc *Document, w io.Writer) error {
	if doc == nil {
		doc = &Document{Title: "Documentation", GeneratedAt: time.Now()}
	}

	pdf := fpdf.New("P", "mm", "A4", "")
	pdf.SetMargins(margin, margin, margin)
	pdf.SetAutoPageBreak(true, margin)
	pdf.AddPage()

	// Title page / header
	pdf.SetFont("Helvetica", "B", 18)
	pdf.CellFormat(0, 10, doc.Title, "", 1, "L", false, 0, "")
	pdf.SetFont("Helvetica", "", 10)
	pdf.CellFormat(0, 6, "Generated: "+doc.GeneratedAt.Format(time.RFC3339), "", 1, "L", false, 0, "")
	pdf.Ln(8)

	for _, m := range doc.Modules {
		// Module title
		pdf.SetFont("Helvetica", "B", 14)
		pdf.CellFormat(0, lineH+2, m.Title, "", 1, "L", false, 0, "")
		pdf.Ln(2)

		// Content: simple markdown-style (headings = larger font, rest = body)
		pdf.SetFont("Helvetica", "", fontSize)
		writeContent(pdf, m.Content)

		// Audit line
		pdf.SetFont("Helvetica", "I", 8)
		audit := "Last updated by " + m.UpdatedBy + " on " + m.UpdatedAt.Format("2006-01-02 15:04") + " (v" + strconv.Itoa(m.Version) + ")"
		pdf.CellFormat(0, lineH, audit, "", 1, "L", false, 0, "")
		pdf.Ln(6)
	}

	// Document history page
	pdf.AddPage()
	pdf.SetFont("Helvetica", "B", 14)
	pdf.CellFormat(0, 10, "Document history", "", 1, "L", false, 0, "")
	pdf.Ln(4)
	pdf.SetFont("Helvetica", "", 9)

	// Table header
	colW := []float64{50, 20, 55, 45}
	pdf.SetFont("Helvetica", "B", 9)
	pdf.CellFormat(colW[0], lineH+1, "Module", "1", 0, "L", true, 0, "")
	pdf.CellFormat(colW[1], lineH+1, "Version", "1", 0, "C", true, 0, "")
	pdf.CellFormat(colW[2], lineH+1, "Last updated by", "1", 0, "L", true, 0, "")
	pdf.CellFormat(colW[3], lineH+1, "Date", "1", 1, "L", true, 0, "")
	pdf.SetFont("Helvetica", "", 9)

	for _, m := range doc.Modules {
		pdf.CellFormat(colW[0], lineH, truncate(m.Title, 35), "1", 0, "L", false, 0, "")
		pdf.CellFormat(colW[1], lineH, strconv.Itoa(m.Version), "1", 0, "C", false, 0, "")
		pdf.CellFormat(colW[2], lineH, truncate(m.UpdatedBy, 25), "1", 0, "L", false, 0, "")
		pdf.CellFormat(colW[3], lineH, m.UpdatedAt.Format("2006-01-02 15:04"), "1", 1, "L", false, 0, "")
	}

	return pdf.Output(w)
}

// writeContent renders markdown-style content: lines starting with #, ##, ### are rendered
// as bold headings with decreasing font size; blank lines add spacing; everything else is body text.
func writeContent(pdf *fpdf.Fpdf, content string) {
	lines := strings.Split(content, "\n")
	for _, line := range lines {
		line = strings.TrimRight(line, "\r")
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			pdf.Ln(lineH * 0.5)
			continue
		}
		// Count leading '#' to detect markdown headings (H1–H3); require space after hashes
		level := 0
		for level < len(trimmed) && trimmed[level] == '#' {
			level++
		}
		if level > 0 && level <= 3 && (level == len(trimmed) || trimmed[level] == ' ') {
			text := strings.TrimSpace(trimmed[level:])
			if text == "" {
				continue
			}
			size := 12.0 - float64(level)*2
			if size < 8 {
				size = 8
			}
			pdf.SetFont("Helvetica", "B", size)
			pdf.MultiCell(bodyW, lineH+1, text, "", "L", false)
			pdf.SetFont("Helvetica", "", fontSize)
			continue
		}
		// Normal paragraph (no markdown heading)
		pdf.MultiCell(bodyW, lineH, line, "", "L", false)
	}
}

// truncate shortens s to at most max runes and appends "…" for display in the history table.
func truncate(s string, max int) string {
	r := []rune(s)
	if len(r) <= max {
		return s
	}
	return string(r[:max-1]) + "…"
}
