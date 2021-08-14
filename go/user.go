package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/jmoiron/sqlx"
	"goji.io/pat"
)

type ItemMap struct {
	ID                        int64     `json:"id" db:"id"`
	SellerID                  int64     `json:"seller_id" db:"seller_id"`
	BuyerID                   int64     `json:"buyer_id" db:"buyer_id"`
	Status                    string    `json:"status" db:"status"`
	Name                      string    `json:"name" db:"name"`
	Price                     int       `json:"price" db:"price"`
	Description               string    `json:"description" db:"description"`
	ImageName                 string    `json:"image_name" db:"image_name"`
	CategoryID                int       `json:"category_id" db:"category_id"`
	CreatedAt                 time.Time `json:"-" db:"created_at"`
	UpdatedAt                 time.Time `json:"-" db:"updated_at"`
	SellerAccountName         string    `json:"seller_account_name" db:"seller_account_name"`
	SellerNumSellItems        int       `json:"seller_num_sell_items" db:"seller_num_sell_items"`
	BuyerAccountName          string    `json:"buyer_account_name" db:"buyer_account_name"`
	BuyerNumSellItems         int       `json:"buyer_num_sell_items" db:"buyer_num_sell_items"`
	TransactionEvidenceID     int64     `json:"transaction_evidence_id" db:"transaction_evidence_id"`
	TransactionEvidenceStatus string    `json:"transaction_evidence_status" db:"transaction_evidence_status"`
	ReserveID                 string    `json:"reserve_id" db:"reserve_id"`
}

func getUserSimpleByID(q sqlx.Queryer, userID int64) (userSimple UserSimple, err error) {
	user := User{}
	err = sqlx.Get(q, &user, "SELECT * FROM `users` WHERE `id` = ?", userID)
	if err != nil {
		return userSimple, err
	}
	userSimple.ID = user.ID
	userSimple.AccountName = user.AccountName
	userSimple.NumSellItems = user.NumSellItems
	return userSimple, err
}

func getTransactions(w http.ResponseWriter, r *http.Request) {

	user, errCode, errMsg := getUser(r)
	if errMsg != "" {
		outputErrorMsg(w, errCode, errMsg)
		return
	}

	query := r.URL.Query()
	itemIDStr := query.Get("item_id")
	var err error
	var itemID int64
	if itemIDStr != "" {
		itemID, err = strconv.ParseInt(itemIDStr, 10, 64)
		if err != nil || itemID <= 0 {
			outputErrorMsg(w, http.StatusBadRequest, "item_id param error")
			return
		}
	}

	createdAtStr := query.Get("created_at")
	var createdAt int64
	if createdAtStr != "" {
		createdAt, err = strconv.ParseInt(createdAtStr, 10, 64)
		if err != nil || createdAt <= 0 {
			outputErrorMsg(w, http.StatusBadRequest, "created_at param error")
			return
		}
	}

	tx := dbx.MustBegin()
	items := []Item{}
	itemMaps := []ItemMap{}
	if itemID > 0 && createdAt > 0 {
		// paging
		err := tx.Select(&items,
			"SELECT * FROM `items` WHERE (`seller_id` = ? OR `buyer_id` = ?) AND `status` IN (?,?,?,?,?) AND (`created_at` < ?  OR (`created_at` <= ? AND `id` < ?)) ORDER BY `created_at` DESC, `id` DESC LIMIT ?",
			user.ID,
			user.ID,
			ItemStatusOnSale,
			ItemStatusTrading,
			ItemStatusSoldOut,
			ItemStatusCancel,
			ItemStatusStop,
			time.Unix(createdAt, 0),
			time.Unix(createdAt, 0),
			itemID,
			TransactionsPerPage+1,
		)
		if err != nil {
			log.Print(err)
			outputErrorMsg(w, http.StatusInternalServerError, "db error")
			tx.Rollback()
			return
		}
	} else {
		// 1st page
		err := tx.Select(&itemMaps,
			"SELECT i.`id`, i.`seller_id`, i.`buyer_id`, i.`status`, i.`name`, i.`price`, i.`description`, i.`image_name`, i.`category_id`, i.`created_at`, i.`updated_at`, us.`account_name` AS seller_account_name, us.`num_sell_items` AS seller_num_sell_items, ub.`account_name` AS buyer_account_name, ub.`num_sell_items` AS buyer_num_sell_item, t.`id` AS transaction_evidence_id, t.`status` AS transaction_evidence_status, s.`reserve_id` FROM `items` AS i INNER JOIN `users` AS us ON i.`seller_id` = us.`id` INNER JOIN `users` AS ub ON i.`buyer_id` = ub.`id` INNER JOIN `transaction_evidences` AS t ON i.`id` = t.`item_id` INNER JOIN `shippings` AS s ON s.`transaction_evidence_id` = t.`id`  WHERE (i.`seller_id` = ? OR i.`buyer_id` = ?) AND i.`status` IN (?,?,?,?,?) ORDER BY i.`created_at` DESC, i.`id` DESC LIMIT ?",
			user.ID,
			user.ID,
			ItemStatusOnSale,
			ItemStatusTrading,
			ItemStatusSoldOut,
			ItemStatusCancel,
			ItemStatusStop,
			TransactionsPerPage+1,
		)
		if err != nil {
			log.Print(err)
			outputErrorMsg(w, http.StatusInternalServerError, "db error")
			tx.Rollback()
			return
		}
	}

	itemDetails := []ItemDetail{}

	for _, itemMap := range itemMaps {
		// seller, err := getUserSimpleByID(tx, item.SellerID)
		// if err != nil {
		// 	outputErrorMsg(w, http.StatusNotFound, "seller not found")
		// 	tx.Rollback()
		// 	return
		// }
		seller := UserSimple{
			ID:           itemMap.SellerID,
			AccountName:  itemMap.SellerAccountName,
			NumSellItems: itemMap.SellerNumSellItems,
		}
		category := getCategoryByID2(itemMap.CategoryID)
		// if err != nil {
		// 	outputErrorMsg(w, http.StatusNotFound, "category not found")
		// 	tx.Rollback()
		// 	return
		// }

		itemDetail := ItemDetail{
			ID:       itemMap.ID,
			SellerID: itemMap.SellerID,
			Seller:   &seller,
			// BuyerID
			// Buyer
			Status:      itemMap.Status,
			Name:        itemMap.Name,
			Price:       itemMap.Price,
			Description: itemMap.Description,
			ImageURL:    getImageURL(itemMap.ImageName),
			CategoryID:  itemMap.CategoryID,
			// TransactionEvidenceID
			// TransactionEvidenceStatus
			// ShippingStatus
			Category:  &category,
			CreatedAt: itemMap.CreatedAt.Unix(),
		}

		// if item.BuyerID != 0 {
		// 	buyer, err := getUserSimpleByID(tx, item.BuyerID)
		// 	if err != nil {
		// 		outputErrorMsg(w, http.StatusNotFound, "buyer not found")
		// 		tx.Rollback()
		// 		return
		// 	}
		// 	itemDetail.BuyerID = item.BuyerID
		// 	itemDetail.Buyer = &buyer
		// }
		buyer := UserSimple{
			ID:           itemMap.BuyerID,
			AccountName:  itemMap.BuyerAccountName,
			NumSellItems: itemMap.BuyerNumSellItems,
		}
		itemDetail.BuyerID = itemMap.BuyerID
		itemDetail.Buyer = &buyer

		// transactionEvidence := TransactionEvidence{}
		// err = tx.Get(&transactionEvidence, "SELECT * FROM `transaction_evidences` WHERE `item_id` = ?", item.ID)
		// if err != nil && err != sql.ErrNoRows {
		// 	// It's able to ignore ErrNoRows
		// 	log.Print(err)
		// 	outputErrorMsg(w, http.StatusInternalServerError, "db error")
		// 	tx.Rollback()
		// 	return
		// }

		if itemMap.TransactionEvidenceID > 0 {
			// shipping := Shipping{}
			// err = tx.Get(&shipping, "SELECT * FROM `shippings` WHERE `transaction_evidence_id` = ?", transactionEvidence.ID)
			// if err == sql.ErrNoRows {
			// 	outputErrorMsg(w, http.StatusNotFound, "shipping not found")
			// 	tx.Rollback()
			// 	return
			// }
			// if err != nil {
			// 	log.Print(err)
			// 	outputErrorMsg(w, http.StatusInternalServerError, "db error")
			// 	tx.Rollback()
			// 	return
			// }
			ssr, err := APIShipmentStatus(getShipmentServiceURL(), &APIShipmentStatusReq{
				ReserveID: itemMap.ReserveID,
			})
			if err != nil {
				log.Print(err)
				outputErrorMsg(w, http.StatusInternalServerError, "failed to request to shipment service")
				tx.Rollback()
				return
			}

			itemDetail.TransactionEvidenceID = itemMap.TransactionEvidenceID
			itemDetail.TransactionEvidenceStatus = itemMap.TransactionEvidenceStatus
			itemDetail.ShippingStatus = ssr.Status
		}

		itemDetails = append(itemDetails, itemDetail)
	}
	tx.Commit()

	hasNext := false
	if len(itemDetails) > TransactionsPerPage {
		hasNext = true
		itemDetails = itemDetails[0:TransactionsPerPage]
	}

	rts := resTransactions{
		Items:   itemDetails,
		HasNext: hasNext,
	}

	w.Header().Set("Content-Type", "application/json;charset=utf-8")
	json.NewEncoder(w).Encode(rts)

}

func getUserItems(w http.ResponseWriter, r *http.Request) {
	userIDStr := pat.Param(r, "user_id")
	userID, err := strconv.ParseInt(userIDStr, 10, 64)
	if err != nil || userID <= 0 {
		outputErrorMsg(w, http.StatusBadRequest, "incorrect user id")
		return
	}

	userSimple, err := getUserSimpleByID(dbx, userID)
	if err != nil {
		outputErrorMsg(w, http.StatusNotFound, "user not found")
		return
	}

	query := r.URL.Query()
	itemIDStr := query.Get("item_id")
	var itemID int64
	if itemIDStr != "" {
		itemID, err = strconv.ParseInt(itemIDStr, 10, 64)
		if err != nil || itemID <= 0 {
			outputErrorMsg(w, http.StatusBadRequest, "item_id param error")
			return
		}
	}

	createdAtStr := query.Get("created_at")
	var createdAt int64
	if createdAtStr != "" {
		createdAt, err = strconv.ParseInt(createdAtStr, 10, 64)
		if err != nil || createdAt <= 0 {
			outputErrorMsg(w, http.StatusBadRequest, "created_at param error")
			return
		}
	}

	items := []Item{}
	if itemID > 0 && createdAt > 0 {
		// paging
		err := dbx.Select(&items,
			"SELECT * FROM `items` WHERE `seller_id` = ? AND `status` IN (?,?,?) AND (`created_at` < ?  OR (`created_at` <= ? AND `id` < ?)) ORDER BY `created_at` DESC, `id` DESC LIMIT ?",
			userSimple.ID,
			ItemStatusOnSale,
			ItemStatusTrading,
			ItemStatusSoldOut,
			time.Unix(createdAt, 0),
			time.Unix(createdAt, 0),
			itemID,
			ItemsPerPage+1,
		)
		if err != nil {
			log.Print(err)
			outputErrorMsg(w, http.StatusInternalServerError, "db error")
			return
		}
	} else {
		// 1st page
		err := dbx.Select(&items,
			"SELECT * FROM `items` WHERE `seller_id` = ? AND `status` IN (?,?,?) ORDER BY `created_at` DESC, `id` DESC LIMIT ?",
			userSimple.ID,
			ItemStatusOnSale,
			ItemStatusTrading,
			ItemStatusSoldOut,
			ItemsPerPage+1,
		)
		if err != nil {
			log.Print(err)
			outputErrorMsg(w, http.StatusInternalServerError, "db error")
			return
		}
	}

	itemSimples := []ItemSimple{}
	for _, item := range items {
		category, err := getCategoryByID(dbx, item.CategoryID)
		if err != nil {
			outputErrorMsg(w, http.StatusNotFound, "category not found")
			return
		}
		itemSimples = append(itemSimples, ItemSimple{
			ID:         item.ID,
			SellerID:   item.SellerID,
			Seller:     &userSimple,
			Status:     item.Status,
			Name:       item.Name,
			Price:      item.Price,
			ImageURL:   getImageURL(item.ImageName),
			CategoryID: item.CategoryID,
			Category:   &category,
			CreatedAt:  item.CreatedAt.Unix(),
		})
	}

	hasNext := false
	if len(itemSimples) > ItemsPerPage {
		hasNext = true
		itemSimples = itemSimples[0:ItemsPerPage]
	}

	rui := resUserItems{
		User:    &userSimple,
		Items:   itemSimples,
		HasNext: hasNext,
	}

	w.Header().Set("Content-Type", "application/json;charset=utf-8")
	json.NewEncoder(w).Encode(rui)
}
