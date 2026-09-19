package httpserver

import (
	"encoding/json"
	"fmt"
	"html/template"
	"strconv"
	"strings"

	"github.com/mekari/pos-phoenix/internal/backoffice"
)

type CatalogItem struct {
	Name        string `json:"name"`
	Amount      int64  `json:"amount"` // in whole units (e.g. 40000)
	AmountLabel string `json:"amount_label"`
	BundleID    int64  `json:"bundle_id,omitempty"`
}

type CatalogCategory struct {
	Name  string        `json:"name"`
	Kind  string        `json:"kind"` // "income" or "expense"
	Items []CatalogItem `json:"items"`
}

var DefaultCatalog = []CatalogCategory{
	{
		Name: "Haircut Services",
		Kind: "income",
		Items: []CatalogItem{
			{Name: "Haircut", Amount: 40000, AmountLabel: "40K"},
		},
	},
	{
		Name: "Add Ons",
		Kind: "income",
		Items: []CatalogItem{
			{Name: "Vitamin", Amount: 3000, AmountLabel: "3K"},
			{Name: "Shampoo & Blow Dry", Amount: 20000, AmountLabel: "20K"},
			{Name: "Beard Trim", Amount: 15000, AmountLabel: "15K"},
		},
	},
	{
		Name: "Chemical Services",
		Kind: "income",
		Items: []CatalogItem{
			{Name: "Color Basic", Amount: 60000, AmountLabel: "60K"},
			{Name: "Color Fashion", Amount: 220000, AmountLabel: "220K"},
			{Name: "Smoothing", Amount: 200000, AmountLabel: "200K"},
			{Name: "Perm", Amount: 200000, AmountLabel: "200K"},
			{Name: "Root Lift & Down", Amount: 200000, AmountLabel: "200K"},
		},
	},
	{
		Name: "Hair Treatment",
		Kind: "income",
		Items: []CatalogItem{
			{Name: "Hair & Face Mask", Amount: 30000, AmountLabel: "30K"},
			{Name: "Smooth Keratin", Amount: 200000, AmountLabel: "200K"},
		},
	},
	{
		Name: "Product",
		Kind: "income",
		Items: []CatalogItem{
			{Name: "Death Waterbased", Amount: 120000, AmountLabel: "120K"},
			{Name: "Death Clay", Amount: 120000, AmountLabel: "120K"},
			{Name: "Death Powder", Amount: 80000, AmountLabel: "80K"},
			{Name: "Serum Pack", Amount: 50000, AmountLabel: "50K"},
			{Name: "Hair Tonic", Amount: 20000, AmountLabel: "20K"},
		},
	},
	{
		Name: "Expense",
		Kind: "expense",
		Items: []CatalogItem{
			{Name: "Supplies & Inventory", Amount: 0, AmountLabel: "Custom"},
			{Name: "Rent", Amount: 0, AmountLabel: "Custom"},
			{Name: "Utilities", Amount: 0, AmountLabel: "Custom"},
			{Name: "Salary / Commission", Amount: 0, AmountLabel: "Custom"},
			{Name: "Operational / Maintenance", Amount: 0, AmountLabel: "Custom"},
			{Name: "Other Expense", Amount: 0, AmountLabel: "Custom"},
		},
	},
}

func CatalogJSON() template.HTML {
	data, _ := json.Marshal(DefaultCatalog)
	return template.HTML(data)
}

func CatalogJS() template.JS {
	data, _ := json.Marshal(DefaultCatalog)
	return template.JS(data)
}

func formatAmountLabel(amount int64) string {
	if amount >= 1000 {
		if amount%1000 == 0 {
			return fmt.Sprintf("%dK", amount/1000)
		}
		return fmt.Sprintf("%.1fK", float64(amount)/1000)
	}
	return strconv.FormatInt(amount, 10)
}

func BuildCatalog(items []backoffice.CatalogItem, bundles []backoffice.DiscountBundle) []CatalogCategory {
	var incomeCategories []CatalogCategory
	categoryMap := make(map[string]int)

	for _, it := range items {
		if !it.IsActive {
			continue
		}
		catName := strings.TrimSpace(it.Category)
		if catName == "" {
			if it.ItemType == "PRODUCT" {
				catName = "Product"
			} else {
				catName = "Haircut Services"
			}
		}
		amt := it.PriceCents / 100
		if amt <= 0 && it.PriceCents > 0 {
			amt = it.PriceCents
		}
		catItem := CatalogItem{
			Name:        it.Name,
			Amount:      amt,
			AmountLabel: formatAmountLabel(amt),
		}

		idx, exists := categoryMap[catName]
		if !exists {
			idx = len(incomeCategories)
			categoryMap[catName] = idx
			incomeCategories = append(incomeCategories, CatalogCategory{
				Name: catName,
				Kind: "income",
			})
		}
		incomeCategories[idx].Items = append(incomeCategories[idx].Items, catItem)
	}

	// If no items in database, fallback to default income categories
	if len(incomeCategories) == 0 {
		for _, c := range DefaultCatalog {
			if c.Kind == "income" {
				incomeCategories = append(incomeCategories, c)
			}
		}
	}

	// Add bundles category if any active bundles
	var bundleCat CatalogCategory
	hasBundles := false
	if len(bundles) > 0 {
		bundleCat = CatalogCategory{
			Name: "Bundling",
			Kind: "income",
		}
		for _, b := range bundles {
			if !b.IsActive || b.Type != "BUNDLE" {
				continue
			}
			amt := b.Value / 100
			if amt <= 0 && b.Value > 0 {
				amt = b.Value
			}
			bundleCat.Items = append(bundleCat.Items, CatalogItem{
				Name:        b.Name,
				Amount:      amt,
				AmountLabel: formatAmountLabel(amt),
				BundleID:    b.ID,
			})
		}
		if len(bundleCat.Items) > 0 {
			hasBundles = true
		}
	}

	cats := make([]CatalogCategory, 0, len(incomeCategories)+2)
	cats = append(cats, incomeCategories...)
	if hasBundles {
		cats = append(cats, bundleCat)
	}

	// Add Expense categories from DefaultCatalog
	for _, c := range DefaultCatalog {
		if c.Kind == "expense" {
			cats = append(cats, c)
		}
	}

	return cats
}

func DynamicCatalogJSON(cats []CatalogCategory) template.HTML {
	data, _ := json.Marshal(cats)
	return template.HTML(data)
}

func DynamicCatalogJS(cats []CatalogCategory) template.JS {
	data, _ := json.Marshal(cats)
	return template.JS(data)
}
