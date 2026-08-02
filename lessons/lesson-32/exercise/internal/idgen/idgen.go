// Package idgen generuje identifikátory produktů.
//
// Leží pod internal/, takže ho smí importovat jen kód zakořeněný
// v lessons/lesson-32/exercise/ — tedy exercise, catalog i pricing.
// Cizí modul dostane při pokusu o import chybu kompilace. Tohle pravidlo
// vynucuje go tool, ne code review.
package idgen

import (
	"fmt"
	"sync"
)

// Gen je čítačový generátor ID. Nulová hodnota není použitelná, protože
// by neměla prefix — vytvářej ho přes New.
type Gen struct {
	mu     sync.Mutex
	prefix string
	n      int64
}

// New vytvoří generátor s daným prefixem.
// Prázdný prefix se nahradí "id". Nulová hodnota Gen není použitelná.
func New(prefix string) *Gen {
	// TODO
	return nil
}

// NewID vrací další identifikátor "<prefix>-000001", "<prefix>-000002", …
// Pořadové číslo na šest míst s nulami. Bezpečné pro souběžné volání.
func (g *Gen) NewID() string {
	// TODO
	return ""
}

// format sestaví identifikátor z prefixu a pořadového čísla.
func format(prefix string, n int64) string {
	return fmt.Sprintf("%s-%06d", prefix, n)
}
