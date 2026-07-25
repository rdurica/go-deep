package exercise_test

import (
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	exercise "github.com/rdurica/go-deep/lessons/lesson-43/exercise"
)

func TestCounterZeroValueIsUsable(t *testing.T) {
	var c exercise.Counter
	if got := c.Value(); got != 0 {
		t.Errorf("Value() na zero value = %d, chci 0", got)
	}
	c.Inc()
	c.Add(41)
	c.Add(-2)
	if got := c.Value(); got != 40 {
		t.Errorf("Value() = %d, chci 40", got)
	}
}

func TestCounterConcurrentInc(t *testing.T) {
	var c exercise.Counter
	const goroutines, perGoroutine = 100, 1000

	var wg sync.WaitGroup
	wg.Add(goroutines)
	for i := 0; i < goroutines; i++ {
		go func() {
			defer wg.Done()
			for j := 0; j < perGoroutine; j++ {
				c.Inc()
			}
		}()
	}
	wg.Wait()

	if want := int64(goroutines * perGoroutine); c.Value() != want {
		t.Errorf("Value() = %d, chci %d — čítač ztrácí zvýšení", c.Value(), want)
	}
}

func TestCounterConcurrentMixedOperations(t *testing.T) {
	var c exercise.Counter
	const goroutines = 64

	var wg sync.WaitGroup
	wg.Add(2 * goroutines)
	for i := 0; i < goroutines; i++ {
		go func() {
			defer wg.Done()
			for j := 0; j < 500; j++ {
				c.Add(3)
			}
		}()
		go func() {
			defer wg.Done()
			for j := 0; j < 500; j++ {
				c.Add(-3)
				_ = c.Value() // souběžné čtení musí být taky bezpečné
			}
		}()
	}
	wg.Wait()

	if got := c.Value(); got != 0 {
		t.Errorf("Value() = %d, chci 0", got)
	}
}

func TestCacheBasicOperations(t *testing.T) {
	c := exercise.NewCache()

	if _, ok := c.Get("chybí"); ok {
		t.Error("Get na neexistující klíč vrátil ok = true")
	}
	if got := c.Len(); got != 0 {
		t.Errorf("Len() = %d, chci 0", got)
	}

	c.Set("a", "1")
	c.Set("b", "2")
	c.Set("a", "3") // přepis

	if got, ok := c.Get("a"); !ok || got != "3" {
		t.Errorf("Get(\"a\") = (%q, %v), chci (\"3\", true)", got, ok)
	}
	if got := c.Len(); got != 2 {
		t.Errorf("Len() = %d, chci 2", got)
	}

	c.Delete("a")
	c.Delete("neexistuje") // nesmí panikovat
	if _, ok := c.Get("a"); ok {
		t.Error("Get po Delete vrátil ok = true")
	}
	if got := c.Len(); got != 1 {
		t.Errorf("Len() po Delete = %d, chci 1", got)
	}
}

func TestCacheConcurrentReadWrite(t *testing.T) {
	c := exercise.NewCache()
	const goroutines = 50

	var wg sync.WaitGroup
	wg.Add(3 * goroutines)
	for i := 0; i < goroutines; i++ {
		key := fmt.Sprintf("klíč-%d", i%10)
		go func() {
			defer wg.Done()
			for j := 0; j < 200; j++ {
				c.Set(key, fmt.Sprintf("hodnota-%d", j))
			}
		}()
		go func() {
			defer wg.Done()
			for j := 0; j < 200; j++ {
				c.Get(key)
				c.Len()
			}
		}()
		go func() {
			defer wg.Done()
			for j := 0; j < 200; j++ {
				c.Delete(key)
			}
		}()
	}
	wg.Wait()
}

func TestCacheGetOrComputeCallsFunctionOncePerKey(t *testing.T) {
	c := exercise.NewCache()
	const goroutines = 100

	var calls atomic.Int64
	start := make(chan struct{})
	results := make([]string, goroutines)

	var wg sync.WaitGroup
	wg.Add(goroutines)
	for i := 0; i < goroutines; i++ {
		go func() {
			defer wg.Done()
			<-start // všichni vyrazí naráz, ať je souběh co nejtvrdší
			results[i] = c.GetOrCompute("drahý", func() string {
				calls.Add(1)
				time.Sleep(20 * time.Millisecond)
				return "spočítáno"
			})
		}()
	}
	close(start)
	wg.Wait()

	if got := calls.Load(); got != 1 {
		t.Errorf("výpočetní funkce zavolána %dx, chci právě 1x", got)
	}
	for i, got := range results {
		if got != "spočítáno" {
			t.Fatalf("goroutina %d dostala %q, chci %q", i, got, "spočítáno")
		}
	}
	if got, ok := c.Get("drahý"); !ok || got != "spočítáno" {
		t.Errorf("Get po GetOrCompute = (%q, %v), chci (\"spočítáno\", true)", got, ok)
	}
}

func TestCacheGetOrComputeUsesCachedValue(t *testing.T) {
	c := exercise.NewCache()
	c.Set("k", "z cache")

	got := c.GetOrCompute("k", func() string {
		t.Error("GetOrCompute zavolal f, přestože klíč v cache byl")
		return "nové"
	})
	if got != "z cache" {
		t.Errorf("GetOrCompute = %q, chci %q", got, "z cache")
	}
}

func TestCacheGetOrComputeDifferentKeys(t *testing.T) {
	c := exercise.NewCache()
	var calls atomic.Int64

	var wg sync.WaitGroup
	wg.Add(20)
	for i := 0; i < 20; i++ {
		key := fmt.Sprintf("k%d", i%5)
		go func() {
			defer wg.Done()
			c.GetOrCompute(key, func() string {
				calls.Add(1)
				time.Sleep(10 * time.Millisecond)
				return key
			})
		}()
	}
	wg.Wait()

	if got := calls.Load(); got != 5 {
		t.Errorf("výpočet proběhl %dx, chci 5x (jednou na klíč)", got)
	}
	if got := c.Len(); got != 5 {
		t.Errorf("Len() = %d, chci 5", got)
	}
}

func TestBankTransferErrors(t *testing.T) {
	b := exercise.NewBank(map[string]int64{"a": 100, "b": 50})

	tests := []struct {
		name             string
		from, to         string
		amount           int64
		want             error
		wantBalanceFromA int64
	}{
		{"neznámý zdroj", "x", "a", 10, exercise.ErrUnknownAccount, 100},
		{"neznámý cíl", "a", "x", 10, exercise.ErrUnknownAccount, 100},
		{"nulová částka", "a", "b", 0, exercise.ErrInvalidAmount, 100},
		{"záporná částka", "a", "b", -5, exercise.ErrInvalidAmount, 100},
		{"nedostatek prostředků", "a", "b", 1000, exercise.ErrInsufficientFunds, 100},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := b.Transfer(tt.from, tt.to, tt.amount)
			if !errors.Is(err, tt.want) {
				t.Fatalf("Transfer(%q, %q, %d) = %v, chci %v", tt.from, tt.to, tt.amount, err, tt.want)
			}
			if got, _ := b.Balance("a"); got != tt.wantBalanceFromA {
				t.Errorf("zůstatek a = %d, chci %d — neúspěšný převod nesmí nic změnit", got, tt.wantBalanceFromA)
			}
		})
	}
}

func TestBankTransferHappyPath(t *testing.T) {
	b := exercise.NewBank(map[string]int64{"a": 100, "b": 50})

	if err := b.Transfer("a", "b", 30); err != nil {
		t.Fatalf("Transfer vrátil %v, chci nil", err)
	}
	if got, _ := b.Balance("a"); got != 70 {
		t.Errorf("zůstatek a = %d, chci 70", got)
	}
	if got, _ := b.Balance("b"); got != 80 {
		t.Errorf("zůstatek b = %d, chci 80", got)
	}
	if err := b.Transfer("a", "a", 10); err != nil {
		t.Errorf("Transfer na stejný účet = %v, chci nil", err)
	}
	if got, _ := b.Balance("a"); got != 70 {
		t.Errorf("zůstatek a po převodu na sebe = %d, chci 70", got)
	}
	if _, ok := b.Balance("neexistuje"); ok {
		t.Error("Balance neexistujícího účtu vrátil ok = true")
	}
	if got := b.Total(); got != 150 {
		t.Errorf("Total() = %d, chci 150", got)
	}
}

func TestBankCrossTransfersKeepTotalAndDoNotDeadlock(t *testing.T) {
	const accounts, perAccount = 6, 1000
	balances := make(map[string]int64, accounts)
	for i := 0; i < accounts; i++ {
		balances[fmt.Sprintf("acc-%d", i)] = perAccount
	}
	b := exercise.NewBank(balances)
	want := int64(accounts * perAccount)

	var wg sync.WaitGroup
	for i := 0; i < accounts; i++ {
		for j := 0; j < accounts; j++ {
			if i == j {
				continue
			}
			from := fmt.Sprintf("acc-%d", i)
			to := fmt.Sprintf("acc-%d", j)
			wg.Add(1)
			// Křížové převody v obou směrech: bez pevného pořadí zámků
			// se tenhle test zaklesne.
			go func() {
				defer wg.Done()
				for k := 0; k < 200; k++ {
					_ = b.Transfer(from, to, 1)
				}
			}()
		}
	}

	// Souběžně s převody čteme celkovou sumu — nikdy nesmí být jiná.
	stop := make(chan struct{})
	readerDone := make(chan struct{})
	go func() {
		defer close(readerDone)
		for {
			select {
			case <-stop:
				return
			default:
				if got := b.Total(); got != want {
					t.Errorf("Total() = %d, chci %d — převod není atomický", got, want)
					return
				}
				time.Sleep(time.Millisecond)
			}
		}
	}()

	finished := make(chan struct{})
	go func() {
		wg.Wait()
		close(finished)
	}()
	select {
	case <-finished:
	case <-time.After(30 * time.Second):
		t.Fatal("převody se zaklesly (deadlock) — zamykáš účty v proměnném pořadí?")
	}
	close(stop)
	<-readerDone

	if got := b.Total(); got != want {
		t.Errorf("Total() po převodech = %d, chci %d", got, want)
	}
	sum := int64(0)
	for i := 0; i < accounts; i++ {
		bal, ok := b.Balance(fmt.Sprintf("acc-%d", i))
		if !ok {
			t.Fatalf("účet acc-%d zmizel", i)
		}
		if bal < 0 {
			t.Errorf("účet acc-%d má záporný zůstatek %d", i, bal)
		}
		sum += bal
	}
	if sum != want {
		t.Errorf("součet zůstatků = %d, chci %d", sum, want)
	}
}
