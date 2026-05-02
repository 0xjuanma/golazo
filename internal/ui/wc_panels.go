package ui

// wc_panels.go is a thin compatibility shim that re-exports the World Cup UI
// components from the internal/ui/worldcup package. All rendering logic now
// lives in that package; this file preserves the existing call sites in
// internal/app/view.go without requiring changes there.

import (
	"github.com/0xjuanma/golazo/internal/api"
	"github.com/0xjuanma/golazo/internal/constants"
	"github.com/0xjuanma/golazo/internal/ui/worldcup"
	"github.com/charmbracelet/bubbles/list"
)

// WCGroupItem re-exported from the worldcup package so that app/wc_handlers.go
// can continue to use ui.WCGroupItem without an import change.
type WCGroupItem = worldcup.WCGroupItem

// NewWCGroupDelegate creates the styled list delegate for WC group items.
func NewWCGroupDelegate() list.DefaultDelegate {
	return worldcup.NewWCGroupDelegate()
}

// RenderWorldCupGroups renders the groups overview list view.
func RenderWorldCupGroups(width, height int, wcData *api.WorldCupData, groupsList list.Model, loading bool, lastErr string, bannerType constants.StatusBannerType) string {
	banner := renderStatusBanner(bannerType, width)
	if banner != "" {
		banner += "\n"
	}
	return worldcup.RenderGroupsList(width, height, wcData, groupsList, loading, lastErr, banner)
}

// RenderWorldCupGroupDetail renders the expanded standings for a single group.
func RenderWorldCupGroupDetail(width, height int, wcData *api.WorldCupData, groupIdx int, bannerType constants.StatusBannerType) string {
	banner := renderStatusBanner(bannerType, width)
	if banner != "" {
		banner += "\n"
	}
	return worldcup.RenderGroupDetail(width, height, wcData, groupIdx, banner)
}

// RenderWorldCupGroupGrid renders the all-groups grid overview.
func RenderWorldCupGroupGrid(width, height int, wcData *api.WorldCupData, selectedGroupIdx int, bannerType constants.StatusBannerType) string {
	banner := renderStatusBanner(bannerType, width)
	if banner != "" {
		banner += "\n"
	}
	return worldcup.RenderGroupGrid(width, height, wcData, selectedGroupIdx, banner)
}

// RenderWorldCupBracket renders the knockout bracket with box-drawing connectors.
func RenderWorldCupBracket(width, height int, wcData *api.WorldCupData, scrollOffset int, bannerType constants.StatusBannerType) string {
	banner := renderStatusBanner(bannerType, width)
	if banner != "" {
		banner += "\n"
	}
	return worldcup.RenderBracket(width, height, wcData, scrollOffset, banner)
}
