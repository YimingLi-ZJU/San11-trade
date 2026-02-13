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
	ExcelID      int       `gorm:"uniqueIndex" json:"excel_id"` // Original ID from Excel (序号)
	Name         string    `gorm:"size:50;not null" json:"name"`
	Command      int       `json:"command"`                    // 统率
	Force        int       `json:"force"`                      // 武力
	Intelligence int       `json:"intelligence"`               // 智力
	Politics     int       `json:"politics"`                   // 政治
	Charm        int       `json:"charm"`                      // 魅力
	Salary       int       `json:"salary"`                     // 薪资/价值
	Affinity     int       `json:"affinity"`                   // 相性
	Spear        string    `gorm:"size:10" json:"spear"`       // 枪适性
	Halberd      string    `gorm:"size:10" json:"halberd"`     // 戟适性
	Crossbow     string    `gorm:"size:10" json:"crossbow"`    // 弩适性
	Cavalry      string    `gorm:"size:10" json:"cavalry"`     // 骑适性
	Soldier      string    `gorm:"size:10" json:"soldier"`     // 兵适性
	Water        string    `gorm:"size:10" json:"water"`       // 水适性
	Skills       string    `gorm:"size:255" json:"skills"`     // 特技
	Morality     string    `gorm:"size:20" json:"morality"`    // 义理
	Ambition     string    `gorm:"size:20" json:"ambition"`    // 野望
	Personality  string    `gorm:"size:20" json:"personality"` // 性格
	Note         string    `gorm:"size:500" json:"note"`       // 改动说明
	PoolType     string    `gorm:"size:20" json:"pool_type"`   // guarantee/normal/draft/second/bigcore/initial_guarantee/initial_normal
	Tier         int       `json:"tier"`                       // Tier level (1-5)
	OwnerID      *uint     `json:"owner_id"`                   // Current owner
	Owner        *User     `gorm:"foreignKey:OwnerID" json:"owner,omitempty"`
	IsAvailable  bool      `gorm:"default:true" json:"is_available"` // Available in pool
	InjuredUntil *int      `json:"injured_until"`                    // Injured until round X
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// Treasure represents an item/treasure in the game
type Treasure struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	ExcelID     int       `gorm:"uniqueIndex" json:"excel_id"` // Original ID from Excel
	Name        string    `gorm:"size:50;not null" json:"name"`
	Type        string    `gorm:"size:20" json:"type"`    // 种类 (短柄/书籍/九鼎等)
	Value       int       `json:"value"`                  // 价值
	Effect      string    `gorm:"size:255" json:"effect"` // 属性效果 (如 统+5)
	Skill       string    `gorm:"size:50" json:"skill"`   // 特技
	OwnerID     *uint     `json:"owner_id"`               // Current owner
	Owner       *User     `gorm:"foreignKey:OwnerID" json:"owner,omitempty"`
	IsAvailable bool      `gorm:"default:true" json:"is_available"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// City represents a city/location in the game
type City struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	ExcelID     int       `json:"excel_id"`                     // 序号
	Name        string    `gorm:"size:50;not null" json:"name"` // 城市名称
	Specialty   string    `gorm:"size:20" json:"specialty"`     // 特产 (马/工/弩等)
	MaxSoldiers int       `json:"max_soldiers"`                 // 最大士兵
	GoldIncome  int       `json:"gold_income"`                  // 金收入
	FoodIncome  int       `json:"food_income"`                  // 粮收入
	Durability  int       `json:"durability"`                   // 耐久
	Tiles       int       `json:"tiles"`                        // 地块数
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// Club represents a club/faction with its policy
type Club struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	ExcelID     int       `json:"excel_id"`                                    // 序号 from Excel
	Name        string    `gorm:"size:50;not null" json:"name"`                // 俱乐部名称
	Description string    `gorm:"size:500" json:"description"`                 // 基础效果
	Policies    []Policy  `gorm:"foreignKey:ClubID" json:"policies,omitempty"` // 国策列表
	BasePrice   int       `json:"base_price"`                                  // Base price for auction
	OwnerID     *uint     `json:"owner_id"`                                    // Current owner
	Owner       *User     `gorm:"foreignKey:OwnerID" json:"owner,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// Policy represents a single policy/strategy of a club
type Policy struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	ClubID    uint      `gorm:"not null;index" json:"club_id"` // 所属俱乐部
	SortOrder int       `json:"sort_order"`                    // 排序顺序
	Condition string    `gorm:"size:500" json:"condition"`     // 条件 (空表示无条件/基础效果)
	Effect    string    `gorm:"size:500" json:"effect"`        // 效果
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// GameRule represents game rules from the rules sheet
type GameRule struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Category  string    `gorm:"size:50" json:"category"`  // 分类 (游戏顺序/小组赛/淘汰赛/资源消耗等)
	Title     string    `gorm:"size:100" json:"title"`    // 标题/事件
	Content   string    `gorm:"type:text" json:"content"` // 详细内容
	SortOrder int       `json:"sort_order"`               // 排序
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
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
	DrawType  string    `gorm:"size:20" json:"draw_type"` // guarantee/normal/initial_guarantee/initial_normal
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

// AuctionRecord records each auction result
type AuctionRecord struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	GeneralID uint      `gorm:"not null;uniqueIndex" json:"general_id"` // Each auction general can only be auctioned once
	General   General   `gorm:"foreignKey:GeneralID" json:"general"`
	UserID    *uint     `json:"user_id"` // Winner user ID, null means unsold (流拍)
	User      *User     `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Price     int       `gorm:"default:0" json:"price"`         // Auction price (space cost)
	IsUnsold  bool      `gorm:"default:false" json:"is_unsold"` // True if no one bid (流拍)
	Remark    string    `gorm:"size:200" json:"remark"`         // Optional remark
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// InviteCode represents an invitation code for registration
type InviteCode struct {
	ID        uint       `gorm:"primaryKey" json:"id"`
	Code      string     `gorm:"uniqueIndex;size:32;not null" json:"code"` // 32-character unique code
	Type      int        `gorm:"default:0" json:"type"`                    // 0=single-use, 1=multi-use
	MaxUses   int        `gorm:"default:1" json:"max_uses"`                // Maximum number of uses
	UsedCount int        `gorm:"default:0" json:"used_count"`              // Number of times used
	ExpiredAt *time.Time `json:"expired_at"`                               // Expiration time, null=never expires
	CreatedBy uint       `json:"created_by"`                               // Creator (admin) ID
	Creator   *User      `gorm:"foreignKey:CreatedBy" json:"creator,omitempty"`
	Remark    string     `gorm:"size:200" json:"remark"` // Optional remark/note
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
}

// InviteCodeUsage records each use of an invite code
type InviteCodeUsage struct {
	ID           uint        `gorm:"primaryKey" json:"id"`
	InviteCodeID uint        `gorm:"index;not null" json:"invite_code_id"`
	InviteCode   *InviteCode `gorm:"foreignKey:InviteCodeID" json:"invite_code,omitempty"`
	UserID       uint        `gorm:"index;not null" json:"user_id"`
	User         *User       `gorm:"foreignKey:UserID" json:"user,omitempty"`
	UsedAt       time.Time   `json:"used_at"`
}

// IsValid checks if the invite code is still valid
func (ic *InviteCode) IsValid() bool {
	// Check if expired
	if ic.ExpiredAt != nil && time.Now().After(*ic.ExpiredAt) {
		return false
	}
	// Check if usage limit reached
	if ic.UsedCount >= ic.MaxUses {
		return false
	}
	return true
}
