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

// demoSeries is a series created by the series seed. The first 16 entries are
// scheduled (two per weekday + "completed") so every Daily pill on the home
// page is populated; the rest are unscheduled and feed the Popular / Newly
// Released / Indie rails.
type demoSeries struct {
	Title        string
	Slug         string
	Genre        string
	Category     string
	OverlayTitle string
	ScheduleDay  string
}

// demoComic is a comic created by the comic seed. Category and genre are
// inherited from the series it belongs to (via SeriesIndex).
type demoComic struct {
	Title       string
	Author      string
	Slug        string
	Description string
	Language    string
	AgeRating   string
	IsPremium   bool
	Published   bool
	Tags        []string
	DaysAgo     int
	SeriesIndex int
	SeriesOrder int
}

var demoSeriesList = []demoSeries{
	{Title: "Space Cat Chronicles", Slug: "space-cat-chronicles", Genre: "Action", Category: "Comic", OverlayTitle: "SPACE CAT CHRONICLES", ScheduleDay: "mon"},
	{Title: "Mecha Academy", Slug: "mecha-academy", Genre: "Sci-fi", Category: "Comic", OverlayTitle: "MECHA ACADEMY", ScheduleDay: "mon"},
	{Title: "Neon Shadows", Slug: "neon-shadows", Genre: "Cyberpunk", Category: "Manga", OverlayTitle: "NEON SHADOWS", ScheduleDay: "tue"},
	{Title: "Robot Blues", Slug: "robot-blues", Genre: "Noir", Category: "Comic", OverlayTitle: "ROBOT BLUES", ScheduleDay: "tue"},
	{Title: "Clockwork Hearts", Slug: "clockwork-hearts-series", Genre: "Romance", Category: "Graphic Novel", OverlayTitle: "CLOCKWORK HEARTS", ScheduleDay: "wed"},
	{Title: "The Witch's Brew", Slug: "witchs-brew", Genre: "Slice-of-life", Category: "Webcomic", OverlayTitle: "THE WITCH'S BREW", ScheduleDay: "wed"},
	{Title: "Quantum Detectives", Slug: "quantum-detectives", Genre: "Mystery", Category: "Comic", OverlayTitle: "QUANTUM DETECTIVES", ScheduleDay: "thu"},
	{Title: "Void City", Slug: "void-city", Genre: "Horror", Category: "Comic", OverlayTitle: "VOID CITY", ScheduleDay: "thu"},
	{Title: "Samurai Rabbit", Slug: "samurai-rabbit", Genre: "Action", Category: "Manga", OverlayTitle: "SAMURAI RABBIT", ScheduleDay: "fri"},
	{Title: "Candy Kingdom", Slug: "candy-kingdom", Genre: "Fantasy", Category: "Comic", OverlayTitle: "CANDY KINGDOM", ScheduleDay: "fri"},
	{Title: "Goblin Market", Slug: "goblin-market", Genre: "Drama", Category: "Graphic Novel", OverlayTitle: "THE GOBLIN MARKET", ScheduleDay: "sat"},
	{Title: "The Faerie Court", Slug: "the-faerie-court", Genre: "Fantasy", Category: "Comic", OverlayTitle: "THE FAERIE COURT", ScheduleDay: "sat"},
	{Title: "Solar Punk", Slug: "solar-punk", Genre: "Drama", Category: "Webcomic", OverlayTitle: "SOLAR PUNK 2077", ScheduleDay: "sun"},
	{Title: "The Ocean Deep", Slug: "the-ocean-deep", Genre: "Horror", Category: "Graphic Novel", OverlayTitle: "THE OCEAN DEEP", ScheduleDay: "sun"},
	{Title: "Ice Prison", Slug: "ice-prison", Genre: "Fantasy", Category: "Graphic Novel", OverlayTitle: "ICE PRISON", ScheduleDay: "completed"},
	{Title: "Star Couriers", Slug: "star-couriers", Genre: "Sci-fi", Category: "Webcomic", OverlayTitle: "STAR COURIERS", ScheduleDay: "completed"},
	{Title: "Dungeon Chef", Slug: "dungeon-chef", Genre: "Fantasy", Category: "Webcomic", OverlayTitle: "DUNGEON CHEF", ScheduleDay: ""},
	{Title: "Little Death", Slug: "little-death", Genre: "Fantasy", Category: "Webcomic", OverlayTitle: "LITTLE DEATH", ScheduleDay: ""},
	{Title: "Ghost in the Shell: Digital Exile", Slug: "ghost-in-the-shell-digital-exile", Genre: "Cyberpunk", Category: "Manga", OverlayTitle: "GHOST: DIGITAL EXILE", ScheduleDay: ""},
	{Title: "Starfall Legion", Slug: "starfall-legion", Genre: "Sci-fi", Category: "Comic", OverlayTitle: "STARFALL LEGION", ScheduleDay: ""},
	{Title: "The Last Lighthouse", Slug: "the-last-lighthouse", Genre: "Horror", Category: "Graphic Novel", OverlayTitle: "THE LAST LIGHTHOUSE", ScheduleDay: ""},
	{Title: "Paper & Ink", Slug: "paper-and-ink", Genre: "Slice-of-life", Category: "Webcomic", OverlayTitle: "PAPER & INK", ScheduleDay: ""},
	{Title: "Cobalt Drift", Slug: "cobalt-drift", Genre: "Adventure", Category: "Webcomic", OverlayTitle: "COBALT DRIFT", ScheduleDay: ""},
	{Title: "The Alchemist's Apprentice", Slug: "the-alchemists-apprentice", Genre: "Fantasy", Category: "Webcomic", OverlayTitle: "THE ALCHEMIST'S APPRENTICE", ScheduleDay: ""},
	{Title: "Neon Mirage", Slug: "neon-mirage", Genre: "Cyberpunk", Category: "Manga", OverlayTitle: "NEON MIRAGE", ScheduleDay: ""},
	{Title: "Beneath the Canopy", Slug: "beneath-the-canopy", Genre: "Adventure", Category: "Comic", OverlayTitle: "BENEATH THE CANOPY", ScheduleDay: ""},
	{Title: "The Clockmaker's Daughter", Slug: "the-clockmakers-daughter", Genre: "Romance", Category: "Manhwa", OverlayTitle: "THE CLOCKMAKER'S DAUGHTER", ScheduleDay: ""},
	{Title: "Ember & Ash", Slug: "ember-and-ash", Genre: "Fantasy", Category: "Manhwa", OverlayTitle: "EMBER & ASH", ScheduleDay: ""},
}

var demoComicList = []demoComic{
	// Space Cat Chronicles (mon)
	{Title: "Space Cat: Into the Void", Author: "Captain Whiskers", Slug: "space-cat-into-the-void", Description: "Follow Captain Whiskers as she explores the galaxy", Language: "en", AgeRating: "all_ages", Published: true, Tags: []string{"sci-fi", "cats", "space", "adventure"}, DaysAgo: 15, SeriesIndex: 0, SeriesOrder: 1},
	{Title: "Space Cat: The Kuiper Caper", Author: "Captain Whiskers", Slug: "space-cat-the-kuiper-caper", Description: "A heist at the frozen edge of the solar system", Language: "en", AgeRating: "all_ages", Published: true, Tags: []string{"sci-fi", "cats", "heist", "comedy"}, DaysAgo: 2, SeriesIndex: 0, SeriesOrder: 2},
	{Title: "Space Cat: Return to Meowstron", Author: "Captain Whiskers", Slug: "space-cat-return-to-meowstron", Description: "The feline hero finally faces her origin", Language: "en", AgeRating: "all_ages", Published: true, Tags: []string{"sci-fi", "cats", "adventure"}, DaysAgo: 1, SeriesIndex: 0, SeriesOrder: 3},

	// Mecha Academy (mon)
	{Title: "Mecha Academy: Freshman Year", Author: "Shinji Kido", Slug: "mecha-academy-freshman-year", Description: "Teenage pilots learning to control giant robots", Language: "en", AgeRating: "all_ages", Published: true, Tags: []string{"sci-fi", "mecha", "academy", "action"}, DaysAgo: 7, SeriesIndex: 1, SeriesOrder: 1},
	{Title: "Mecha Academy: Training Day", Author: "Shinji Kido", Slug: "mecha-academy-training-day", Description: "First live-fire exercise in the hangar district", Language: "en", AgeRating: "teen", Published: true, Tags: []string{"sci-fi", "mecha", "action"}, DaysAgo: 6, SeriesIndex: 1, SeriesOrder: 2},
	{Title: "Mecha Academy: The Tournament", Author: "Shinji Kido", Slug: "mecha-academy-the-tournament", Description: "The inter-academy cup pits rival pilots head to head", Language: "en", AgeRating: "teen", IsPremium: true, Published: true, Tags: []string{"sci-fi", "mecha", "tournament"}, DaysAgo: 3, SeriesIndex: 1, SeriesOrder: 3},

	// Neon Shadows (tue)
	{Title: "Neon Shadows: The Awakening", Author: "Yuki Tanaka", Slug: "neon-shadows-the-awakening", Description: "In Neo-Tokyo 2142, one detective fights the system", Language: "ja", AgeRating: "teen", IsPremium: true, Published: true, Tags: []string{"cyberpunk", "noir", "mystery", "sci-fi"}, DaysAgo: 14, SeriesIndex: 2, SeriesOrder: 1},
	{Title: "Neon Shadows: Digital Rain", Author: "Yuki Tanaka", Slug: "neon-shadows-digital-rain", Description: "A data heist unravels a city-wide conspiracy", Language: "ja", AgeRating: "mature", IsPremium: true, Published: true, Tags: []string{"cyberpunk", "noir", "action"}, DaysAgo: 5, SeriesIndex: 2, SeriesOrder: 2},
	{Title: "Neon Shadows: Zero Protocol", Author: "Yuki Tanaka", Slug: "neon-shadows-zero-protocol", Description: "The final gambit against the megacorp", Language: "ja", AgeRating: "mature", IsPremium: true, Published: true, Tags: []string{"cyberpunk", "noir", "thriller"}, DaysAgo: 1, SeriesIndex: 2, SeriesOrder: 3},

	// Robot Blues (tue)
	{Title: "Robot Blues", Author: "Chip Silicon", Slug: "robot-blues", Description: "A noir detective story where the detective is a robot", Language: "en", AgeRating: "teen", Published: true, Tags: []string{"noir", "sci-fi", "detective", "robots"}, DaysAgo: 5, SeriesIndex: 3, SeriesOrder: 1},
	{Title: "Robot Blues: Memory Fault", Author: "Chip Silicon", Slug: "robot-blues-memory-fault", Description: "A corrupted memory bank hides a murder", Language: "en", AgeRating: "teen", Published: true, Tags: []string{"noir", "sci-fi", "detective"}, DaysAgo: 4, SeriesIndex: 3, SeriesOrder: 2},

	// Clockwork Hearts (wed)
	{Title: "Clockwork Hearts", Author: "Ada Lovelace", Slug: "clockwork-hearts", Description: "Victorian steampunk romance with mechanical hearts", Language: "en", AgeRating: "teen", IsPremium: true, Published: true, Tags: []string{"steampunk", "victorian", "mystery", "drama"}, DaysAgo: 13, SeriesIndex: 4, SeriesOrder: 1},
	{Title: "Clockwork Hearts: The Gilded Cage", Author: "Ada Lovelace", Slug: "clockwork-hearts-the-gilded-cage", Description: "A duchess and an automaton plot their escape", Language: "en", AgeRating: "teen", IsPremium: true, Published: true, Tags: []string{"steampunk", "romance", "drama"}, DaysAgo: 3, SeriesIndex: 4, SeriesOrder: 2},
	{Title: "Clockwork Hearts: The Final Gear", Author: "Ada Lovelace", Slug: "clockwork-hearts-the-final-gear", Description: "The clocktower holds the key to everything", Language: "en", AgeRating: "teen", IsPremium: true, Published: true, Tags: []string{"steampunk", "romance"}, DaysAgo: 1, SeriesIndex: 4, SeriesOrder: 3},

	// The Witch's Brew (wed)
	{Title: "The Witch's Brew Cafe", Author: "Elara Moonshadow", Slug: "the-witchs-brew-cafe", Description: "A cozy witch runs a magical cafe for fantasy creatures", Language: "en", AgeRating: "all_ages", Published: true, Tags: []string{"fantasy", "slice-of-life", "cozy", "witches"}, DaysAgo: 6, SeriesIndex: 5, SeriesOrder: 1},
	{Title: "The Witch's Brew: Full Moon Menu", Author: "Elara Moonshadow", Slug: "the-witchs-brew-full-moon-menu", Description: "A lunar festival brings fae customers and trouble", Language: "en", AgeRating: "all_ages", Published: true, Tags: []string{"fantasy", "slice-of-life", "cozy"}, DaysAgo: 4, SeriesIndex: 5, SeriesOrder: 2},

	// Quantum Detectives (thu)
	{Title: "Quantum Detectives: Case File #42", Author: "Artemis Blake", Slug: "quantum-detectives-case-file-42", Description: "Solving crimes across parallel universes", Language: "en", AgeRating: "teen", Published: true, Tags: []string{"sci-fi", "mystery", "detective", "quantum"}, DaysAgo: 12, SeriesIndex: 6, SeriesOrder: 1},
	{Title: "Quantum Detectives: The Entangled Witness", Author: "Artemis Blake", Slug: "quantum-detectives-the-entangled-witness", Description: "A witness exists in two timelines at once", Language: "en", AgeRating: "teen", Published: true, Tags: []string{"sci-fi", "mystery", "quantum"}, DaysAgo: 3, SeriesIndex: 6, SeriesOrder: 2},
	{Title: "Quantum Detectives: Collapse", Author: "Artemis Blake", Slug: "quantum-detectives-collapse", Description: "All timelines converge on a single crime", Language: "en", AgeRating: "mature", IsPremium: true, Published: true, Tags: []string{"sci-fi", "mystery", "thriller"}, DaysAgo: 1, SeriesIndex: 6, SeriesOrder: 3},

	// Void City (thu)
	{Title: "Void City Chronicles", Author: "Luna Nightshade", Slug: "void-city-chronicles", Description: "Dimension-hopping horror in the city between worlds", Language: "en", AgeRating: "mature", Published: true, Tags: []string{"horror", "mystery", "dimensions", "surreal"}, DaysAgo: 8, SeriesIndex: 7, SeriesOrder: 1},
	{Title: "Void City: The Hollow District", Author: "Luna Nightshade", Slug: "void-city-the-hollow-district", Description: "A district where nobody has a reflection", Language: "en", AgeRating: "mature", IsPremium: true, Published: true, Tags: []string{"horror", "surreal"}, DaysAgo: 2, SeriesIndex: 7, SeriesOrder: 2},

	// Samurai Rabbit (fri)
	{Title: "Samurai Rabbit: Path of the Carrot", Author: "Takeshi Miyagi", Slug: "samurai-rabbit-path-of-the-carrot", Description: "Hiyoko the rabbit seeks the legendary golden carrot", Language: "ja", AgeRating: "all_ages", IsPremium: true, Published: true, Tags: []string{"fantasy", "samurai", "animals", "japan"}, DaysAgo: 9, SeriesIndex: 8, SeriesOrder: 1},
	{Title: "Samurai Rabbit: The Bamboo Duel", Author: "Takeshi Miyagi", Slug: "samurai-rabbit-the-bamboo-duel", Description: "A rival ronin challenges Hiyoko to a duel", Language: "ja", AgeRating: "all_ages", Published: true, Tags: []string{"fantasy", "samurai", "action"}, DaysAgo: 3, SeriesIndex: 8, SeriesOrder: 2},
	{Title: "Samurai Rabbit: The Golden Carrot", Author: "Takeshi Miyagi", Slug: "samurai-rabbit-the-golden-carrot", Description: "The search ends where it began", Language: "ja", AgeRating: "all_ages", IsPremium: true, Published: true, Tags: []string{"fantasy", "samurai", "adventure"}, DaysAgo: 1, SeriesIndex: 8, SeriesOrder: 3},

	// Candy Kingdom (fri)
	{Title: "Candy Kingdom: Sweet Rebellion", Author: "Ginger Snap", Slug: "candy-kingdom-sweet-rebellion", Description: "Gummy bears revolt against the tyrannical Licorice King", Language: "en", AgeRating: "all_ages", Published: true, Tags: []string{"fantasy", "comedy", "candy", "rebellion"}, DaysAgo: 3, SeriesIndex: 9, SeriesOrder: 1},
	{Title: "Candy Kingdom: The Licorice Conspiracy", Author: "Ginger Snap", Slug: "candy-kingdom-the-licorice-conspiracy", Description: "A secret plot threatens the sugar mines", Language: "en", AgeRating: "all_ages", Published: true, Tags: []string{"fantasy", "comedy", "mystery"}, DaysAgo: 2, SeriesIndex: 9, SeriesOrder: 2},

	// Goblin Market (sat)
	{Title: "The Goblin Market", Author: "Morgan Darkwood", Slug: "the-goblin-market", Description: "Dark fantasy set in an underground goblin bazaar", Language: "en", AgeRating: "mature", Published: true, Tags: []string{"fantasy", "dark-fantasy", "goblins", "magic"}, DaysAgo: 11, SeriesIndex: 10, SeriesOrder: 1},
	{Title: "Goblin Market: The Midnight Barter", Author: "Morgan Darkwood", Slug: "goblin-market-the-midnight-barter", Description: "A deal struck at midnight carries a terrible price", Language: "en", AgeRating: "mature", IsPremium: true, Published: true, Tags: []string{"fantasy", "dark-fantasy"}, DaysAgo: 2, SeriesIndex: 10, SeriesOrder: 2},

	// The Faerie Court (sat)
	{Title: "The Faerie Court: Changeling's Gambit", Author: "Oberon Wildwood", Slug: "the-faerie-court-changelings-gambit", Description: "Political intrigue in the hidden kingdom of Fae", Language: "en", AgeRating: "teen", IsPremium: true, Published: true, Tags: []string{"fantasy", "faerie", "political", "mystery"}, DaysAgo: 1, SeriesIndex: 11, SeriesOrder: 1},
	{Title: "The Faerie Court: Titania's Edict", Author: "Oberon Wildwood", Slug: "the-faerie-court-titanias-edict", Description: "The Queen's decree shakes the court", Language: "en", AgeRating: "teen", IsPremium: true, Published: true, Tags: []string{"fantasy", "faerie", "political"}, DaysAgo: 0, SeriesIndex: 11, SeriesOrder: 2},

	// Solar Punk (sun)
	{Title: "Solar Punk 2077", Author: "Rei Ayanami", Slug: "solar-punk-2077", Description: "A green utopia turns into a hacker's battleground", Language: "en", AgeRating: "teen", Published: true, Tags: []string{"sci-fi", "solarpunk", "dystopia", "hacker"}, DaysAgo: 10, SeriesIndex: 12, SeriesOrder: 1},
	{Title: "Solar Punk: The Greenhouse", Author: "Rei Ayanami", Slug: "solar-punk-the-greenhouse", Description: "A hidden greenhouse grows more than plants", Language: "en", AgeRating: "teen", Published: true, Tags: []string{"sci-fi", "solarpunk"}, DaysAgo: 1, SeriesIndex: 12, SeriesOrder: 2},

	// The Ocean Deep (sun)
	{Title: "The Ocean Deep", Author: "Howard Stevens", Slug: "the-ocean-deep", Description: "Lovecraftian horror at the bottom of the Mariana Trench", Language: "en", AgeRating: "mature", IsPremium: true, Published: true, Tags: []string{"horror", "ocean", "deep-sea", "lovecraftian"}, DaysAgo: 4, SeriesIndex: 13, SeriesOrder: 1},
	{Title: "The Ocean Deep: Abyssal Choir", Author: "Howard Stevens", Slug: "the-ocean-deep-abyssal-choir", Description: "Something is singing below the trench", Language: "en", AgeRating: "mature", IsPremium: true, Published: true, Tags: []string{"horror", "deep-sea"}, DaysAgo: 1, SeriesIndex: 13, SeriesOrder: 2},

	// Ice Prison (completed)
	{Title: "Ice Prison: Escape from Helheim", Author: "Bjorn Ironhand", Slug: "ice-prison-escape-from-helheim", Description: "Norse warriors trapped in the frozen underworld", Language: "en", AgeRating: "mature", Published: true, Tags: []string{"fantasy", "norse", "mythology", "adventure"}, DaysAgo: 5, SeriesIndex: 14, SeriesOrder: 1},
	{Title: "Ice Prison: The Frost Gate", Author: "Bjorn Ironhand", Slug: "ice-prison-the-frost-gate", Description: "The only way out is through the frozen gate", Language: "en", AgeRating: "mature", Published: true, Tags: []string{"fantasy", "norse", "adventure"}, DaysAgo: 1, SeriesIndex: 14, SeriesOrder: 2},

	// Star Couriers (completed)
	{Title: "Star Couriers: Priority Delivery", Author: "Zephyr Nova", Slug: "star-couriers-priority-delivery", Description: "Intergalactic delivery service in a race against time", Language: "en", AgeRating: "all_ages", Published: true, Tags: []string{"sci-fi", "action", "comedy", "delivery"}, DaysAgo: 2, SeriesIndex: 15, SeriesOrder: 1},
	{Title: "Star Couriers: The Black Hole Route", Author: "Zephyr Nova", Slug: "star-couriers-the-black-hole-route", Description: "The fastest route is also the most dangerous", Language: "en", AgeRating: "all_ages", Published: true, Tags: []string{"sci-fi", "action", "comedy"}, DaysAgo: 1, SeriesIndex: 15, SeriesOrder: 2},

	// Dungeon Chef
	{Title: "Dungeon Chef: Cooking with Monsters", Author: "Gordon Ramsters", Slug: "dungeon-chef-cooking-with-monsters", Description: "A fantasy cook battles monsters for ingredients", Language: "en", AgeRating: "all_ages", Published: false, Tags: []string{"fantasy", "cooking", "comedy", "adventure"}, DaysAgo: 0, SeriesIndex: 16, SeriesOrder: 1},
	{Title: "Dungeon Chef: The Dragon's Pantry", Author: "Gordon Ramsters", Slug: "dungeon-chef-the-dragons-pantry", Description: "The ultimate ingredient lies in a dragon's hoard", Language: "en", AgeRating: "all_ages", Published: true, Tags: []string{"fantasy", "cooking", "comedy"}, DaysAgo: 0, SeriesIndex: 16, SeriesOrder: 2},

	// Little Death
	{Title: "Little Death: A Grim Reaper Story", Author: "Mort O'Brien", Slug: "little-death-a-grim-reaper-story", Description: "A 12-year-old becomes the Grim Reaper's apprentice", Language: "en", AgeRating: "teen", Published: true, Tags: []string{"fantasy", "death", "coming-of-age", "comedy"}, DaysAgo: 0, SeriesIndex: 17, SeriesOrder: 1},

	// Ghost in the Shell: Digital Exile
	{Title: "Ghost in the Shell: Digital Exile", Author: "Masamune Shiro", Slug: "ghost-in-the-shell-digital-exile", Description: "An AI consciousness questions the nature of existence", Language: "ja", AgeRating: "mature", IsPremium: true, Published: true, Tags: []string{"cyberpunk", "sci-fi", "philosophy", "ai"}, DaysAgo: 1, SeriesIndex: 18, SeriesOrder: 1},

	// Starfall Legion
	{Title: "Starfall Legion: First Contact", Author: "Marcus Vale", Slug: "starfall-legion-first-contact", Description: "Humanity's first alien encounter turns hostile", Language: "en", AgeRating: "teen", Published: true, Tags: []string{"sci-fi", "military", "aliens"}, DaysAgo: 1, SeriesIndex: 19, SeriesOrder: 1},
	{Title: "Starfall Legion: The Iron Rain", Author: "Marcus Vale", Slug: "starfall-legion-the-iron-rain", Description: "Drop pods rain down on the last colony", Language: "en", AgeRating: "mature", IsPremium: true, Published: true, Tags: []string{"sci-fi", "military", "action"}, DaysAgo: 0, SeriesIndex: 19, SeriesOrder: 2},

	// The Last Lighthouse
	{Title: "The Last Lighthouse: Keeper's Log", Author: "Ida Brin", Slug: "the-last-lighthouse-keepers-log", Description: "The final keeper on a lighthouse at the edge of the world", Language: "en", AgeRating: "mature", Published: true, Tags: []string{"horror", "isolation", "mystery"}, DaysAgo: 1, SeriesIndex: 20, SeriesOrder: 1},

	// Paper & Ink
	{Title: "Paper & Ink: The Cartoonist's Muse", Author: "Marisol Reyes", Slug: "paper-and-ink-the-cartoonists-muse", Description: "A struggling cartoonist meets an impossible muse", Language: "en", AgeRating: "all_ages", Published: true, Tags: []string{"slice-of-life", "art", "drama"}, DaysAgo: 0, SeriesIndex: 21, SeriesOrder: 1},

	// Cobalt Drift
	{Title: "Cobalt Drift: Beyond the Reef", Author: "Kai Delgado", Slug: "cobalt-drift-beyond-the-reef", Description: "A young diver discovers a sunken kingdom", Language: "en", AgeRating: "all_ages", Published: true, Tags: []string{"adventure", "ocean", "fantasy"}, DaysAgo: 1, SeriesIndex: 22, SeriesOrder: 1},

	// The Alchemist's Apprentice
	{Title: "The Alchemist's Apprentice: First Transmutation", Author: "Greta Voss", Slug: "the-alchemists-apprentice-first-transmutation", Description: "A first spell gone wrong changes everything", Language: "en", AgeRating: "teen", Published: true, Tags: []string{"fantasy", "magic", "adventure"}, DaysAgo: 0, SeriesIndex: 23, SeriesOrder: 1},

	// Neon Mirage
	{Title: "Neon Mirage: Ghost Protocol", Author: "Ren Hoshino", Slug: "neon-mirage-ghost-protocol", Description: "A corporate spy is erased from every database", Language: "ja", AgeRating: "mature", IsPremium: true, Published: false, Tags: []string{"cyberpunk", "thriller", "action"}, DaysAgo: 0, SeriesIndex: 24, SeriesOrder: 1},

	// Beneath the Canopy
	{Title: "Beneath the Canopy: The Whispering Woods", Author: "Elowen Reed", Slug: "beneath-the-canopy-the-whispering-woods", Description: "The forest only speaks to those who listen", Language: "en", AgeRating: "all_ages", Published: true, Tags: []string{"adventure", "fantasy", "nature"}, DaysAgo: 1, SeriesIndex: 25, SeriesOrder: 1},

	// The Clockmaker's Daughter
	{Title: "The Clockmaker's Daughter: The First Gear", Author: "Clara Bennet", Slug: "the-clockmakers-daughter-the-first-gear", Description: "A daughter inherits her father's impossible clock", Language: "en", AgeRating: "teen", Published: true, Tags: []string{"romance", "steampunk", "drama"}, DaysAgo: 0, SeriesIndex: 26, SeriesOrder: 1},

	// Ember & Ash
	{Title: "Ember & Ash: The Cinder Crown", Author: "Jun Park", Slug: "ember-and-ash-the-cinder-crown", Description: "A fallen kingdom rises from its own ashes", Language: "en", AgeRating: "teen", IsPremium: true, Published: true, Tags: []string{"fantasy", "adventure"}, DaysAgo: 0, SeriesIndex: 27, SeriesOrder: 1},
}

var demoUploaders = []string{
	"10000000-0000-0000-0000-000000000002",
	"10000000-0000-0000-0000-000000000003",
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

// seededRange returns a deterministic pseudo-random integer in [min, max)
// derived from seed, so re-seeding produces stable engagement numbers.
func seededRange(seed, min, max int) int {
	if max <= min {
		return min
	}
	x := uint32(seed+1) * 2654435761
	return min + int(x%uint32(max-min))
}

// seedCoverKey returns the Cloudflare Images UUID for a comic/series cover,
// cycling through the seed image pool.
func seedCoverKey(i int) string {
	return seedCoverImages[i%len(seedCoverImages)]
}

// seedPageKeys builds a comic's page key list (12..16 images cycling the pool;
// the leading entries double as the gallery previews).
func seedPageKeys(i int) []string {
	n := 12 + i%5
	keys := make([]string, n)
	for j := 0; j < n; j++ {
		keys[j] = seedCoverImages[(i+j)%len(seedCoverImages)]
	}
	return keys
}

// seedPageDims builds the parallel page-dimension array (2:3 pages).
func seedPageDims(n int) []PageDimension {
	dims := make([]PageDimension, n)
	for j := range dims {
		dims[j] = PageDimension{Width: 800, Height: 1200}
	}
	return dims
}

// seedIdentifiers returns a fake ISBN / UPC / ISSN (or empty) for variety.
func seedIdentifiers(i int) (isbn, upc, issn string) {
	switch i % 4 {
	case 0:
		return fmt.Sprintf("978%010d", 1000000000+i), "", ""
	case 1:
		return "", fmt.Sprintf("%012d", 100000000000+i), ""
	case 2:
		return "", "", fmt.Sprintf("%04d-%04d", 1000+i/100, i%100)
	default:
		return "", "", ""
	}
}

//encore:api public method=POST path=/dev/seed-comics
func DevSeedComics(ctx context.Context, p *SeedParams) (*SeedComicsResponse, error) {
	if !isDevTokenValid(p.Token) {
		return nil, &errs.Error{Code: errs.PermissionDenied, Message: "invalid dev seed token"}
	}

	now := time.Now()

	created := 0
	skipped := 0

	for i, c := range demoComicList {
		id := fmt.Sprintf("20000000-0000-0000-0000-%012x", i+1)
		var exists bool
		db.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM comics WHERE id = $1 OR slug = $2)`, id, c.Slug).Scan(&exists)
		if exists {
			skipped++
			continue
		}

		s := demoSeriesList[c.SeriesIndex]
		category := s.Category
		genre := s.Genre
		readingDirection := "ltr"
		if category == "Manga" || category == "Manhwa" {
			readingDirection = "rtl"
		}

		pageKeys := seedPageKeys(i)
		pageKeysJSON, _ := json.Marshal(pageKeys)
		pageDimsJSON, _ := json.Marshal(seedPageDims(len(pageKeys)))
		tagsJSON, _ := json.Marshal(c.Tags)

		archiveMimetype := "application/vnd.comicbook+zip"
		if i%5 == 0 {
			archiveMimetype = "application/pdf"
		}
		isbn, upc, issn := seedIdentifiers(i)

		status := "published"
		if !c.Published {
			status = "pending_review"
		}

		var publishedAt interface{}
		if c.Published {
			publishedAt = now.AddDate(0, 0, -c.DaysAgo)
		}

		uploaderID := demoUploaders[i%len(demoUploaders)]

		_, err := db.Exec(ctx, `
			INSERT INTO comics (id, uploader_id, title, author, slug, description, content_language, status,
				category, genre, cover_key, file_key, page_keys, page_dimensions, page_count, reading_direction,
				file_size_bytes, age_rating, is_premium, tags, archive_mimetype, isbn, upc, issn,
				published_at, view_count, download_count, like_count, fav_count, dislike_count, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21, $22, $23, $24, $25, $26, $27, $28, $29, $30, $31, $31)
		`, id, uploaderID, c.Title, c.Author, c.Slug, c.Description, c.Language, status,
			category, genre,
			seedCoverKey(i), fmt.Sprintf("seed/%s/archive", c.Slug),
			pageKeysJSON, pageDimsJSON, len(pageKeys), readingDirection,
			1234567, c.AgeRating, c.IsPremium, tagsJSON, archiveMimetype, nulOrValue(isbn), nulOrValue(upc), nulOrValue(issn),
			publishedAt,
			int64(seededRange(i, 100, 12000)), int64(seededRange(i, 20, 800)), seededRange(i, 10, 600), seededRange(i, 5, 200), seededRange(i, 1, 60),
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

type SeedSeriesResponse struct {
	Created int    `json:"created"`
	Skipped int    `json:"skipped"`
	Message string `json:"message"`
}

//encore:api public method=POST path=/dev/seed-series
func DevSeedSeries(ctx context.Context, p *SeedParams) (*SeedSeriesResponse, error) {
	if !isDevTokenValid(p.Token) {
		return nil, &errs.Error{Code: errs.PermissionDenied, Message: "invalid dev seed token"}
	}

	created := 0
	skipped := 0
	for i, s := range demoSeriesList {
		id := fmt.Sprintf("30000000-0000-0000-0000-%012x", i+1)
		var exists bool
		db.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM series WHERE id = $1 OR slug = $2)`, id, s.Slug).Scan(&exists)
		if exists {
			skipped++
			continue
		}

		uploaderID := demoUploaders[i%len(demoUploaders)]

		_, err := db.Exec(ctx, `
			INSERT INTO series (id, title, slug, description, cover_key, genre, category, overlay_title, schedule_day, uploader_id)
			VALUES ($1, $2, $3, '', $4, $5, $6, $7, $8, $9)
		`, id, s.Title, s.Slug, seedCoverKey(i), s.Genre, s.Category, s.OverlayTitle, nulOrValue(s.ScheduleDay), uploaderID)
		if err != nil {
			return nil, err
		}

		// Assign the comics belonging to this series in order.
		for _, c := range demoComicList {
			if c.SeriesIndex != i {
				continue
			}
			db.Exec(ctx, `UPDATE comics SET series_id = $1, series_order = $2 WHERE slug = $3`, id, c.SeriesOrder, c.Slug)
		}

		// Backfill engagement aggregates from the assigned comics (mirrors
		// migration 22 so the home page shows meaningful views/hearts).
		db.Exec(ctx, `
			UPDATE series s SET
				views_count  = COALESCE((SELECT SUM(c.view_count) FROM comics c WHERE c.series_id = s.id), 0),
				hearts_count = COALESCE((SELECT SUM(c.fav_count)  FROM comics c WHERE c.series_id = s.id), 0)
			WHERE s.id = $1
		`, id)

		created++
	}

	return &SeedSeriesResponse{
		Created: created,
		Skipped: skipped,
		Message: fmt.Sprintf("Seeded %d series, skipped %d (already exist)", created, skipped),
	}, nil
}

var _ = myauth.AuthData{}
