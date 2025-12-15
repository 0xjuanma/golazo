# League ID Reference

This document tracks the league IDs used by the FotMob API in this application.

## Supported Leagues

The application currently supports 5 major leagues:
- Premier League (England)
- La Liga (Spain)
- Bundesliga (Germany)
- Serie A (Italy)
- Ligue 1 (France)

## FotMob API League IDs

**Location:** `internal/fotmob/client.go`

```go
SupportedLeagues = []int{
    47, // Premier League
    87, // La Liga
    54, // Bundesliga
    55, // Serie A
    53, // Ligue 1
}
```

**API Endpoint:** `https://www.fotmob.com/api/leagues?id={leagueID}&tab={tab}`

Where `tab` can be:
- `fixtures` - Upcoming matches
- `results` - Finished matches

## Notes

- **FotMob** is used for both the **Live Matches** and **Stats** views
- When adding new leagues, update `internal/fotmob/client.go` and this document

