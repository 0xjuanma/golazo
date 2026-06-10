package ui

import (
	"strings"
	"testing"

	"github.com/0xjuanma/golazo/internal/constants"
	"github.com/0xjuanma/golazo/internal/data"
)

func TestRenderStatusBanner_DebugUsesDebugLogPath(t *testing.T) {
	got := renderStatusBanner(constants.StatusBannerDebug, 120)
	want := data.DebugLogPath()

	if !strings.Contains(got, want) {
		t.Errorf("renderStatusBanner(StatusBannerDebug) = %q; want it to contain %q", got, want)
	}
	if !strings.Contains(got, "[DEBUG MODE]") {
		t.Errorf("renderStatusBanner(StatusBannerDebug) = %q; want it to contain %q", got, "[DEBUG MODE]")
	}
}

func TestRenderStatusBanner_NoneReturnsEmpty(t *testing.T) {
	if got := renderStatusBanner(constants.StatusBannerNone, 120); got != "" {
		t.Errorf("renderStatusBanner(StatusBannerNone) = %q; want empty", got)
	}
}
