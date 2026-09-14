package httpserver

import (
	"encoding/json"
	"html/template"
)

type CatalogItem struct {
	Name        string `json:"name"`
	Amount      int64  `json:"amount"` // in whole units (e.g. 40000)
	AmountLabel string `json:"amount_label"`
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
