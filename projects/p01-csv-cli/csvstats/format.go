package csvstats

import (
	"fmt"
	"io"
	"text/tabwriter"
)

// RenderSummary zapíše tabulku kategorií a celkový součet.
// Výstup je deterministický, takže se dá porovnávat s golden souborem.
func RenderSummary(w io.Writer, s Summary) error {
	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)

	if _, err := fmt.Fprintln(tw, "KATEGORIE\tPOČET\tSOUČET\tPRŮMĚR"); err != nil {
		return fmt.Errorf("render summary: %w", err)
	}
	for _, c := range s.Categories {
		if _, err := fmt.Fprintf(tw, "%s\t%d\t%.2f\t%.2f\n", c.Category, c.Count, c.Total, c.Average); err != nil {
			return fmt.Errorf("render summary: %w", err)
		}
	}
	if _, err := fmt.Fprintf(tw, "CELKEM\t%d\t%.2f\n", s.Records, s.Total); err != nil {
		return fmt.Errorf("render summary: %w", err)
	}
	if err := tw.Flush(); err != nil {
		return fmt.Errorf("render summary: %w", err)
	}
	return nil
}

// RenderTop zapíše tabulku jednotlivých záznamů.
func RenderTop(w io.Writer, recs []Record) error {
	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)

	if _, err := fmt.Fprintln(tw, "JMÉNO\tKATEGORIE\tČÁSTKA"); err != nil {
		return fmt.Errorf("render top: %w", err)
	}
	for _, rec := range recs {
		if _, err := fmt.Fprintf(tw, "%s\t%s\t%.2f\n", rec.Name, rec.Category, rec.Amount); err != nil {
			return fmt.Errorf("render top: %w", err)
		}
	}
	if err := tw.Flush(); err != nil {
		return fmt.Errorf("render top: %w", err)
	}
	return nil
}
