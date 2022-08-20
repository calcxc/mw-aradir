package main

import "strings"

func StripAnd(name string) string {
	if strings.Contains(name, " &") {
		return strings.Replace(name, " &", "", -1)
	} else if strings.Contains(name, "&") {
		return strings.Replace(name, "&", "", -1)
	}
	return name
}

func insertIntoSlice(a []string, index int, values []string) []string {
	if len(a) == index { // nil or empty slice or after last element
		var lines = a
		for _, val := range values {
			lines = append(a, val)
		}
		return lines
	}
	a = append(a[:index+1], a[index:]...)
	for i, val := range values {
		a[index+i] = val
	}

	return a
}
