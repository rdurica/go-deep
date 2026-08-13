package exercise_test

import (
	"errors"
	"fmt"
	"sync"
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
				_ = c.Value()
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
	c.Set("a", "3")

	if got, ok := c.Get("a"); !ok || got != "3" {
		t.Errorf("Get(\"a\") = (%q, %v), chci (\"3\", true)", got, ok)
	}
	if got := c.Len(); got != 2 {
		t.Errorf("Len() = %d, chci 2", got)
	}

	c.Delete("a")
	c.Delete("neexistuje")
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

func TestBankTransferErrors(t *testing.T) {
	b := exercise.NewBank(map[string]int64{"a": 100, "b": 50})

	tests := []struct {
		name             string
		from, to         string
		amount           int64
		want             error
		wantBalanceFromA int64
	}{
		{"unknown source", "x", "a", 10, exercise.ErrUnknownAccount, 100},
		{"unknown destination", "a", "x", 10, exercise.ErrUnknownAccount, 100},
		{"zero amount", "a", "b", 0, exercise.ErrInvalidAmount, 100},
		{"negative amount", "a", "b", -5, exercise.ErrInvalidAmount, 100},
		{"insufficient funds", "a", "b", 1000, exercise.ErrInsufficientFunds, 100},
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
		t.Errorf("součet zůstatků = %d, chci 150", got)
	}
}

func TestBankCrossTransfersKeepTotalAndDoNotDeadlock(t *testing.T) {
	const accounts, perAccount = 6, 1000
	balances := make(map[string]int64, accounts)
	names := make([]string, accounts)
	for i := 0; i < accounts; i++ {
		name := fmt.Sprintf("acc-%d", i)
		names[i] = name
		balances[name] = perAccount
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
			go func() {
				defer wg.Done()
				for k := 0; k < 200; k++ {
					_ = b.Transfer(from, to, 1)
				}
			}()
		}
	}

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
					t.Errorf("součet zůstatků = %d, chci %d — převod není atomický", got, want)
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
		t.Errorf("součet zůstatků po převodech = %d, chci %d", got, want)
	}
	for _, name := range names {
		bal, ok := b.Balance(name)
		if !ok {
			t.Fatalf("účet %s zmizel", name)
		}
		if bal < 0 {
			t.Errorf("účet %s má záporný zůstatek %d", name, bal)
		}
	}
}
