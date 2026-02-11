package sports

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/log"

	"github.com/multimarket-labs/event-pod-services/common/retry"
	"github.com/multimarket-labs/event-pod-services/database"
)

type NBACrawler struct {
	db              *database.DB
	apiKey          string
	accessLevel     string
	languageCode    []string
	baseURL         string
	httpClient      *http.Client
	Category        string
	Ecosystem       string
	Period          string
	defaultLanguage string
}

type SportradarNBAResponse struct {
	Date   string    `json:"date"`
	Games  []NBAGame `json:"games"`
	League NBALeague `json:"league"`
}

type NBAGame struct {
	ID           string       `json:"id"`
	Status       string       `json:"status"` // scheduled, inprogress, closed
	Scheduled    string       `json:"scheduled"`
	HomePoints   *int         `json:"home_points"`
	AwayPoints   *int         `json:"away_points"`
	Coverage     string       `json:"coverage"`
	TrackOnCourt bool         `json:"track_on_court"`
	SRID         string       `json:"sr_id"`
	Reference    string       `json:"reference"`
	TimeZones    NBATimeZones `json:"time_zones"`
	Season       NBASeason    `json:"season"`
	Home         NBATeam      `json:"home"`
	Away         NBATeam      `json:"away"`
}

type NBATimeZones struct {
	Venue string `json:"venue"`
	Home  string `json:"home"`
	Away  string `json:"away"`
}

type NBATeam struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Alias     string `json:"alias"`
	SRID      string `json:"sr_id"`
	Reference string `json:"reference"`
}

type NBALeague struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Alias string `json:"alias"`
}

type NBASeason struct {
	ID   string `json:"id"`
	Year int    `json:"year"`
	Type string `json:"type"`
}

func NewNBACrawler(db *database.DB, config NBACrawlerConfig) (*NBACrawler, error) {
	accessLevel := config.AccessLevel
	if accessLevel == "" {
		accessLevel = "trial"
	}

	baseURL := fmt.Sprintf("https://api.sportradar.com/nba/%s/v8", accessLevel)

	return &NBACrawler{
		db:              db,
		apiKey:          config.ApiKey,
		accessLevel:     accessLevel,
		languageCode:    config.LanguageCode,
		baseURL:         baseURL,
		httpClient:      &http.Client{Timeout: 30 * time.Second},
		Category:        config.CategoryGUID,
		Ecosystem:       config.EcosystemGUID,
		Period:          config.PeriodGUID,
		defaultLanguage: config.LanguageGUID,
	}, nil
}

type NBACrawlerConfig struct {
	AccessLevel   string
	LanguageCode  []string
	CategoryGUID  string
	EcosystemGUID string
	PeriodGUID    string
	LanguageGUID  string
	ApiKey        string
}

func (c *NBACrawler) SyncDailySchedule(ctx context.Context, year, month, day int) error {
	log.Info("Starting NBA schedule sync", "date", fmt.Sprintf("%d-%02d-%02d", year, month, day))

	langCode := "en"
	url := fmt.Sprintf("%s/%s/games/%d/%02d/%02d/schedule.json",
		c.baseURL, langCode, year, month, day)

	retryStrategy := &retry.ExponentialStrategy{
		Min:       1 * time.Second,
		Max:       10 * time.Second,
		MaxJitter: 500 * time.Millisecond,
	}

	var apiResp SportradarNBAResponse

	_, err := retry.Do[interface{}](ctx, 3, retryStrategy, func() (interface{}, error) {
		req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
		if err != nil {
			return nil, fmt.Errorf("failed to create request: %w", err)
		}

		req.Header.Add("accept", "application/json")
		req.Header.Add("x-api-key", c.apiKey)

		// Call API
		resp, err := c.httpClient.Do(req)
		if err != nil {
			log.Warn("API call failed, will retry", "err", err, "url", url)
			return nil, fmt.Errorf("failed to call Sportradar API: %w", err)
		}
		defer resp.Body.Close()

		body, readErr := io.ReadAll(resp.Body)
		if readErr != nil {
			log.Warn("Failed to read response body, will retry", "err", readErr)
			return nil, fmt.Errorf("failed to read response body: %w", readErr)
		}

		if resp.StatusCode != http.StatusOK {
			if resp.StatusCode >= 400 && resp.StatusCode < 500 {
				return nil, fmt.Errorf("API returned client error %d: %s", resp.StatusCode, string(body))
			}
			log.Warn("API returned server error, will retry", "status", resp.StatusCode, "url", url)
			return nil, fmt.Errorf("API returned server error %d: %s", resp.StatusCode, string(body))
		}

		if err := json.Unmarshal(body, &apiResp); err != nil {
			log.Warn("Failed to parse response, will retry", "err", err)
			return nil, fmt.Errorf("failed to decode API response: %w", err)
		}

		return nil, nil
	})

	if err != nil {
		return fmt.Errorf("failed to sync NBA schedule after retries: %w", err)
	}

	log.Info("Retrieved game data", "count", len(apiResp.Games))

	for _, game := range apiResp.Games {
		if err := c.processGame(game, apiResp); err != nil {
			log.Error("Failed to process game", "game_id", game.ID, "err", err)
			continue
		}
	}

	log.Info("NBA schedule sync completed", "date", fmt.Sprintf("%d-%02d-%02d", year, month, day))
	return nil
}

func (c *NBACrawler) processGame(game NBAGame, apiResp SportradarNBAResponse) error {
	return c.db.Transaction(func(txDB *database.DB) error {
		homeTeamGUID, err := c.ensureTeam(txDB, game.Home)
		if err != nil {
			return fmt.Errorf("failed to ensure home team: %w", err)
		}

		awayTeamGUID, err := c.ensureTeam(txDB, game.Away)
		if err != nil {
			return fmt.Errorf("failed to ensure away team: %w", err)
		}

		existingEvent, err := txDB.Event.GetEventByExternalID(game.ID)
		if err != nil {
			return fmt.Errorf("failed to check existing event: %w", err)
		}

		isLive, stage := c.determineGameStatus(game.Status)

		homeScore := "0"
		awayScore := "0"
		if game.HomePoints != nil {
			homeScore = strconv.Itoa(*game.HomePoints)
		}
		if game.AwayPoints != nil {
			awayScore = strconv.Itoa(*game.AwayPoints)
		}

		gameTime := c.parseScheduledTime(game.Scheduled)

		eventPeriodGUID, err := c.ensureEventPeriod(txDB, game.ID, gameTime)
		if err != nil {
			return fmt.Errorf("failed to ensure event period: %w", err)
		}

		eventInfo := database.JSONB{
			"external_id":    game.ID,
			"game_id":        game.ID,
			"sr_id":          game.SRID,
			"reference":      game.Reference,
			"coverage":       game.Coverage,
			"track_on_court": game.TrackOnCourt,
			"league_id":      apiResp.League.ID,
			"league_name":    apiResp.League.Name,
			"season_id":      game.Season.ID,
			"season_year":    game.Season.Year,
			"season_type":    game.Season.Type,
			"scheduled":      game.Scheduled,
			"time_zones": map[string]string{
				"venue": game.TimeZones.Venue,
				"home":  game.TimeZones.Home,
				"away":  game.TimeZones.Away,
			},
			"status": game.Status,
		}

		eventTitle := fmt.Sprintf("%s vs %s", game.Home.Name, game.Away.Name)

		//todo: get team logo
		//logo := c.buildLogoURL(game.Home.Alias, game.Away.Alias)

		if existingEvent != nil {
			eventInfo["open_time"] = gameTime

			updates := map[string]interface{}{
				"main_score":              homeScore,
				"cluster_score":           awayScore,
				"is_live":                 isLive,
				"stage":                   stage,
				"info":                    eventInfo,
				"main_team_group_guid":    homeTeamGUID,
				"cluster_team_group_guid": awayTeamGUID,
				"event_period_guid":       eventPeriodGUID, // Update event_period
			}

			if game.Status == "closed" {
				updates["is_live"] = int16(2) // Closed
				updates["is_online"] = true   // Closed games can be displayed with results
			}

			if err := txDB.Event.UpdateEventFields(existingEvent.GUID, updates); err != nil {
				return fmt.Errorf("failed to update event: %w", err)
			}

			if err := c.updateEventLanguage(txDB, existingEvent.GUID, eventTitle); err != nil {
				log.Warn("Failed to update event language", "err", err)
			}

			log.Info("Updated NBA event", "game_id", game.ID, "event_guid", existingEvent.GUID)
		} else {
			eventInfo["open_time"] = gameTime

			event := &database.Event{
				CategoryGUID:         c.Category,
				EcosystemGUID:        c.Ecosystem,
				EventPeriodGUID:      eventPeriodGUID,
				MainTeamGroupGUID:    homeTeamGUID,
				ClusterTeamGroupGUID: awayTeamGUID,
				ExternalId:           game.ID,
				MainScore:            homeScore,
				ClusterScore:         awayScore,
				Logo:                 "",
				EventType:            0,
				ExperimentResult:     "",
				Info:                 eventInfo,
				IsOnline:             false,
				IsLive:               isLive,
				IsSports:             true,
				Stage:                stage,
			}

			if err := txDB.Event.CreateEvent(event); err != nil {
				return fmt.Errorf("failed to create event: %w", err)
			}

			createdEvent, err := txDB.Event.GetEventByExternalID(game.ID)
			if err != nil {
				return fmt.Errorf("failed to retrieve created event: %w", err)
			}
			if createdEvent == nil {
				return fmt.Errorf("created event not found by external_id: %s", game.ID)
			}

			log.Info("Created NBA event", "game_id", game.ID, "event_guid", createdEvent.GUID)
		}

		return nil
	})
}

func (c *NBACrawler) ensureEventPeriod(txDB *database.DB, gameID string, scheduled string) (string, error) {
	existingPeriod, err := txDB.EventPeriod.GetEventPeriodByCode(gameID)
	if err != nil {
		return "", fmt.Errorf("failed to check existing event period: %w", err)
	}

	if existingPeriod != nil {
		return existingPeriod.GUID, nil
	}

	newPeriod := &database.EventPeriod{
		Code:      gameID,
		IsActive:  true,
		Scheduled: scheduled,
		Remark:    fmt.Sprintf("NBA game date: %s", scheduled),
		Extra:     database.JSONB{},
	}

	if err := txDB.EventPeriod.CreateEventPeriod(newPeriod); err != nil {
		return "", fmt.Errorf("failed to create event period: %w", err)
	}

	createdPeriod, err := txDB.EventPeriod.GetEventPeriodByCode(gameID)
	if err != nil || createdPeriod == nil {
		return "", fmt.Errorf("failed to retrieve created event period: %w", err)
	}

	log.Info("Created event period", "scheduled", scheduled, "guid", createdPeriod.GUID)
	return createdPeriod.GUID, nil
}

func (c *NBACrawler) ensureTeam(txDB *database.DB, team NBATeam) (string, error) {
	nbaEcosystem, err := c.db.Ecosystem.GetEcosystemByCode("NBA")
	if err != nil {
		return "", fmt.Errorf("failed to get NBA ecosystem: %w", err)
	}
	if nbaEcosystem == nil {
		return "", fmt.Errorf("NBA ecosystem not found")
	}

	existingTeam, err := txDB.TeamGroup.GetTeamGroupByExternalId(team.ID, nbaEcosystem.GUID)
	if err != nil {
		return "", fmt.Errorf("failed to check existing team: %w", err)
	}

	if existingTeam != nil {
		return existingTeam.GUID, nil
	}

	teamGroup := &database.TeamGroup{
		EcosystemGUID: nbaEcosystem.GUID,
		ExternalId:    team.ID,
		Logo:          "",
	}

	if err := txDB.TeamGroup.CreateTeamGroup(teamGroup); err != nil {
		return "", fmt.Errorf("failed to create team group: %w", err)
	}

	createdTeam, err := txDB.TeamGroup.GetTeamGroupByGUID(teamGroup.GUID)
	if err != nil || createdTeam == nil {
		return "", fmt.Errorf("failed to retrieve created team: %w", err)
	}

	teamLang := &database.TeamGroupLanguage{
		TeamGroupGUID: createdTeam.GUID,
		LanguageGUID:  c.defaultLanguage,
		Alias:         team.Alias,
		Name:          team.Name,
	}

	if err := txDB.TeamGroup.CreateTeamGroupLanguage(teamLang); err != nil {
		return "", fmt.Errorf("failed to create team language: %w", err)
	}

	log.Info("Created NBA team", "team_id", team.ID, "name", team.Name, "guid", createdTeam.GUID)
	return createdTeam.GUID, nil
}

func (c *NBACrawler) determineGameStatus(status string) (int16, string) {
	switch status {
	case "scheduled":
		return 1, "Q1"
	case "inprogress":
		return 0, "Q1"
	case "closed":
		return 2, "FT"
	default:
		return 1, "Q1"
	}
}

func (c *NBACrawler) parseScheduledTime(scheduled string) string {
	if scheduled == "" {
		return ""
	}

	layouts := []string{
		time.RFC3339,
		time.RFC3339Nano,
		"2006-01-02T15:04:05Z07:00",
		"2006-01-02T15:04:05Z",
	}

	for _, layout := range layouts {
		if t, err := time.Parse(layout, scheduled); err == nil {
			return t.Format("2006-01-02 15:04:05")
		}
	}

	return scheduled
}

func (c *NBACrawler) buildEventRules(game NBAGame) string {
	rules := fmt.Sprintf("NBA %s season game. Home: %s, Away: %s.",
		game.Season.Type, game.Home.Name, game.Away.Name)
	return rules
}

func (c *NBACrawler) buildTeamLogoURL(alias string) string {
	return fmt.Sprintf("https://cdn.sportradar.com/images/nba/teams/%s.png", strings.ToLower(alias))
}

func (c *NBACrawler) updateEventLanguage(txDB *database.DB, eventGUID, title string) error {
	eventLang, err := txDB.Event.GetEventLanguage(eventGUID, c.defaultLanguage)
	if err != nil {
		return err
	}

	if eventLang == nil {
		newEventLang := &database.EventLanguage{
			EventGUID:    eventGUID,
			LanguageGUID: c.defaultLanguage,
			Title:        title,
			Rules:        "",
		}
		return txDB.Event.CreateEventLanguage(newEventLang)
	}

	eventLang.Title = title
	return txDB.Event.UpdateEventLanguage(eventLang)
}

func (c *NBACrawler) SyncToday(ctx context.Context) error {
	now := time.Now()
	return c.SyncDailySchedule(ctx, now.Year(), int(now.Month()), now.Day())
}

func (c *NBACrawler) SyncDateRange(ctx context.Context, startDate, endDate time.Time) error {
	current := startDate
	for !current.After(endDate) {
		if err := c.SyncDailySchedule(ctx, current.Year(), int(current.Month()), current.Day()); err != nil {
			log.Error("同步日期失败", "date", current.Format("2006-01-02"), "err", err)
		}
		current = current.AddDate(0, 0, 1)
	}
	return nil
}
