package main

// Fixed exchange rate: 1 EUR = 1.95583 BGN
const EURtoBGN = 1.95583

func convertToBGN(amount float64, isEuro bool) float64 {
	if isEuro {
		return amount * EURtoBGN
	}
	return amount
}

func convertToEUR(amount float64, isEuro bool) float64 {
	if !isEuro {
		return amount / EURtoBGN
	}
	return amount
}
