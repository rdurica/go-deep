package order

import "errors"

// Doménové chyby. Adaptéry je poznávají přes errors.Is a překládají na
// protokol, který zrovna obsluhují — doména sama o HTTP statusech neví.
var (
	// ErrInvalidCurrency znamená, že kód měny není trojpísmenný ISO kód.
	ErrInvalidCurrency = errors.New("neplatný kód měny")
	// ErrNegativeAmount znamená, že částka nebo násobek je záporný.
	ErrNegativeAmount = errors.New("částka nesmí být záporná")
	// ErrCurrencyMismatch znamená, že se počítá s částkami v různých měnách.
	ErrCurrencyMismatch = errors.New("nelze míchat různé měny")
	// ErrAmountOverflow znamená, že výsledek se nevejde do int64.
	ErrAmountOverflow = errors.New("částka přetekla")

	// ErrMissingID znamená, že objednávka nemá identifikátor.
	ErrMissingID = errors.New("objednávka nemá ID")
	// ErrMissingCustomer znamená, že objednávka nemá zákazníka.
	ErrMissingCustomer = errors.New("objednávka nemá zákazníka")
	// ErrMissingTimestamp znamená, že objednávka nemá čas založení.
	ErrMissingTimestamp = errors.New("objednávka nemá čas založení")
	// ErrEmptyOrder znamená, že objednávka nemá žádnou položku.
	ErrEmptyOrder = errors.New("objednávka nemá žádnou položku")
	// ErrInvalidLine znamená, že položka objednávky porušuje invariant.
	ErrInvalidLine = errors.New("neplatná položka objednávky")
	// ErrInvalidTransition znamená, že přechod mezi stavy není povolený.
	ErrInvalidTransition = errors.New("nepovolený přechod stavu")
	// ErrNotFound znamená, že objednávka neexistuje.
	ErrNotFound = errors.New("objednávka nenalezena")
)
