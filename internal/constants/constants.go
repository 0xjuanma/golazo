package constants

import "time"

// MainViewCheckDelay is the delay before navigating to a selected view in the main menu.
// Set to 1.5 seconds to allow API preloading while showing transition animation.
const MainViewCheckDelay = 1500 * time.Millisecond
