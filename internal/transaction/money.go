package transaction

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

var moneyPattern = regexp.MustCompile(`^[0-9]+(?:\.[0-9]{1,2})?$`)

func ParseCents(value string) (int64, error) {
	value = strings.TrimSpace(value)
	if !moneyPattern.MatchString(value) {
		return 0, errors.New("enter a positive amount with up to two decimals")
	}
	parts := strings.SplitN(value, ".", 2)
	whole, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		return 0, errors.New("amount is too large")
	}
	fraction := int64(0)
	if len(parts) == 2 {
		f := parts[1]
		if len(f) == 1 {
			f += "0"
		}
		fraction, _ = strconv.ParseInt(f, 10, 64)
	}
	if whole > (1<<63-1-fraction)/100 {
		return 0, errors.New("amount is too large")
	}
	cents := whole*100 + fraction
	if cents <= 0 {
		return 0, errors.New("amount must be positive")
	}
	return cents, nil
}

func FormatCents(cents int64) string {
	sign := ""
	if cents < 0 {
		sign = "-"
		cents = -cents
	}
	return fmt.Sprintf("%s%d.%02d", sign, cents/100, cents%100)
}
