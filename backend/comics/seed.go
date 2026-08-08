package comics

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	myauth "comics-galore/backend/auth"

	"encore.dev/beta/errs"
)

type SeedParams struct {
	Token string `json:"token"`
}

type SeedComicsResponse struct {
	Created int    `json:"created"`
	Skipped int    `json:"skipped"`
	Message string `json:"message"`
}

type demoComic struct {
	ID          string
	Title       string
	Author      string
	Slug        string
	Description string
	Language    string
	Status      string
	AgeRating   string
	IsPremium   bool
	Published   bool
	Tags        []string
	DaysAgo     int
}

func isDevTokenValid(token string) bool {
	if token == "" {
		return false
	}
	return token == "dev-secret"
}

func getAuthSeedToken() string {
	return "dev-secret"
}

//encore:api public method=POST path=/dev/seed-comics
func DevSeedComics(ctx context.Context, p *SeedParams) (*SeedComicsResponse, error) {
	if !isDevTokenValid(p.Token) {
		return nil, &errs.Error{Code: errs.PermissionDenied, Message: "invalid dev seed token"}
	}

	uploaderID := "10000000-0000-0000-0000-000000000003"
	now := time.Now()

	demoComics := []demoComic{
		{ID: "20000000-0000-0000-0000-000000000001", Title: "Space Cat: Into the Void", Author: "Captain Whiskers", Slug: "space-cat-into-the-void", Description: "Follow Captain Whiskers as she explores the galaxy", Language: "en", Status: "published", AgeRating: "all_ages", IsPremium: false, Published: true, Tags: []string{"sci-fi", "cats", "space", "adventure"}, DaysAgo: 15},
		{ID: "20000000-0000-0000-0000-000000000002", Title: "Neon Shadows: The Awakening", Author: "Yuki Tanaka", Slug: "neon-shadows-the-awakening", Description: "In Neo-Tokyo 2142, one detective fights the system", Language: "en", Status: "published", AgeRating: "teen", IsPremium: true, Published: true, Tags: []string{"cyberpunk", "noir", "mystery", "sci-fi"}, DaysAgo: 14},
		{ID: "20000000-0000-0000-0000-000000000003", Title: "Dungeon Chef: Cooking with Monsters", Author: "Gordon Ramsters", Slug: "dungeon-chef-cooking-with-monsters", Description: "A fantasy cook battles monsters for ingredients", Language: "en", Status: "pending_review", AgeRating: "all_ages", IsPremium: false, Published: false, Tags: []string{"fantasy", "cooking", "comedy", "adventure"}, DaysAgo: 0},
		{ID: "20000000-0000-0000-0000-000000000004", Title: "Clockwork Hearts", Author: "Ada Lovelace", Slug: "clockwork-hearts", Description: "Victorian steampunk romance with mechanical hearts", Language: "en", Status: "published", AgeRating: "teen", IsPremium: true, Published: true, Tags: []string{"steampunk", "victorian", "mystery", "drama"}, DaysAgo: 13},
		{ID: "20000000-0000-0000-0000-000000000005", Title: "Quantum Detectives: Case File #42", Author: "Artemis Blake", Slug: "quantum-detectives-case-file-42", Description: "Solving crimes across parallel universes", Language: "en", Status: "published", AgeRating: "teen", IsPremium: false, Published: true, Tags: []string{"sci-fi", "mystery", "detective", "quantum"}, DaysAgo: 12},
		{ID: "20000000-0000-0000-0000-000000000006", Title: "The Goblin Market", Author: "Morgan Darkwood", Slug: "the-goblin-market", Description: "Dark fantasy set in an underground goblin bazaar", Language: "en", Status: "published", AgeRating: "mature", IsPremium: false, Published: true, Tags: []string{"fantasy", "dark-fantasy", "goblins", "magic"}, DaysAgo: 11},
		{ID: "20000000-0000-0000-0000-000000000007", Title: "Solar Punk 2077", Author: "Rei Ayanami", Slug: "solar-punk-2077", Description: "A green utopia turns into a hacker's battleground", Language: "en", Status: "published", AgeRating: "teen", IsPremium: false, Published: true, Tags: []string{"sci-fi", "solarpunk", "dystopia", "hacker"}, DaysAgo: 10},
		{ID: "20000000-0000-0000-0000-000000000008", Title: "Samurai Rabbit: Path of the Carrot", Author: "Takeshi Miyagi", Slug: "samurai-rabbit-path-of-the-carrot", Description: "Hiyoko the rabbit seeks the legendary golden carrot", Language: "en", Status: "published", AgeRating: "all_ages", IsPremium: true, Published: true, Tags: []string{"fantasy", "samurai", "animals", "japan"}, DaysAgo: 9},
		{ID: "20000000-0000-0000-0000-000000000009", Title: "Void City Chronicles", Author: "Luna Nightshade", Slug: "void-city-chronicles", Description: "Dimension-hopping horror in the city between worlds", Language: "en", Status: "published", AgeRating: "mature", IsPremium: false, Published: true, Tags: []string{"horror", "mystery", "dimensions", "surreal"}, DaysAgo: 8},
		{ID: "20000000-0000-0000-0000-000000000010", Title: "Mecha Academy: Freshman Year", Author: "Shinji Kido", Slug: "mecha-academy-freshman-year", Description: "Teenage pilots learning to control giant robots", Language: "en", Status: "published", AgeRating: "all_ages", IsPremium: false, Published: true, Tags: []string{"sci-fi", "mecha", "academy", "action"}, DaysAgo: 7},
		{ID: "20000000-0000-0000-0000-000000000011", Title: "The Witch's Brew Cafe", Author: "Elara Moonshadow", Slug: "the-witchs-brew-cafe", Description: "A cozy witch runs a magical cafe for fantasy creatures", Language: "en", Status: "published", AgeRating: "all_ages", IsPremium: false, Published: true, Tags: []string{"fantasy", "slice-of-life", "cozy", "witches"}, DaysAgo: 6},
		{ID: "20000000-0000-0000-0000-000000000012", Title: "Ice Prison: Escape from Helheim", Author: "Bjorn Ironhand", Slug: "ice-prison-escape-from-helheim", Description: "Norse warriors trapped in the frozen underworld", Language: "en", Status: "published", AgeRating: "mature", IsPremium: false, Published: true, Tags: []string{"fantasy", "norse", "mythology", "adventure"}, DaysAgo: 5},
		{ID: "20000000-0000-0000-0000-000000000013", Title: "Robot Blues", Author: "Chip Silicon", Slug: "robot-blues", Description: "A noir detective story where the detective is a robot", Language: "en", Status: "published", AgeRating: "teen", IsPremium: false, Published: true, Tags: []string{"noir", "sci-fi", "detective", "robots"}, DaysAgo: 5},
		{ID: "20000000-0000-0000-0000-000000000014", Title: "The Ocean Deep", Author: "Howard Stevens", Slug: "the-ocean-deep", Description: "Lovecraftian horror at the bottom of the Mariana Trench", Language: "en", Status: "published", AgeRating: "mature", IsPremium: true, Published: true, Tags: []string{"horror", "ocean", "deep-sea", "lovecraftian"}, DaysAgo: 4},
		{ID: "20000000-0000-0000-0000-000000000015", Title: "Candy Kingdom: Sweet Rebellion", Author: "Ginger Snap", Slug: "candy-kingdom-sweet-rebellion", Description: "Gummy bears revolt against the tyrannical Licorice King", Language: "en", Status: "published", AgeRating: "all_ages", IsPremium: false, Published: true, Tags: []string{"fantasy", "comedy", "candy", "rebellion"}, DaysAgo: 3},
		{ID: "20000000-0000-0000-0000-000000000016", Title: "Star Couriers: Priority Delivery", Author: "Zephyr Nova", Slug: "star-couriers-priority-delivery", Description: "Intergalactic delivery service in a race against time", Language: "en", Status: "published", AgeRating: "all_ages", IsPremium: false, Published: true, Tags: []string{"sci-fi", "action", "comedy", "delivery"}, DaysAgo: 2},
		{ID: "20000000-0000-0000-0000-000000000017", Title: "The Faerie Court: Changeling's Gambit", Author: "Oberon Wildwood", Slug: "the-faerie-court-changelings-gambit", Description: "Political intrigue in the hidden kingdom of Fae", Language: "en", Status: "published", AgeRating: "teen", IsPremium: true, Published: true, Tags: []string{"fantasy", "faerie", "political", "mystery"}, DaysAgo: 1},
		{ID: "20000000-0000-0000-0000-000000000018", Title: "Ghost in the Shell: Digital Exile", Author: "Masamune Shiro", Slug: "ghost-in-the-shell-digital-exile", Description: "An AI consciousness questions the nature of existence", Language: "en", Status: "published", AgeRating: "mature", IsPremium: true, Published: true, Tags: []string{"cyberpunk", "sci-fi", "philosophy", "ai"}, DaysAgo: 1},
		{ID: "20000000-0000-0000-0000-000000000019", Title: "Little Death: A Grim Reaper Story", Author: "Mort O'Brien", Slug: "little-death-a-grim-reaper-story", Description: "A 12-year-old becomes the Grim Reaper's apprentice", Language: "en", Status: "published", AgeRating: "teen", IsPremium: false, Published: true, Tags: []string{"fantasy", "death", "coming-of-age", "comedy"}, DaysAgo: 0},
	}

	created := 0
	skipped := 0

	for _, c := range demoComics {
		var exists bool
		db.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM comics WHERE id = $1 OR slug = $2)`, c.ID, c.Slug).Scan(&exists)
		if exists {
			skipped++
			continue
		}

		var pageKeys, tags []byte
		pageKeys, _ = json.Marshal([]string{fmt.Sprintf("seed/%s/page-1", c.Slug)})
		tags, _ = json.Marshal(c.Tags)

		var publishedAt interface{}
		if c.Published {
			publishedAt = now.AddDate(0, 0, -c.DaysAgo)
		}

		_, err := db.Exec(ctx, `
			INSERT INTO comics (id, uploader_id, title, author, slug, description, content_language, status,
				cover_key, file_key, page_keys, file_size_bytes, age_rating, tags,
				published_at, view_count, download_count, like_count, fav_count, dislike_count, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21, $21)
		`, c.ID, uploaderID, c.Title, c.Author, c.Slug, c.Description, c.Language, c.Status,
			fmt.Sprintf("seed/%s/cover", c.Slug), fmt.Sprintf("seed/%s/archive", c.Slug),
			pageKeys, 1234567, c.AgeRating, tags,
			publishedAt,
			randomRange(50, 5000), randomRange(10, 300), randomRange(5, 200), randomRange(3, 80), randomRange(1, 40),
			now.AddDate(0, 0, -c.DaysAgo))
		if err != nil {
			return nil, err
		}
		created++
	}

	return &SeedComicsResponse{
		Created: created,
		Skipped: skipped,
		Message: fmt.Sprintf("Seeded %d comics, skipped %d (already exist)", created, skipped),
	}, nil
}

func randomRange(min, max int) int {
	return min + int(time.Now().UnixNano()%int64(max-min))
}

var _ = myauth.AuthData{}
