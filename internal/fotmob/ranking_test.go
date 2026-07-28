package fotmob

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestFIFARanking(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/data/fifarankings/period":
			_, _ = w.Write([]byte(`[{"periodId":"20260401","periodName":"01.04.2026"}]`))
		case "/data/fifarankings/ranking":
			if got := r.URL.Query().Get("periodId"); got != "20260401" {
				t.Errorf("periodId = %q", got)
			}
			_, _ = w.Write([]byte(`[{"name":"Spain","id":6720,"rank":1,"totalPoints":1875.37,"previousPoints":1877.18,"pointsDiff":-1.81,"gainedRank":true,"lostRank":false}]`))
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	client := NewClient()
	client.baseURL = server.URL
	client.httpClient = server.Client()
	got, err := client.FIFARanking(context.Background(), "men")
	if err != nil {
		t.Fatal(err)
	}
	if got.Period.ID != "20260401" || len(got.Teams) != 1 {
		t.Fatalf("unexpected ranking: %#v", got)
	}
	if got.Teams[0].Name != "Spain" || got.Teams[0].PointsDiff != -1.81 {
		t.Fatalf("unexpected row: %#v", got.Teams[0])
	}
}

func TestFIFARankingRejectsInvalidGender(t *testing.T) {
	if _, err := NewClient().FIFARanking(context.Background(), "other"); err == nil {
		t.Fatal("expected invalid gender error")
	}
}
