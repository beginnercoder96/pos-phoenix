package transaction

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"
)

type Item struct {
	ID               int64
	TransactionID    int64
	Category         string
	ItemName         string
	AmountCents      int64
	ItemType         string
	BarberID         int64
	DiscountAmount   int64
	CommissionEarned int64
	BundleID         int64
}

type Entry struct {
	ID, OperatorID, AmountCents        int64
	BranchID                           int64
	ReversalOfID, ReversedByID         sql.NullInt64
	OperatorName, Kind, Category, Note string
	OccurredAt                         time.Time
	Items                              []Item
}
type Summary struct{ IncomeCents, ExpenseCents int64 }
type Page struct {
	Entries            []Entry
	Summary            Summary
	Total, Page, Pages int
}
type Repository struct{ DB *sql.DB }

func (r Repository) Create(ctx context.Context, e Entry) error {
	if e.Kind != "income" && e.Kind != "expense" {
		return errors.New("invalid transaction kind")
	}
	if len(e.Items) > 0 {
		var totalCents int64
		var categories []string
		seen := make(map[string]bool)
		for _, item := range e.Items {
			if item.AmountCents <= 0 {
				return errors.New("each item amount must be positive")
			}
			cat := strings.TrimSpace(item.Category)
			if cat == "" {
				return errors.New("each item category is required")
			}
			totalCents += item.AmountCents
			if !seen[cat] {
				seen[cat] = true
				categories = append(categories, cat)
			}
		}
		if e.AmountCents <= 0 {
			e.AmountCents = totalCents
		}
		if strings.TrimSpace(e.Category) == "" {
			e.Category = strings.Join(categories, ", ")
		}
	}
	if e.AmountCents <= 0 {
		return errors.New("amount must be positive")
	}
	tx, err := r.DB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	var bID any = nil
	if e.BranchID > 0 {
		bID = e.BranchID
	}
	result, err := tx.ExecContext(ctx, `INSERT INTO transactions(operator_id,kind,amount_cents,category,note,occurred_at,branch_id) VALUES(?,?,?,?,?,?,?)`, e.OperatorID, e.Kind, e.AmountCents, e.Category, e.Note, e.OccurredAt.UTC(), bID)
	if err != nil {
		return err
	}
	transactionID, err := result.LastInsertId()
	if err != nil {
		return err
	}
	if len(e.Items) > 0 {
		for _, item := range e.Items {
			cat := strings.TrimSpace(item.Category)
			itemName := strings.TrimSpace(item.ItemName)
			itemType := item.ItemType
			if itemType == "" {
				itemType = "SERVICE"
			}
			var barberID any = nil
			if item.BarberID > 0 {
				barberID = item.BarberID
			}
			var bundleID any = nil
			if item.BundleID > 0 {
				bundleID = item.BundleID
			}
			if _, err := tx.ExecContext(ctx, `INSERT INTO transaction_items(transaction_id,category,item_name,amount_cents,item_type,barber_id,discount_amount,commission_earned,bundle_id) VALUES(?,?,?,?,?,?,?,?,?)`,
				transactionID, cat, itemName, item.AmountCents, itemType, barberID, item.DiscountAmount, item.CommissionEarned, bundleID); err != nil {
				return err
			}
		}
	} else if strings.TrimSpace(e.Category) != "" {
		if _, err := tx.ExecContext(ctx, `INSERT INTO transaction_items(transaction_id,category,item_name,amount_cents,item_type) VALUES(?,?,?,?,'SERVICE')`, transactionID, strings.TrimSpace(e.Category), "", e.AmountCents); err != nil {
			return err
		}
	}
	if _, err = tx.ExecContext(ctx, `INSERT INTO audit_events(actor_id,action,entity_type,entity_id,details) VALUES(?,'transaction.created','transaction',?,'')`, e.OperatorID, transactionID); err != nil {
		return err
	}
	return tx.Commit()
}

func (r Repository) Reverse(ctx context.Context, actorID, transactionID int64, reason string, occurredAt time.Time) error {
	reason = strings.TrimSpace(reason)
	if len(reason) < 5 || len(reason) > 250 {
		return errors.New("reason must be 5 to 250 characters")
	}
	tx, err := r.DB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	var original Entry
	err = tx.QueryRowContext(ctx, `SELECT operator_id,kind,amount_cents,category,note,reversal_of_id FROM transactions WHERE id=?`, transactionID).Scan(&original.OperatorID, &original.Kind, &original.AmountCents, &original.Category, &original.Note, &original.ReversalOfID)
	if err == sql.ErrNoRows {
		return errors.New("transaction not found")
	}
	if err != nil {
		return err
	}
	if original.ReversalOfID.Valid {
		return errors.New("a reversal cannot be reversed")
	}
	var existing int
	if err := tx.QueryRowContext(ctx, `SELECT COUNT(*) FROM transactions WHERE reversal_of_id=?`, transactionID).Scan(&existing); err != nil {
		return err
	}
	if existing > 0 {
		return errors.New("transaction was already reversed")
	}
	reversalKind := "expense"
	if original.Kind == "expense" {
		reversalKind = "income"
	}
	note := fmt.Sprintf("Reversal of #%d: %s", transactionID, reason)
	result, err := tx.ExecContext(ctx, `INSERT INTO transactions(operator_id,kind,amount_cents,category,note,occurred_at,reversal_of_id) VALUES(?,?,?,?,?,?,?)`, original.OperatorID, reversalKind, original.AmountCents, "Reversal: "+original.Category, note, occurredAt.UTC(), transactionID)
	if err != nil {
		return err
	}
	reversalID, err := result.LastInsertId()
	if err != nil {
		return err
	}
	itemRows, err := tx.QueryContext(ctx, `SELECT category, item_name, amount_cents FROM transaction_items WHERE transaction_id=?`, transactionID)
	if err == nil {
		type revItem struct {
			category, itemName string
			cents              int64
		}
		var revItems []revItem
		for itemRows.Next() {
			var it revItem
			if err := itemRows.Scan(&it.category, &it.itemName, &it.cents); err == nil {
				revItems = append(revItems, it)
			}
		}
		_ = itemRows.Close()
		for _, it := range revItems {
			if _, err := tx.ExecContext(ctx, `INSERT INTO transaction_items(transaction_id, category, item_name, amount_cents) VALUES(?,?,?,?)`, reversalID, it.category, it.itemName, it.cents); err != nil {
				return err
			}
		}
	}
	details := fmt.Sprintf("reversal_id=%d; reason=%s", reversalID, reason)
	if _, err := tx.ExecContext(ctx, `INSERT INTO audit_events(actor_id,action,entity_type,entity_id,details) VALUES(?,'transaction.reversed','transaction',?,?)`, actorID, transactionID, details); err != nil {
		return err
	}
	return tx.Commit()
}

func (r Repository) List(ctx context.Context, userID int64, isAdmin bool, from, to time.Time, page, pageSize int) (Page, error) {
	if err := validateWindow(from, to); err != nil {
		return Page{}, err
	}
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 25
	}
	where := ` WHERE t.occurred_at>=? AND t.occurred_at<?`
	args := []any{from.UTC(), to.UTC()}
	if !isAdmin {
		where += ` AND t.operator_id=?`
		args = append(args, userID)
	}
	var result Page
	if err := r.DB.QueryRowContext(ctx, `SELECT COUNT(*),COALESCE(SUM(CASE WHEN kind='income' THEN amount_cents ELSE 0 END),0),COALESCE(SUM(CASE WHEN kind='expense' THEN amount_cents ELSE 0 END),0) FROM transactions t`+where, args...).Scan(&result.Total, &result.Summary.IncomeCents, &result.Summary.ExpenseCents); err != nil {
		return Page{}, err
	}
	result.Pages = (result.Total + pageSize - 1) / pageSize
	if result.Pages == 0 {
		result.Pages = 1
	}
	if page > result.Pages {
		page = result.Pages
	}
	result.Page = page
	query := `SELECT t.id,t.operator_id,u.display_name,t.kind,t.amount_cents,t.category,t.note,t.occurred_at,t.reversal_of_id,(SELECT rev.id FROM transactions rev WHERE rev.reversal_of_id=t.id) FROM transactions t JOIN users u ON u.id=t.operator_id` + where + ` ORDER BY t.occurred_at DESC,t.id DESC LIMIT ? OFFSET ?`
	queryArgs := append(append([]any{}, args...), pageSize, (page-1)*pageSize)
	rows, err := r.DB.QueryContext(ctx, query, queryArgs...)
	if err != nil {
		return Page{}, err
	}
	defer rows.Close()
	for rows.Next() {
		var entry Entry
		if err := rows.Scan(&entry.ID, &entry.OperatorID, &entry.OperatorName, &entry.Kind, &entry.AmountCents, &entry.Category, &entry.Note, &entry.OccurredAt, &entry.ReversalOfID, &entry.ReversedByID); err != nil {
			return Page{}, err
		}
		result.Entries = append(result.Entries, entry)
	}
	if err := rows.Err(); err != nil {
		return Page{}, err
	}
	if err := populateItems(ctx, r.DB, result.Entries); err != nil {
		return Page{}, err
	}
	return result, nil
}

func (r Repository) All(ctx context.Context, userID int64, isAdmin bool, from, to time.Time) ([]Entry, error) {
	if err := validateWindow(from, to); err != nil {
		return nil, err
	}
	query := `SELECT t.id,t.operator_id,u.display_name,t.kind,t.amount_cents,t.category,t.note,t.occurred_at,t.reversal_of_id,(SELECT rev.id FROM transactions rev WHERE rev.reversal_of_id=t.id) FROM transactions t JOIN users u ON u.id=t.operator_id WHERE t.occurred_at>=? AND t.occurred_at<?`
	args := []any{from.UTC(), to.UTC()}
	if !isAdmin {
		query += ` AND t.operator_id=?`
		args = append(args, userID)
	}
	query += ` ORDER BY t.occurred_at DESC,t.id DESC`
	rows, err := r.DB.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var entries []Entry
	for rows.Next() {
		var entry Entry
		if err := rows.Scan(&entry.ID, &entry.OperatorID, &entry.OperatorName, &entry.Kind, &entry.AmountCents, &entry.Category, &entry.Note, &entry.OccurredAt, &entry.ReversalOfID, &entry.ReversedByID); err != nil {
			return nil, err
		}
		entries = append(entries, entry)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if err := populateItems(ctx, r.DB, entries); err != nil {
		return nil, err
	}
	return entries, nil
}

func populateItems(ctx context.Context, db *sql.DB, entries []Entry) error {
	if len(entries) == 0 {
		return nil
	}
	ids := make([]any, len(entries))
	placeholders := make([]string, len(entries))
	for i, entry := range entries {
		ids[i] = entry.ID
		placeholders[i] = "?"
	}
	query := `SELECT id, transaction_id, category, item_name, amount_cents FROM transaction_items WHERE transaction_id IN (` + strings.Join(placeholders, ",") + `) ORDER BY id ASC`
	rows, err := db.QueryContext(ctx, query, ids...)
	if err != nil {
		return err
	}
	defer rows.Close()
	itemMap := make(map[int64][]Item)
	for rows.Next() {
		var it Item
		if err := rows.Scan(&it.ID, &it.TransactionID, &it.Category, &it.ItemName, &it.AmountCents); err != nil {
			return err
		}
		itemMap[it.TransactionID] = append(itemMap[it.TransactionID], it)
	}
	if err := rows.Err(); err != nil {
		return err
	}
	for i := range entries {
		entries[i].Items = itemMap[entries[i].ID]
	}
	return nil
}

func validateWindow(from, to time.Time) error {
	if !to.After(from) || to.Sub(from) > 366*2*24*time.Hour {
		return errors.New("report window must be ordered and not exceed two years")
	}
	return nil
}
