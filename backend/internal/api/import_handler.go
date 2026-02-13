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
		"clubs":     result.ClubsCount,
	})
}

// ImportResult holds import statistics
type ImportResult struct {
	GeneralsCount  int
	TreasuresCount int
	ClubsCount     int
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

	// Parse generals from "总表" sheet
	// Format: 序号 | 姓名 | 价值 | 统御 | 武力 | 智力 | 政治 | 魅力 | 五维 | 相性 | 枪 | 戟 | 弩 | 骑 | 兵 | ...
	if rows, err := f.GetRows("总表"); err == nil && len(rows) > 1 {
		for _, row := range rows[1:] { // Skip header row
			general := parseGeneralRowV2(row)
			if general != nil {
				// Check if exists by ExcelID
				var existing model.General
				if db.Where("excel_id = ?", general.ExcelID).First(&existing).Error == nil {
					// Update existing
					db.Model(&existing).Updates(general)
				} else {
					// Create new
					db.Create(general)
				}
				result.GeneralsCount++
			}
		}
	}

	// Parse treasures from "宝物" sheet
	// Format: 序号 | 名称 | 种类 | 价值 | 特技 | 属性
	if rows, err := f.GetRows("宝物"); err == nil && len(rows) > 1 {
		for _, row := range rows[1:] { // Skip header row
			treasure := parseTreasureRowV2(row)
			if treasure != nil {
				// Check if exists by ExcelID
				var existing model.Treasure
				if db.Where("excel_id = ?", treasure.ExcelID).First(&existing).Error == nil {
					// Update existing
					db.Model(&existing).Updates(treasure)
				} else {
					// Create new
					db.Create(treasure)
				}
				result.TreasuresCount++
			}
		}
	}

	// Parse clubs from "俱乐部" sheet
	// Complex structure: each club spans multiple rows
	if rows, err := f.GetRows("俱乐部"); err == nil && len(rows) > 4 {
		clubs := parseClubsV2(rows)
		for _, club := range clubs {
			var existing model.Club
			if db.Where("name = ?", club.Name).First(&existing).Error == nil {
				// Update existing
				db.Model(&existing).Updates(club)
			} else {
				// Create new
				db.Create(club)
			}
			result.ClubsCount++
		}
	}

	return result, nil
}

// parseGeneralRowV2 parses a general from Excel row
// Format: 序号 | 姓名 | 价值 | 统御 | 武力 | 智力 | 政治 | 魅力 | 五维 | 相性 | 枪 | 戟 | 弩 | 骑 | 兵 | ...
func parseGeneralRowV2(row []string) *model.General {
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
		return nil // Skip header or empty rows
	}

	// Column C: 价值 (作为薪资)
	salary, _ := strconv.Atoi(strings.TrimSpace(row[2]))

	// Columns D-H: 统御/武力/智力/政治/魅力
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
		PoolType:     "normal", // Default, will be updated based on which sheet
		Tier:         3,        // Default tier
		IsAvailable:  true,
	}

	// Column J: 相性 (index 9)
	if len(row) > 9 && row[9] != "" {
		general.Affinity, _ = strconv.Atoi(strings.TrimSpace(row[9]))
	}

	// Columns K-O: 枪/戟/弩/骑/兵 (indices 10-14)
	if len(row) > 10 && row[10] != "" {
		general.Spear = strings.TrimSpace(row[10])
	}
	if len(row) > 11 && row[11] != "" {
		general.Halberd = strings.TrimSpace(row[11])
	}
	if len(row) > 12 && row[12] != "" {
		general.Crossbow = strings.TrimSpace(row[12])
	}
	if len(row) > 13 && row[13] != "" {
		general.Cavalry = strings.TrimSpace(row[13])
	}
	if len(row) > 14 && row[14] != "" {
		general.Soldier = strings.TrimSpace(row[14])
	}

	// Try to get skills from later columns (might be in different positions)
	// Column P onwards might have skills/特技
	if len(row) > 15 && row[15] != "" {
		general.Skills = strings.TrimSpace(row[15])
	}

	return general
}

// parseTreasureRowV2 parses a treasure from Excel row
// Format: 序号 | 名称 | 种类 | 价值 | 特技 | 属性
func parseTreasureRowV2(row []string) *model.Treasure {
	if len(row) < 3 {
		return nil
	}

	// Column A: 序号 (ExcelID)
	excelID, err := strconv.Atoi(strings.TrimSpace(row[0]))
	if err != nil || excelID <= 0 {
		return nil
	}

	// Column B: 名称
	name := strings.TrimSpace(row[1])
	if name == "" || name == "名称" {
		return nil // Skip header or empty rows
	}

	treasure := &model.Treasure{
		ExcelID:     excelID,
		Name:        name,
		IsAvailable: true,
	}

	// Column C: 种类
	if len(row) > 2 && row[2] != "" {
		treasure.Type = strings.TrimSpace(row[2])
	}

	// Column D: 价值
	if len(row) > 3 && row[3] != "" {
		treasure.Value, _ = strconv.Atoi(strings.TrimSpace(row[3]))
	}

	// Column E: 特技
	if len(row) > 4 && row[4] != "" {
		treasure.Skill = strings.TrimSpace(row[4])
	}

	// Column F: 属性 (如 "统+5")
	if len(row) > 5 && row[5] != "" {
		treasure.Effect = strings.TrimSpace(row[5])
	}

	return treasure
}

// parseClubsV2 parses clubs from the complex "俱乐部" sheet structure
// Structure:
//
//	Row with number: "1", "条件", "效果"  -> marks start of a new club section
//	Next row: "俱乐部名称", "", "基础效果"
//	Following rows: "", "条件N", "效果N"  -> until empty row or next club section
func parseClubsV2(rows [][]string) []*model.Club {
	var clubs []*model.Club

	i := 0
	for i < len(rows) {
		row := rows[i]

		// Look for club name row pattern:
		// First column is club name (Chinese), second column is empty or has content
		// Club names are typically Chinese like "AC米兰", "国际米兰" etc.
		if len(row) > 0 {
			firstCell := strings.TrimSpace(row[0])

			// Check if this is a club name row (not a number, not empty, not "备注", not just "条件")
			if isClubName(firstCell) {
				club := &model.Club{
					Name: firstCell,
				}

				// Get base effect from column C (index 2)
				if len(row) > 2 && row[2] != "" {
					club.Description = strings.TrimSpace(row[2])
				}

				// Collect all policies for this club
				var policies []string
				if club.Description != "" {
					policies = append(policies, "基础效果: "+club.Description)
				}

				// Look at following rows for conditions and effects
				j := i + 1
				for j < len(rows) {
					nextRow := rows[j]
					if len(nextRow) < 3 {
						j++
						continue
					}

					nextFirst := strings.TrimSpace(nextRow[0])

					// Stop if we hit the next club section (number + "条件" + "效果")
					// or another club name
					if isClubName(nextFirst) || isClubSectionHeader(nextRow) {
						break
					}

					// Empty first column means this is a condition row
					if nextFirst == "" || nextFirst == "条件" {
						condition := strings.TrimSpace(nextRow[1])
						effect := ""
						if len(nextRow) > 2 {
							effect = strings.TrimSpace(nextRow[2])
						}

						if condition != "" && condition != "条件" {
							policyLine := fmt.Sprintf("条件: %s => 效果: %s", condition, effect)
							policies = append(policies, policyLine)
						}
					}

					j++
				}

				// Join all policies
				club.Policy = strings.Join(policies, "\n")

				clubs = append(clubs, club)
				i = j
				continue
			}
		}

		i++
	}

	return clubs
}

// isClubName checks if a string looks like a club name
func isClubName(s string) bool {
	s = strings.TrimSpace(s)
	if s == "" {
		return false
	}

	// Skip known non-club values
	skipValues := []string{
		"备注", "条件", "效果", "组别", "1档", "2档", "3档", "4档",
		"玩家", "A", "B", "C", "D", "E", "F", "G", "H",
	}
	for _, skip := range skipValues {
		if s == skip {
			return false
		}
	}

	// Skip if it starts with a number followed by specific patterns
	if _, err := strconv.Atoi(s); err == nil {
		return false
	}

	// Skip if it starts with "1." "2." etc.
	if len(s) > 1 && s[0] >= '1' && s[0] <= '9' && s[1] == '.' {
		return false
	}

	// Club names typically contain Chinese characters
	// and are not purely numeric or single letters
	if len(s) >= 2 {
		// Check if contains Chinese characters
		for _, r := range s {
			if r >= 0x4E00 && r <= 0x9FFF {
				return true
			}
		}
	}

	return false
}

// isClubSectionHeader checks if a row is a club section header like "1", "条件", "效果"
func isClubSectionHeader(row []string) bool {
	if len(row) < 3 {
		return false
	}

	first := strings.TrimSpace(row[0])
	second := strings.TrimSpace(row[1])

	// Check if first is a number and second is "条件"
	if _, err := strconv.Atoi(first); err == nil {
		if second == "条件" {
			return true
		}
	}

	return false
}
