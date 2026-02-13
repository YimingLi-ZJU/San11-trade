package api

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"san11-trade/internal/database"
	"san11-trade/internal/model"

	"github.com/gin-gonic/gin"
	"github.com/xuri/excelize/v2"
	"gorm.io/gorm"
)

// ImportData handles Excel data import (admin only)
func ImportData(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "please upload an Excel file"})
		return
	}

	// Save uploaded file temporarily
	tempPath := filepath.Join(os.TempDir(), file.Filename)
	if err := c.SaveUploadedFile(file, tempPath); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save file"})
		return
	}
	defer os.Remove(tempPath)

	// Parse Excel file
	result, err := parseExcelFile(tempPath)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("failed to parse Excel: %v", err)})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":   "数据导入成功",
		"generals":  result.GeneralsCount,
		"treasures": result.TreasuresCount,
		"cities":    result.CitiesCount,
		"clubs":     result.ClubsCount,
		"policies":  result.PoliciesCount,
		"rules":     result.RulesCount,
	})
}

// ImportResult holds import statistics
type ImportResult struct {
	GeneralsCount  int
	TreasuresCount int
	CitiesCount    int
	ClubsCount     int
	PoliciesCount  int
	RulesCount     int
}

// parseExcelFile parses the Excel file and imports data
func parseExcelFile(filePath string) (*ImportResult, error) {
	f, err := excelize.OpenFile(filePath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	result := &ImportResult{}
	db := database.GetDB()

	// 1. Parse generals from "总表" sheet
	// Format: 序号|姓名|价值|统御|武力|智力|政治|魅力|五维|相性|枪|戟|弩|骑|兵|水|特技|义理|野望|性格|统武和|改动
	if rows, err := f.GetRows("总表"); err == nil && len(rows) > 1 {
		for _, row := range rows[1:] { // Skip header row
			general := parseGeneralRow(row)
			if general != nil {
				var existing model.General
				if db.Where("excel_id = ?", general.ExcelID).First(&existing).Error == nil {
					db.Model(&existing).Updates(general)
				} else {
					db.Create(general)
				}
				result.GeneralsCount++
			}
		}
	}

	// 2. Parse treasures from "宝物" sheet
	// Format: 序号|名称|种类|价值|特技|属性
	if rows, err := f.GetRows("宝物"); err == nil && len(rows) > 1 {
		for _, row := range rows[1:] { // Skip header row
			treasure := parseTreasureRow(row)
			if treasure != nil {
				var existing model.Treasure
				if db.Where("excel_id = ?", treasure.ExcelID).First(&existing).Error == nil {
					db.Model(&existing).Updates(treasure)
				} else {
					db.Create(treasure)
				}
				result.TreasuresCount++
			}
		}
	}

	// 3. Parse cities from "城市" sheet
	// Format: 序号|名称|特产|最大士兵|金收入|粮收入|耐久|地块
	if rows, err := f.GetRows("城市"); err == nil && len(rows) > 1 {
		for _, row := range rows[1:] { // Skip header row
			city := parseCityRow(row)
			if city != nil {
				var existing model.City
				if db.Where("name = ?", city.Name).First(&existing).Error == nil {
					db.Model(&existing).Updates(city)
				} else {
					db.Create(city)
				}
				result.CitiesCount++
			}
		}
	}

	// 4. Parse clubs and policies from "国策" sheet
	if rows, err := f.GetRows("国策"); err == nil && len(rows) > 1 {
		clubs, totalPolicies := parseClubsAndPolicies(rows, db)
		result.ClubsCount = clubs
		result.PoliciesCount = totalPolicies
	}

	// 5. Parse game rules from "规则" sheet
	if rows, err := f.GetRows("规则"); err == nil && len(rows) > 0 {
		result.RulesCount = parseGameRules(rows, db)
	}

	return result, nil
}

// parseGeneralRow parses a general from Excel row
// Format: 序号|姓名|价值|统御|武力|智力|政治|魅力|五维|相性|枪|戟|弩|骑|兵|水|特技|义理|野望|性格|统武和|改动
func parseGeneralRow(row []string) *model.General {
	if len(row) < 8 {
		return nil
	}

	// Column A: 序号 (ExcelID)
	excelID, err := strconv.Atoi(strings.TrimSpace(row[0]))
	if err != nil || excelID <= 0 {
		return nil
	}

	// Column B: 姓名
	name := strings.TrimSpace(row[1])
	if name == "" || name == "姓名" {
		return nil
	}

	// Column C: 价值 (薪资)
	salary, _ := strconv.Atoi(strings.TrimSpace(row[2]))

	// Columns D-H: 统御/武力/智力/政治/魅力 (indices 3-7)
	command, _ := strconv.Atoi(strings.TrimSpace(row[3]))
	force, _ := strconv.Atoi(strings.TrimSpace(row[4]))
	intelligence, _ := strconv.Atoi(strings.TrimSpace(row[5]))
	politics, _ := strconv.Atoi(strings.TrimSpace(row[6]))
	charm, _ := strconv.Atoi(strings.TrimSpace(row[7]))

	general := &model.General{
		ExcelID:      excelID,
		Name:         name,
		Salary:       salary,
		Command:      command,
		Force:        force,
		Intelligence: intelligence,
		Politics:     politics,
		Charm:        charm,
		PoolType:     "normal",
		Tier:         3,
		IsAvailable:  true,
	}

	// Column J: 相性 (index 9)
	if len(row) > 9 && row[9] != "" {
		general.Affinity, _ = strconv.Atoi(strings.TrimSpace(row[9]))
	}

	// Columns K-P: 枪/戟/弩/骑/兵/水 (indices 10-15)
	if len(row) > 10 {
		general.Spear = strings.TrimSpace(row[10])
	}
	if len(row) > 11 {
		general.Halberd = strings.TrimSpace(row[11])
	}
	if len(row) > 12 {
		general.Crossbow = strings.TrimSpace(row[12])
	}
	if len(row) > 13 {
		general.Cavalry = strings.TrimSpace(row[13])
	}
	if len(row) > 14 {
		general.Soldier = strings.TrimSpace(row[14])
	}
	if len(row) > 15 {
		general.Water = strings.TrimSpace(row[15])
	}

	// Column Q: 特技 (index 16)
	if len(row) > 16 {
		general.Skills = strings.TrimSpace(row[16])
	}

	// Column R: 义理 (index 17)
	if len(row) > 17 {
		general.Morality = strings.TrimSpace(row[17])
	}

	// Column S: 野望 (index 18)
	if len(row) > 18 {
		general.Ambition = strings.TrimSpace(row[18])
	}

	// Column T: 性格 (index 19)
	if len(row) > 19 {
		general.Personality = strings.TrimSpace(row[19])
	}

	// Column V: 改动 (index 21)
	if len(row) > 21 {
		general.Note = strings.TrimSpace(row[21])
	}

	return general
}

// parseTreasureRow parses a treasure from Excel row
// Format: 序号|名称|种类|价值|特技|属性
func parseTreasureRow(row []string) *model.Treasure {
	if len(row) < 3 {
		return nil
	}

	excelID, err := strconv.Atoi(strings.TrimSpace(row[0]))
	if err != nil || excelID <= 0 {
		return nil
	}

	name := strings.TrimSpace(row[1])
	if name == "" || name == "名称" {
		return nil
	}

	treasure := &model.Treasure{
		ExcelID:     excelID,
		Name:        name,
		IsAvailable: true,
	}

	if len(row) > 2 {
		treasure.Type = strings.TrimSpace(row[2])
	}
	if len(row) > 3 {
		treasure.Value, _ = strconv.Atoi(strings.TrimSpace(row[3]))
	}
	if len(row) > 4 {
		treasure.Skill = strings.TrimSpace(row[4])
	}
	if len(row) > 5 {
		treasure.Effect = strings.TrimSpace(row[5])
	}

	return treasure
}

// parseCityRow parses a city from Excel row
// Format: 序号|名称|特产|最大士兵|金收入|粮收入|耐久|地块
func parseCityRow(row []string) *model.City {
	if len(row) < 2 {
		return nil
	}

	name := strings.TrimSpace(row[1])
	if name == "" || name == "名称" {
		return nil
	}

	city := &model.City{
		Name: name,
	}

	// 序号可能为空
	if row[0] != "" {
		city.ExcelID, _ = strconv.Atoi(strings.TrimSpace(row[0]))
	}

	if len(row) > 2 {
		city.Specialty = strings.TrimSpace(row[2])
	}
	if len(row) > 3 {
		city.MaxSoldiers, _ = strconv.Atoi(strings.TrimSpace(row[3]))
	}
	if len(row) > 4 {
		city.GoldIncome, _ = strconv.Atoi(strings.TrimSpace(row[4]))
	}
	if len(row) > 5 {
		city.FoodIncome, _ = strconv.Atoi(strings.TrimSpace(row[5]))
	}
	if len(row) > 6 {
		city.Durability, _ = strconv.Atoi(strings.TrimSpace(row[6]))
	}
	if len(row) > 7 {
		city.Tiles, _ = strconv.Atoi(strings.TrimSpace(row[7]))
	}

	return city
}

// parseClubsAndPolicies parses clubs and their policies from "国策" sheet
// Structure:
//
//	Row: "N"(序号)  | "条件" | "效果"     <- 俱乐部段开始标记
//	Row: "俱乐部名" | ""     | "基础效果"  <- 俱乐部名和基础效果
//	Row: ""        | "条件1" | "效果1"    <- 国策条目
//	Row: ""        | "条件2" | "效果2"    <- 国策条目
//	空行 -> 下一个俱乐部
func parseClubsAndPolicies(rows [][]string, db *gorm.DB) (int, int) {
	clubsCount := 0
	policiesCount := 0

	i := 0
	for i < len(rows) {
		row := rows[i]

		// Look for section header: number + "条件" + "效果"
		if len(row) >= 3 && isClubSectionHeader(row) {
			excelID, _ := strconv.Atoi(strings.TrimSpace(row[0]))

			// Next row should be club name + base effect
			i++
			if i >= len(rows) {
				break
			}

			clubRow := rows[i]
			if len(clubRow) < 1 {
				continue
			}

			clubName := strings.TrimSpace(clubRow[0])
			if clubName == "" || !containsChinese(clubName) {
				continue
			}

			baseEffect := ""
			if len(clubRow) > 2 {
				baseEffect = strings.TrimSpace(clubRow[2])
			}

			// Create or update the club
			club := &model.Club{
				ExcelID:     excelID,
				Name:        clubName,
				Description: baseEffect,
			}

			var existingClub model.Club
			if db.Where("name = ?", clubName).First(&existingClub).Error == nil {
				db.Model(&existingClub).Updates(club)
				club.ID = existingClub.ID
			} else {
				db.Create(club)
			}
			clubsCount++

			// Delete old policies for this club
			db.Where("club_id = ?", club.ID).Delete(&model.Policy{})

			// Create base effect as the first policy (if exists)
			sortOrder := 0
			if baseEffect != "" {
				policy := &model.Policy{
					ClubID:    club.ID,
					SortOrder: sortOrder,
					Condition: "", // Empty condition means base effect
					Effect:    baseEffect,
				}
				db.Create(policy)
				policiesCount++
				sortOrder++
			}

			// Parse following policy rows
			i++
			for i < len(rows) {
				policyRow := rows[i]

				// Empty row or new section header means end of this club
				if len(policyRow) < 2 {
					break
				}

				firstCell := strings.TrimSpace(policyRow[0])

				// New club section header
				if isClubSectionHeader(policyRow) {
					break
				}

				// If first cell is a club name (Chinese), it's a new club
				if firstCell != "" && containsChinese(firstCell) {
					break
				}

				// Parse condition and effect
				condition := ""
				effect := ""
				if len(policyRow) > 1 {
					condition = strings.TrimSpace(policyRow[1])
				}
				if len(policyRow) > 2 {
					effect = strings.TrimSpace(policyRow[2])
				}

				// Skip if both are empty or header row
				if (condition == "" && effect == "") || condition == "条件" {
					i++
					continue
				}

				// Create policy
				if condition != "" || effect != "" {
					policy := &model.Policy{
						ClubID:    club.ID,
						SortOrder: sortOrder,
						Condition: condition,
						Effect:    effect,
					}
					db.Create(policy)
					policiesCount++
					sortOrder++
				}

				i++
			}
			continue
		}

		i++
	}

	return clubsCount, policiesCount
}

// parseGameRules parses game rules from "规则" sheet
func parseGameRules(rows [][]string, db *gorm.DB) int {
	// Clear existing rules
	db.Where("1 = 1").Delete(&model.GameRule{})

	count := 0
	currentCategory := ""
	sortOrder := 0

	for _, row := range rows {
		if len(row) < 2 {
			continue
		}

		firstCell := strings.TrimSpace(row[0])
		secondCell := strings.TrimSpace(row[1])

		// Skip empty rows
		if firstCell == "" && secondCell == "" {
			continue
		}

		// Check if this is a category header (e.g., "游戏顺序", "小组赛", "淘汰赛")
		if firstCell != "" && !isNumeric(firstCell) {
			// Could be a category name
			if secondCell == "" || isNumeric(firstCell) == false {
				currentCategory = firstCell
			}
		}

		// Skip pure header rows
		if firstCell == "游戏顺序" || firstCell == "小组赛" || firstCell == "淘汰赛" {
			currentCategory = firstCell
			continue
		}

		// Parse rule content
		title := ""
		content := ""

		if isNumeric(firstCell) {
			// Numbered rule: "1", "事件描述"
			title = secondCell
		} else if firstCell != "" {
			// Named rule: "分组", "内容"
			title = firstCell
			content = secondCell
		}

		if title == "" {
			continue
		}

		// Collect additional content from subsequent columns
		if len(row) > 2 {
			for j := 2; j < len(row); j++ {
				cellVal := strings.TrimSpace(row[j])
				if cellVal != "" {
					if content != "" {
						content += " | "
					}
					content += cellVal
				}
			}
		}

		rule := &model.GameRule{
			Category:  currentCategory,
			Title:     title,
			Content:   content,
			SortOrder: sortOrder,
		}
		db.Create(rule)
		count++
		sortOrder++
	}

	return count
}

// Helper functions

func isClubSectionHeader(row []string) bool {
	if len(row) < 2 {
		return false
	}
	first := strings.TrimSpace(row[0])
	second := strings.TrimSpace(row[1])
	return isNumeric(first) && second == "条件"
}

func isNumeric(s string) bool {
	s = strings.TrimSpace(s)
	if s == "" {
		return false
	}
	_, err := strconv.Atoi(s)
	return err == nil
}

func containsChinese(s string) bool {
	for _, r := range s {
		if r >= 0x4E00 && r <= 0x9FFF {
			return true
		}
	}
	return false
}

// GetCities returns all cities
func GetCities(c *gin.Context) {
	db := database.GetDB()
	var cities []model.City
	if err := db.Order("excel_id asc").Find(&cities).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, cities)
}

// GetGameRules returns all game rules
func GetGameRules(c *gin.Context) {
	db := database.GetDB()
	var rules []model.GameRule
	if err := db.Order("sort_order asc").Find(&rules).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, rules)
}

// GetClubDetail returns a club with its policies
func GetClubDetail(c *gin.Context) {
	id := c.Param("id")
	db := database.GetDB()

	var club model.Club
	if err := db.Preload("Policies", func(db *gorm.DB) *gorm.DB {
		return db.Order("sort_order asc")
	}).Preload("Owner").First(&club, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "club not found"})
		return
	}

	c.JSON(http.StatusOK, club)
}
