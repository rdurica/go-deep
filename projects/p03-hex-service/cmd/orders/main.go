// Command orders spouští službu objednávek.
//
// Tenhle soubor je jediné místo v celé službě, kde se potkají všechny
// adaptéry. Vybírá implementace portů, čte konfiguraci a řeší životní
// cyklus procesu. Žádné doménové rozhodnutí tu nebydlí.
package main

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/rdurica/go-deep/projects/p03-hex-service/internal/app"
	"github.com/rdurica/go-deep/projects/p03-hex-service/internal/httpapi"
	"github.com/rdurica/go-deep/projects/p03-hex-service/internal/memstore"
)

// config je celá konfigurace služby. Čte se jednou, na startu, a dál se
// předává hodnotou — žádné os.Getenv rozeseté po kódu.
type config struct {
	addr            string
	readTimeout     time.Duration
	writeTimeout    time.Duration
	idleTimeout     time.Duration
	shutdownTimeout time.Duration
}

func loadConfig() (config, error) {
	cfg := config{
		addr:            envString("ORDERS_ADDR", ":8080"),
		readTimeout:     5 * time.Second,
		writeTimeout:    10 * time.Second,
		idleTimeout:     60 * time.Second,
		shutdownTimeout: 10 * time.Second,
	}

	var err error
	if cfg.readTimeout, err = envDuration("ORDERS_READ_TIMEOUT", cfg.readTimeout); err != nil {
		return config{}, err
	}
	if cfg.writeTimeout, err = envDuration("ORDERS_WRITE_TIMEOUT", cfg.writeTimeout); err != nil {
		return config{}, err
	}
	if cfg.shutdownTimeout, err = envDuration("ORDERS_SHUTDOWN_TIMEOUT", cfg.shutdownTimeout); err != nil {
		return config{}, err
	}
	return cfg, nil
}

func envString(key, fallback string) string {
	if v := strings.TrimSpace(os.Getenv(key)); v != "" {
		return v
	}
	return fallback
}

func envDuration(key string, fallback time.Duration) (time.Duration, error) {
	raw := strings.TrimSpace(os.Getenv(key))
	if raw == "" {
		return fallback, nil
	}
	d, err := time.ParseDuration(raw)
	if err != nil {
		return 0, fmt.Errorf("%s=%q: %w", key, raw, err)
	}
	if d <= 0 {
		return 0, fmt.Errorf("%s=%q: musí být kladné", key, raw)
	}
	return d, nil
}

// randomID je adaptér portu app.IDGen.
func randomID() string {
	var buf [8]byte
	if _, err := rand.Read(buf[:]); err != nil {
		// crypto/rand selže jen při rozbitém systému; radši padnout
		// hned, než rozdávat předvídatelná ID.
		panic("crypto/rand: " + err.Error())
	}
	return "ord_" + hex.EncodeToString(buf[:])
}

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	if err := run(logger); err != nil {
		logger.Error("služba skončila chybou", slog.Any("err", err))
		os.Exit(1)
	}
}

func run(logger *slog.Logger) error {
	cfg, err := loadConfig()
	if err != nil {
		return fmt.Errorf("konfigurace: %w", err)
	}

	// Ruční wiring. Žádný kontejner, žádná reflexe — je vidět, co na čem
	// závisí, a kompilátor to hlídá.
	repo := memstore.New()
	svc := app.NewService(repo, app.ClockFunc(time.Now), app.IDGenFunc(randomID))
	handler := httpapi.NewHandler(svc)

	srv := &http.Server{
		Addr:              cfg.addr,
		Handler:           handler,
		ReadHeaderTimeout: cfg.readTimeout,
		ReadTimeout:       cfg.readTimeout,
		WriteTimeout:      cfg.writeTimeout,
		IdleTimeout:       cfg.idleTimeout,
	}

	// NotifyContext zruší kontext při SIGINT/SIGTERM. Druhý signál už
	// obsluhu nemá, takže netrpělivý operátor proces zabije natvrdo.
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	errc := make(chan error, 1)
	go func() {
		logger.Info("server startuje", slog.String("addr", cfg.addr))
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errc <- err
			return
		}
		errc <- nil
	}()

	select {
	case err := <-errc:
		return err
	case <-ctx.Done():
		logger.Info("dostal jsem signál, ukončuji", slog.Duration("timeout", cfg.shutdownTimeout))
	}

	// Graceful shutdown: doběhnou rozpracované požadavky, nové se
	// nepřijímají. Bez timeoutu by dlouhý požadavek držel proces navždy.
	shutdownCtx, cancel := context.WithTimeout(context.Background(), cfg.shutdownTimeout)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("graceful shutdown: %w", err)
	}
	if err := <-errc; err != nil {
		return err
	}
	logger.Info("server ukončen")
	return nil
}
