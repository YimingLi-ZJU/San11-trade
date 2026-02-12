package model

import (
	"time"

	"gorm.io/gorm"
)

// User represents a player or admin
type User struct {
	ID           uint           `gorm:"primaryKey" json:"id"`
	Username     string         `gorm:"uniqueIndex;size:50;not null" json:"username"`
	Password     string         `gorm:"size:255;not null" json:"-"`
	Nickname     string         `gorm:"size:50" json:"nickname"`
	IsAdmin      bool           `gorm:"default:false" json:"is_admin"`
	IsRegistered bool           `gorm:"default:false" json:"is_registered"` // Has signed up for the league
	Space        int            `gorm:"default:350" json:"space"`           // Available space for generals
	UsedSpace    int            `gorm:"default:0" json:"used_space"`        // Used space
	ClubID       *uint          `json:"club_id"`                            // Selected club
	Club         *Club          `gorm:"foreignKey:ClubID" json:"club,omitempty"`
	Generals     []General      `gorm:"many2many:player_generals;" json:"generals,omitempty"`
	Treasures    []Treasure     `gorm:"many2many:player_treasures;" json:"treasures,omitempty"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`
}

// General represents a warrior/general in the game
type General struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	Name         string    `gorm:"size:50;not null" json:"name"`
	Command      int       `json:"command"`                  // 统率
	Force        int       `json:"force"`                    // 武力
	Intelligence int       `json:"intelligence"`             // 智力
	Politics     int       `json:"politics"`                 // 政治
	Charm        int       `json:"charm"`                    // 魅力
	Salary       int       `json:"salary"`                   // 薪资/空间占用
	PoolType     string    `gorm:"size:20" json:"pool_type"` // guarantee/normal/draft/second/bigcore
	Tier         int       `json:"tier"`                     // Tier level (1-5)
	Skills       string    `gorm:"size:255" json:"skills"`   // Comma-separated skills
	OwnerID      *uint     `json:"owner_id"`                 // Current owner
	Owner        *User     `gorm:"foreignKey:OwnerID" json:"owner,omitempty"`
	IsAvailable  bool      `gorm:"default:true" json:"is_available"` // Available in pool
	InjuredUntil *int      `json:"injured_until"`                    // Injured until round X
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// Treasure represents an item/treasure in the game
type Treasure struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	Name        string    `gorm:"size:50;not null" json:"name"`
	Type        string    `gorm:"size:20" json:"type"`    // weapon/book/horse/accessory
	Value       int       `json:"value"`                  // Value/price
	Effect      string    `gorm:"size:255" json:"effect"` // Effect description
	Skill       string    `gorm:"size:50" json:"skill"`   // Special skill granted
	OwnerID     *uint     `json:"owner_id"`               // Current owner
	Owner       *User     `gorm:"foreignKey:OwnerID" json:"owner,omitempty"`
	IsAvailable bool      `gorm:"default:true" json:"is_available"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// Club represents a club/faction with its policy
type Club struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	Name        string    `gorm:"size:50;not null" json:"name"`
	Description string    `gorm:"size:500" json:"description"`
	Policy      string    `gorm:"type:text" json:"policy"` // Policy description in JSON or text
	BasePrice   int       `json:"base_price"`              // Base price for auction
	OwnerID     *uint     `json:"owner_id"`                // Current owner
	Owner       *User     `gorm:"foreignKey:OwnerID" json:"owner,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// Trade represents a trade proposal between two players
type Trade struct {
	ID               uint      `gorm:"primaryKey" json:"id"`
	ProposerID       uint      `gorm:"not null" json:"proposer_id"`
	Proposer         User      `gorm:"foreignKey:ProposerID" json:"proposer"`
	ReceiverID       uint      `gorm:"not null" json:"receiver_id"`
	Receiver         User      `gorm:"foreignKey:ReceiverID" json:"receiver"`
	OfferGenerals    string    `gorm:"type:text" json:"offer_generals"`       // JSON array of general IDs
	OfferTreasures   string    `gorm:"type:text" json:"offer_treasures"`      // JSON array of treasure IDs
	OfferSpace       int       `json:"offer_space"`                           // Space offered
	RequestGenerals  string    `gorm:"type:text" json:"request_generals"`     // JSON array of general IDs
	RequestTreasures string    `gorm:"type:text" json:"request_treasures"`    // JSON array of treasure IDs
	RequestSpace     int       `json:"request_space"`                         // Space requested
	Status           string    `gorm:"size:20;default:pending" json:"status"` // pending/accepted/rejected/cancelled
	Message          string    `gorm:"size:500" json:"message"`               // Optional message
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

// GamePhase represents the current phase of the game
type GamePhase struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	CurrentPhase string    `gorm:"size:30;not null" json:"current_phase"` // signup/guarantee_draw/normal_draw/draft/trading/match
	RoundNumber  int       `gorm:"default:1" json:"round_number"`         // Current round number
	DraftRound   int       `gorm:"default:0" json:"draft_round"`          // Current draft round (1-4)
	DraftOrder   string    `gorm:"type:text" json:"draft_order"`          // JSON array of user IDs in draft order
	Config       string    `gorm:"type:text" json:"config"`               // Additional configuration in JSON
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// DrawRecord records each draw action
type DrawRecord struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	UserID    uint      `gorm:"not null" json:"user_id"`
	User      User      `gorm:"foreignKey:UserID" json:"user"`
	GeneralID uint      `gorm:"not null" json:"general_id"`
	General   General   `gorm:"foreignKey:GeneralID" json:"general"`
	DrawType  string    `gorm:"size:20" json:"draw_type"` // guarantee/normal
	CreatedAt time.Time `json:"created_at"`
}

// DraftRecord records each draft pick
type DraftRecord struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	UserID    uint      `gorm:"not null" json:"user_id"`
	User      User      `gorm:"foreignKey:UserID" json:"user"`
	GeneralID uint      `gorm:"not null" json:"general_id"`
	General   General   `gorm:"foreignKey:GeneralID" json:"general"`
	Round     int       `json:"round"` // Draft round (1-4)
	Pick      int       `json:"pick"`  // Pick number in the round
	CreatedAt time.Time `json:"created_at"`
}

// TradeLog records trade history for audit
type TradeLog struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	TradeID     uint      `gorm:"not null" json:"trade_id"`
	Action      string    `gorm:"size:20" json:"action"` // created/accepted/rejected/cancelled
	PerformedBy uint      `json:"performed_by"`
	Details     string    `gorm:"type:text" json:"details"` // JSON details
	CreatedAt   time.Time `json:"created_at"`
}
