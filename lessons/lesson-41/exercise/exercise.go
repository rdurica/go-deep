// Package exercise obsahuje cvičení lekce 41.
package exercise

// --- Stupeň: jednoduchý ---

// ForgetClose pošle všechna čísla do kanálu v goroutině, ale kanál nezavře.
// Collect na výsledku by visel navždy.
//
// POZOR: kód níže je ZÁMĚRNĚ VADNÝ. Najdi chybu a oprav ji — testy před opravou padají.
func ForgetClose(nums ...int) <-chan int {
	out := make(chan int)
	go func() {
		for _, n := range nums {
			out <- n
		}
		// Chybí close(out) — příjemce nikdy nedostane signál konce.
	}()
	return out
}

// --- Stupeň: střední ---

// Generate pošle všechna čísla do kanálu v goroutině a kanál sám zavře.
// Bez Collect nesmí blokovat volajícího.
func Generate(nums ...int) <-chan int {
	// TODO
	return closedInt()
}

// Collect přečte kanál do zavření a vrátí hodnoty v pořadí doručení.
// Prázdný nebo nil kanál → prázdný slice, ne panika.
func Collect(ch <-chan int) []int {
	// TODO
	return nil
}

// --- Stupeň: obtížný ---

// Merge sloučí vstupní kanály (fan-in). Výstup zavře jednou po zavření všech vstupů.
// Bez argumentů → rovnou zavřený kanál.
func Merge(chs ...<-chan int) <-chan int {
	// TODO
	return closedInt()
}

// closedInt je fail-fast stub: nil kanál by v testech visel navždy.
func closedInt() <-chan int {
	ch := make(chan int)
	close(ch)
	return ch
}
