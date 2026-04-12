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

// newBenchAppFull returns a minimal Fiber app wired with GET /api/sessions for
// pagination and filtering benchmarks. The app is shared across iterations, not
// re-created inside the loop, so only the HTTP dispatch cost is measured.
func newBenchAppFull() (*fiber.App, *services.SessionService) {
	svc := services.NewSessionService()
	sh := NewSessionHandler(svc)

	app := fiber.New(fiber.Config{DisableStartupMessage: true})
	app.Get("/api/sessions", sh.ListSessions)
	app.Post("/api/sessions", sh.CreateSession)
	return app, svc
}

// BenchmarkListSessionsPaginated measures GET /api/sessions?page=1&page_size=10
// with 100 sessions pre-loaded. Only the HTTP dispatch is measured; setup is
// excluded via b.StopTimer / b.StartTimer.
func BenchmarkListSessionsPaginated(b *testing.B) {
	app, svc := newBenchAppFull()

	b.StopTimer()
	for i := 0; i < 100; i++ {
		brand := "Prusa"
		if i%2 == 0 {
			brand = "Creality"
		}
		if _, err := svc.CreateSession(brand, "Model", "", ""); err != nil {
			b.Fatalf("CreateSession: %v", err)
		}
	}
	b.StartTimer()

	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest(http.MethodGet, "/api/sessions?page=1&page_size=10", nil)
		resp, err := app.Test(req, -1)
		if err != nil {
			b.Fatalf("request error: %v", err)
		}
		if resp.StatusCode != http.StatusOK {
			b.Fatalf("unexpected status: %d", resp.StatusCode)
		}
	}
}

// BenchmarkListSessionsFiltered measures GET /api/sessions?device_brand=Prusa
// with 100 sessions loaded (50 Prusa, 50 Creality). Only the HTTP dispatch is
// measured; setup is excluded via b.StopTimer / b.StartTimer.
func BenchmarkListSessionsFiltered(b *testing.B) {
	app, svc := newBenchAppFull()

	b.StopTimer()
	for i := 0; i < 100; i++ {
		brand := "Prusa"
		if i%2 == 0 {
			brand = "Creality"
		}
		if _, err := svc.CreateSession(brand, "Model", "", ""); err != nil {
			b.Fatalf("CreateSession: %v", err)
		}
	}
	b.StartTimer()

	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest(http.MethodGet, "/api/sessions?device_brand=Prusa", nil)
		resp, err := app.Test(req, -1)
		if err != nil {
			b.Fatalf("request error: %v", err)
		}
		if resp.StatusCode != http.StatusOK {
			b.Fatalf("unexpected status: %d", resp.StatusCode)
		}
	}
}

// BenchmarkStatsHandler measures the throughput of GET /api/stats with 10
// pre-created sessions so the handler iterates a non-trivial list.
func BenchmarkStatsHandler(b *testing.B) {
	svc := services.NewSessionService()
	sh := NewStatsHandler(svc)

	app := fiber.New(fiber.Config{DisableStartupMessage: true})
	app.Get("/api/stats", sh.GetStats)

	b.StopTimer()
	for i := 0; i < 10; i++ {
		if _, err := svc.CreateSession("Prusa", "MK4", "", ""); err != nil {
			b.Fatalf("CreateSession: %v", err)
		}
	}
	b.StartTimer()

	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest(http.MethodGet, "/api/stats", nil)
		resp, err := app.Test(req, -1)
		if err != nil {
			b.Fatalf("request error: %v", err)
		}
		if resp.StatusCode != http.StatusOK {
			b.Fatalf("unexpected status: %d", resp.StatusCode)
		}
	}
}

// BenchmarkVersionHandler measures the throughput of GET /version. The handler
// serves a static JSON response baked in at startup, so the benchmark primarily
// measures Fiber routing and serialization overhead.
func BenchmarkVersionHandler(b *testing.B) {
	app := fiber.New(fiber.Config{DisableStartupMessage: true})
	app.Get("/version", VersionHandler("0.1.0", "2026-04-12T00:00:00Z", "go1.25", "dev"))
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest(http.MethodGet, "/version", nil)
		resp, err := app.Test(req, -1)
		if err != nil {
			b.Fatalf("request error: %v", err)
		}
		if resp.StatusCode != http.StatusOK {
			b.Fatalf("unexpected status: %d", resp.StatusCode)
		}
	}
}

// BenchmarkGetSessionSteps measures GET /api/sessions/:id/steps with a session
// that has 3 of 5 steps completed. Setup is excluded via b.StopTimer / b.StartTimer.
func BenchmarkGetSessionSteps(b *testing.B) {
	svc := services.NewSessionService()
	sh := NewSessionHandler(svc)

	app := fiber.New(fiber.Config{DisableStartupMessage: true})
	app.Get("/api/sessions/:id/steps", sh.GetSessionSteps)

	b.StopTimer()
	session, err := svc.CreateSession("Bambu", "X1C", "", "")
	if err != nil {
		b.Fatalf("CreateSession: %v", err)
	}
	if err := svc.UpdateStep(session.SessionID, 3, 5); err != nil {
		b.Fatalf("UpdateStep: %v", err)
	}
	path := "/api/sessions/" + session.SessionID + "/steps"
	b.StartTimer()

	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest(http.MethodGet, path, nil)
		resp, err := app.Test(req, -1)
		if err != nil {
			b.Fatalf("request error: %v", err)
		}
		if resp.StatusCode != http.StatusOK {
			b.Fatalf("unexpected status: %d", resp.StatusCode)
		}
	}
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
