// Package ratelimit obsahuje token bucket bez závislostí mimo standardní knihovnu.
//
// Čas je do bucketu injektovaný funkcí Clock, takže testy nemusí nic prospat
// a jsou deterministické.
package ratelimit

import (
	"context"
	"errors"
	"sync"
	"time"
)

// Chyby konfigurace limiteru.
var (
	// ErrInvalidRate znamená nekladnou rychlost doplňování.
	ErrInvalidRate = errors.New("ratelimit: rate must be positive")
	// ErrInvalidBurst znamená nekladnou kapacitu bucketu.
	ErrInvalidBurst = errors.New("ratelimit: burst must be positive")
)

// Clock vrací aktuální čas. V produkci time.Now, v testech vlastní posuvný čas.
type Clock func() time.Time

// Bucket je token bucket: doplňuje se rychlostí rate tokenů za sekundu
// a pojme nejvýš burst tokenů. Je bezpečný pro souběžné použití.
type Bucket struct {
	rate  float64
	burst float64
	now   Clock

	mu     sync.Mutex
	tokens float64
	last   time.Time
}

// New vytvoří bucket s danou rychlostí (tokenů za sekundu) a kapacitou.
// Bucket začíná plný. Když je clock nil, použije se time.Now.
func New(rate float64, burst int, clock Clock) (*Bucket, error) {
	if rate <= 0 {
		return nil, ErrInvalidRate
	}
	if burst <= 0 {
		return nil, ErrInvalidBurst
	}
	if clock == nil {
		clock = time.Now
	}
	return &Bucket{
		rate:   rate,
		burst:  float64(burst),
		now:    clock,
		tokens: float64(burst),
		last:   clock(),
	}, nil
}

// Allow odebere jeden token, pokud je k dispozici, a vrátí true.
// Nikdy neblokuje.
func (b *Bucket) Allow() bool {
	return b.AllowN(1)
}

// AllowN odebere n tokenů, pokud jsou k dispozici, a vrátí true.
// Pro n <= 0 vrací true a nic neodebírá. Pro n větší než kapacita vrací false.
func (b *Bucket) AllowN(n int) bool {
	if n <= 0 {
		return true
	}
	b.mu.Lock()
	defer b.mu.Unlock()

	b.refillLocked()
	if b.tokens < float64(n) {
		return false
	}
	b.tokens -= float64(n)
	return true
}

// Reserve vrátí, jak dlouho je potřeba počkat na jeden token.
// Nula znamená "hned" a token se rovnou odebere.
func (b *Bucket) Reserve() time.Duration {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.refillLocked()
	if b.tokens >= 1 {
		b.tokens--
		return 0
	}
	// Token si rezervujeme dopředu (tokens jde do mínusu), jinak by si
	// dva souběžní volající naplánovali stejný okamžik.
	missing := 1 - b.tokens
	b.tokens--
	return time.Duration(missing / b.rate * float64(time.Second))
}

// Wait počká na jeden token, nebo skončí s ctx.Err() při zrušení kontextu.
func (b *Bucket) Wait(ctx context.Context) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	wait := b.Reserve()
	if wait <= 0 {
		return nil
	}

	timer := time.NewTimer(wait)
	defer timer.Stop()

	select {
	case <-timer.C:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

// Tokens vrací aktuální počet tokenů (může být i záporný kvůli rezervacím).
func (b *Bucket) Tokens() float64 {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.refillLocked()
	return b.tokens
}

// refillLocked doplní tokeny podle uplynulého času. Volej jen se zamčeným mu.
func (b *Bucket) refillLocked() {
	now := b.now()
	elapsed := now.Sub(b.last)
	if elapsed <= 0 {
		// Clock se nesmí vracet zpět; když se to stane, jen posuneme razítko.
		b.last = now
		return
	}

	b.last = now
	b.tokens += elapsed.Seconds() * b.rate
	if b.tokens > b.burst {
		b.tokens = b.burst
	}
}
