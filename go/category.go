package main

import "github.com/jmoiron/sqlx"

type Category struct {
	ID                 int    `json:"id" db:"id"`
	ParentID           int    `json:"parent_id" db:"parent_id"`
	CategoryName       string `json:"category_name" db:"category_name"`
	ParentCategoryName string `json:"parent_category_name,omitempty" db:"-"`
}

type SimpleCategory struct {
	ParentID     int    `json:"parent_id" db:"parent_id"`
	CategoryName string `json:"category_name" db:"category_name"`
}

func getCategoryByID(q sqlx.Queryer, categoryID int) (category Category, err error) {
	err = sqlx.Get(q, &category, "SELECT * FROM `categories` WHERE `id` = ?", categoryID)
	if category.ParentID != 0 {
		parentCategory, err := getCategoryByID(q, category.ParentID)
		if err != nil {
			return category, err
		}
		category.ParentCategoryName = parentCategory.CategoryName
	}
	return category, err
}

func getCategoryByID2(categoryID int) (category Category) {
	c := categories[categoryID]
	if c.ParentID != 0 {
		p := categories[c.ParentID]
		return Category{
			ID:                 categoryID,
			ParentID:           c.ParentID,
			CategoryName:       c.CategoryName,
			ParentCategoryName: p.CategoryName,
		}
	}
	return Category{
		ID:           categoryID,
		ParentID:     c.ParentID,
		CategoryName: c.CategoryName,
	}
}

var categories = map[int]SimpleCategory{
	1:  {0, "ソファー"},
	2:  {1, "一人掛けソファー"},
	3:  {1, "二人掛けソファー"},
	4:  {1, "コーナーソファー"},
	5:  {1, "二段ソファー"},
	6:  {1, "ソファーベッド"},
	10: {0, "家庭用チェア"},
	11: {10, "スツール"},
	12: {10, "クッションスツール"},
	13: {10, "ダイニングチェア"},
	14: {10, "リビングチェア"},
	15: {10, "カウンターチェア"},
	20: {0, "キッズチェア"},
	21: {20, "学習チェア"},
	22: {20, "ベビーソファ"},
	23: {20, "キッズハイチェア"},
	24: {20, "テーブルチェア"},
	30: {0, "オフィスチェア"},
	31: {30, "デスクチェア"},
	32: {30, "ビジネスチェア"},
	33: {30, "回転チェア"},
	34: {30, "リクライニングチェア"},
	35: {30, "投擲用椅子"},
	40: {0, "折りたたみ椅子"},
	41: {40, "パイプ椅子"},
	42: {40, "木製折りたたみ椅子"},
	43: {40, "キッチンチェア"},
	44: {40, "アウトドアチェア"},
	45: {40, "作業椅子"},
	50: {0, "ベンチ"},
	51: {50, "一人掛けベンチ"},
	52: {50, "二人掛けベンチ"},
	53: {50, "アウトドア用ベンチ"},
	54: {50, "収納付きベンチ"},
	55: {50, "背もたれ付きベンチ"},
	56: {50, "ベンチマーク"},
	60: {0, "座椅子"},
	61: {60, "和風座椅子"},
	62: {60, "高座椅子"},
	63: {60, "ゲーミング座椅子"},
	64: {60, "ロッキングチェア"},
	65: {60, "座布団"},
	66: {60, "空気椅子"},
}
