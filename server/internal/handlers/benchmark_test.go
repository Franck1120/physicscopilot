package handlers

import (
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"

	"github.com/Franck1120/physicscopilot/server/internal/services"
)

// newBenchApp builds a minimal Fiber app with the routes under test.
// Extracted here to avoid repeating setup in each benchmark.
func newBenchApp() *fiber.App {
	sessionSvc := services.NewSessionService()
	convSvc := services.NewConversationService(sessionSvc, nil, nil)
	ws := NewWSHandler(convSvc, sessionSvc)

	app := fiber.New(fiber.Config{DisableStartupMessage: true})
	app.Get("/health", NewHealthHandler("bench", time.Now(), ws, nil))

	sh := NewSessionHandler(sessionSvc)
	app.Post("/api/sessions", sh.CreateSession)

	return app
}

// BenchmarkHealthHandler measures the throughput of GET /health.
func BenchmarkHealthHandler(b *testing.B) {
	app := newBenchApp()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest(http.MethodGet, "/health", nil)
		resp, err := app.Test(req, -1)
		if err != nil {
			b.Fatalf("request error: %v", err)
		}
		if resp.StatusCode != http.StatusOK {
			b.Fatalf("unexpected status: %d", resp.StatusCode)
		}
	}
}

// BenchmarkSessionCreate measures the throughput of POST /api/sessions.
func BenchmarkSessionCreate(b *testing.B) {
	app := newBenchApp()
	body := `{"device_brand":"Prusa","device_model":"MK4"}`
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest(http.MethodPost, "/api/sessions",
			strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req, -1)
		if err != nil {
			b.Fatalf("request error: %v", err)
		}
		if resp.StatusCode != http.StatusCreated {
			b.Fatalf("unexpected status: %d", resp.StatusCode)
		}
	}
}

// newBenchRAGService creates a RAGService backed by a temporary KB file.
func newBenchRAGService(b *testing.B) *services.RAGService {
	b.Helper()

	const kbJSON = `{"problems":[
		{"id":"clog","name":"Clogged Nozzle","category":"extrusion","description":"Nozzle is blocked causing under-extrusion","visual_symptoms":["thin lines","no filament"],"probable_causes":[],"solutions":[]},
		{"id":"warp","name":"Warping","category":"bed_adhesion","description":"Print corners lift off the bed","visual_symptoms":["lifted corners"],"probable_causes":[],"solutions":[]},
		{"id":"layer","name":"Layer Adhesion","category":"temperature","description":"Layers not bonding properly","visual_symptoms":["weak layers"],"probable_causes":[],"solutions":[]},
		{"id":"stringing","name":"Stringing","category":"retraction","description":"Thin strings between printed parts","visual_symptoms":["thin filament strings"],"probable_causes":[],"solutions":[]},
		{"id":"overextrusion","name":"Over Extrusion","category":"extrusion","description":"Too much filament deposited","visual_symptoms":["blobbing","blobs on surface"],"probable_causes":[],"solutions":[]}
	]}`

	f, err := os.CreateTemp("", "kb-bench-*.json")
	if err != nil {
		b.Fatalf("create temp KB: %v", err)
	}
	f.WriteString(kbJSON) //nolint:errcheck
	f.Close()
	b.Cleanup(func() { os.Remove(f.Name()) })

	b.Setenv("KB_PATH", f.Name())
	svc, err := services.NewRAGService()
	if err != nil {
		b.Fatalf("NewRAGService: %v", err)
	}
	return svc
}

// BenchmarkRAGQuery measures the throughput of RAGService.QueryKB.
func BenchmarkRAGQuery(b *testing.B) {
	svc := newBenchRAGService(b)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		results := svc.QueryKB("clogged nozzle under extrusion", 3)
		if len(results) == 0 {
			b.Fatal("expected non-empty results")
		}
	}
}
