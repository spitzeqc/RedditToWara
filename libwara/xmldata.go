package libwara

import (
	"encoding/xml"
)

const MAX_POSTS uint = 260

type Feeling int

const (
	FEELING_DEFAULT Feeling = iota
	FEELING_DANCE
	FEELING_EXCITED
	FEELING_SHOCKED
	FEELING_ANGRY
	FEELING_SAD
)

// The structure of a single post
type Post struct {
	XMLName                    xml.Name `xml:"post"`
	Body                       string   `xml:"body"`
	CommunityId                int      `xml:"community_id"`
	CountryId                  int      `xml:"country_id"`
	CreatedAt                  string   `xml:"created_at"` // Replace with some sort of marshalled date?
	FeelingId                  Feeling  `xml:"feeling_id"`
	Id                         string   `xml:"id"`                          // Empty?                          // Blank
	IsAutopost                 int      `xml:"is_autopost"`                 // Bool?
	IsCommunityPrivateAutopost int      `xml:"is_communityPrivateAutopost"` // Bool?
	IsSpoiler                  int      `xml:"is_spoiler"`                  // Bool?
	IsAppJumpable              int      `xml:"is_app_jumpable"`             // Bool?
	EmpathyCount               string   `xml:"empathy_count"`               // Blank?
	LanguageId                 int      `xml:"language_id"`
	MiiData                    string   `xml:"mii"`
	MiiFaceUrl                 string   `xml:"mii_face_url"` // Blank
	Number                     int      `xml:"number"`
	PaintingFormat             string   `xml:"painting>format"`
	PaintingContent            string   `xml:"painting>content"` // Blank
	PaintingSize               int      `xml:"painting>size"`
	PaintingUrl                string   `xml:"painting>url"` // Always "http://botu"?
	PID                        string   `xml:"pid"`          // Blank
	PlatformId                 int      `xml:"platform_id"`
	RegionId                   int      `xml:"region_id"`
	ReplyCount                 int      `xml:"reply_count"`
	ScreenName                 string   `xml:"screen_name"`
	TitleId                    string   `xml:"title_id"` // Blank
}

// Structure for a person
type Person struct {
	XMLName xml.Name `xml:"person"`
	Posts   []Post   `xml:"posts>post"`
}

// Structure for a topic
type Topic struct {
	XMLName          xml.Name `xml:"topic"`
	Icon             string   `xml:"icon"`
	TitleId          uint     `xml:"title_id"`
	CommunityId      uint     `xml:"community_id"`
	IsRecommended    int      `xml:"is_recommended"` // Bool
	Name             string   `xml:"name"`
	ParticipantCount uint     `xml:"participant_count"`
	Posts            []Post   `xml:"people>person>posts>post"`
	EmpathyCount     int      `xml:"empathy_count"`
	HasShopPage      int      `xml:"has_shop_page"` // Bool?
	ModifiedAt       string   `xml:"modified_at"`
	Position         int      `xml:"position"` // Always 2?
}

// Structure for the 1stNUP
type Nup struct {
	XMLName     xml.Name `xml:"result"`
	Version     int      `xml:"version"`      // Always 1?
	HasError    int      `xml:"has_error"`    // Always 0?
	RequestName string   `xml:"request_name"` // Always "topics"?
	Expire      string   `xml:"expire"`
	Topics      []Topic  `xml:"topics>topic"`

	totalPosts uint // keep track of the total number of posts in the structure
}

// List of "characters" that need to be unreplaced
type badChar struct {
	Target      string
	Replacement string
}

var replaceTable = []badChar{
	{"&#xA;", "\n"},
	{"&#34;", "\""},
	{"&#39;", "'"},
}

var defaultTitleIds = []uint{
	1407581310509322,
	1407443871760640,
	1407443872637440,
	1407581310505226,
	1407581310492938,
	1407443871834368,
	1407581310501130,
	1407443872632576,
	1407443871756544,
	1407443871768832,
}

var defaultCommunityId uint = 4294967295

// May want to replace with something public domain
var defaultIcon string = "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"
var defaultMii string = "AwEAQAAAAAAAAAAAIAAAAAAAAAAAAAAAABBOAEkATgBUAEUATgBEAE8AAAAAAEYoUQAfCDAkZBQCEkUOUA4fZAwAKC0AWkhQTgBJAE4AVABFAE4ARABPAAAAAAAAAMvX"
