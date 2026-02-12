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

	// Parse generals from "武将" sheet
	if rows, err := f.GetRows("武将"); err == nil && len(rows) > 1 {
		for i, row := range rows[1:] { // Skip header row
			if len(row) < 6 {
				continue
			}

			general := parseGeneralRow(row, i+1)
			if general != nil {
				// Check if exists
				var existing model.General
				if db.Where("name = ?", general.Name).First(&existing).Error == nil {
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

	// Try alternative sheet names
	alternativeSheets := []string{"将领", "全部武将", "武将列表", "武将池"}
	for _, sheetName := range alternativeSheets {
		if rows, err := f.GetRows(sheetName); err == nil && len(rows) > 1 && result.GeneralsCount == 0 {
			for i, row := range rows[1:] {
				if len(row) < 6 {
					continue
				}
				general := parseGeneralRow(row, i+1)
				if general != nil {
					var existing model.General
					if db.Where("name = ?", general.Name).First(&existing).Error == nil {
						db.Model(&existing).Updates(general)
					} else {
						db.Create(general)
					}
					result.GeneralsCount++
				}
			}
			break
		}
	}

	// Parse treasures from "宝物" sheet
	treasureSheets := []string{"宝物", "道具", "物品"}
	for _, sheetName := range treasureSheets {
		if rows, err := f.GetRows(sheetName); err == nil && len(rows) > 1 {
			for _, row := range rows[1:] {
				if len(row) < 3 {
					continue
				}
				treasure := parseTreasureRow(row)
				if treasure != nil {
					var existing model.Treasure
					if db.Where("name = ?", treasure.Name).First(&existing).Error == nil {
						db.Model(&existing).Updates(treasure)
					} else {
						db.Create(treasure)
					}
					result.TreasuresCount++
				}
			}
			break
		}
	}

	// Parse clubs from "俱乐部" or "国策" sheet
	clubSheets := []string{"俱乐部", "国策", "势力"}
	for _, sheetName := range clubSheets {
		if rows, err := f.GetRows(sheetName); err == nil && len(rows) > 1 {
			for _, row := range rows[1:] {
				if len(row) < 2 {
					continue
				}
				club := parseClubRow(row)
				if club != nil {
					var existing model.Club
					if db.Where("name = ?", club.Name).First(&existing).Error == nil {
						db.Model(&existing).Updates(club)
					} else {
						db.Create(club)
					}
					result.ClubsCount++
				}
			}
			break
		}
	}

	return result, nil
}

// parseGeneralRow parses a general from Excel row
// Expected format: 姓名, 统率, 武力, 智力, 政治, 魅力, [薪资], [池类型], [档次], [特技]
func parseGeneralRow(row []string, index int) *model.General {
	if len(row) < 6 || strings.TrimSpace(row[0]) == "" {
		return nil
	}

	name := strings.TrimSpace(row[0])
	command, _ := strconv.Atoi(strings.TrimSpace(row[1]))
	force, _ := strconv.Atoi(strings.TrimSpace(row[2]))
	intelligence, _ := strconv.Atoi(strings.TrimSpace(row[3]))
	politics, _ := strconv.Atoi(strings.TrimSpace(row[4]))
	charm, _ := strconv.Atoi(strings.TrimSpace(row[5]))

	general := &model.General{
		Name:         name,
		Command:      command,
		Force:        force,
		Intelligence: intelligence,
		Politics:     politics,
		Charm:        charm,
		PoolType:     "normal", // Default pool type
		Tier:         3,        // Default tier
		IsAvailable:  true,
	}

	// Parse optional fields
	if len(row) > 6 && row[6] != "" {
		general.Salary, _ = strconv.Atoi(strings.TrimSpace(row[6]))
	} else {
		// Calculate salary based on stats
		general.Salary = calculateSalary(command, force, intelligence, politics, charm)
	}

	if len(row) > 7 && row[7] != "" {
		general.PoolType = normalizePoolType(strings.TrimSpace(row[7]))
	}

	if len(row) > 8 && row[8] != "" {
		general.Tier, _ = strconv.Atoi(strings.TrimSpace(row[8]))
	}

	if len(row) > 9 && row[9] != "" {
		general.Skills = strings.TrimSpace(row[9])
	}

	return general
}

// calculateSalary calculates salary based on stats
func calculateSalary(command, force, intelligence, politics, charm int) int {
	// Simple formula: average of top 3 stats
	stats := []int{command, force, intelligence, politics, charm}
	// Sort descending
	for i := 0; i < len(stats)-1; i++ {
		for j := i + 1; j < len(stats); j++ {
			if stats[j] > stats[i] {
				stats[i], stats[j] = stats[j], stats[i]
			}
		}
	}
	avg := (stats[0] + stats[1] + stats[2]) / 3

	// Map to salary ranges
	switch {
	case avg >= 95:
		return 50
	case avg >= 90:
		return 40
	case avg >= 85:
		return 30
	case avg >= 80:
		return 25
	case avg >= 75:
		return 20
	case avg >= 70:
		return 15
	default:
		return 10
	}
}

// normalizePoolType normalizes pool type names
func normalizePoolType(poolType string) string {
	poolType = strings.ToLower(poolType)
	switch {
	case strings.Contains(poolType, "保底"):
		return "guarantee"
	case strings.Contains(poolType, "普通"):
		return "normal"
	case strings.Contains(poolType, "选秀"):
		return "draft"
	case strings.Contains(poolType, "二抽"):
		return "second"
	case strings.Contains(poolType, "大核"):
		return "bigcore"
	default:
		return poolType
	}
}

// parseTreasureRow parses a treasure from Excel row
// Expected format: 名称, 类型, [价值], [效果], [特技]
func parseTreasureRow(row []string) *model.Treasure {
	if len(row) < 2 || strings.TrimSpace(row[0]) == "" {
		return nil
	}

	treasure := &model.Treasure{
		Name:        strings.TrimSpace(row[0]),
		Type:        strings.TrimSpace(row[1]),
		IsAvailable: true,
	}

	if len(row) > 2 && row[2] != "" {
		treasure.Value, _ = strconv.Atoi(strings.TrimSpace(row[2]))
	}

	if len(row) > 3 && row[3] != "" {
		treasure.Effect = strings.TrimSpace(row[3])
	}

	if len(row) > 4 && row[4] != "" {
		treasure.Skill = strings.TrimSpace(row[4])
	}

	return treasure
}

// parseClubRow parses a club from Excel row
// Expected format: 名称, 描述, [国策], [底价]
func parseClubRow(row []string) *model.Club {
	if len(row) < 1 || strings.TrimSpace(row[0]) == "" {
		return nil
	}

	club := &model.Club{
		Name: strings.TrimSpace(row[0]),
	}

	if len(row) > 1 && row[1] != "" {
		club.Description = strings.TrimSpace(row[1])
	}

	if len(row) > 2 && row[2] != "" {
		club.Policy = strings.TrimSpace(row[2])
	}

	if len(row) > 3 && row[3] != "" {
		club.BasePrice, _ = strconv.Atoi(strings.TrimSpace(row[3]))
	}

	return club
}
