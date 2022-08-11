package parser

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

// TODO: Не информативные ошибки в контроле данных.
// Тут надо возвращать смысл, а на уровень выше добавлять контекст

func Snils(data string) (string, error) {
	// FIXME: в функции 3 цикла for, можно обойтись одним прохождением по массиву
	var fSnils string
	for _, b := range data {
		if b >= '0' && b <= '9' {
			fSnils += string(b)
		}
	}

	if len(fSnils) != 11 {
		return fSnils, fmt.Errorf("invalid snils length: %s", data)
	}

	letters := strings.Split(fSnils, "")
	snilsArr := make([]int, 0, len(letters))
	for _, letter := range letters {
		number, _ := strconv.Atoi(letter) // проверка на ошибки не проводится, все не цифры отсеяли выше
		snilsArr = append(snilsArr, number)
	}

	hashSum := 0
	hashLen := len(fSnils) - 2

	for i, v := range snilsArr[:hashLen] {
		hashSum += v * (hashLen - i)
	}

	checksumInt := hashSum % 101

	var checksum string
	if checksumInt == 100 {
		// Частный случай, например если hashSum = 201 то 201 % 101 = 100, а для 100 корректное контрольное число "00"
		checksum = "00"
	} else {
		checksum = strconv.Itoa(checksumInt)
	}

	if len(checksum) == 1 {
		// если контрольное число однозначное, то добавляем в начало ноль
		checksum = "0" + checksum
	}

	if checksum != fSnils[hashLen:] {
		return fSnils, fmt.Errorf("invalid snils, incrorrect checksum: %s", string(data))
	}

	return fSnils, nil
}

func Date(data string) (time.Time, error) {
	var newStr string
	for _, b := range data {
		if b >= '0' && b <= '9' {
			newStr += string(b)
		}
	}
	t, err := time.Parse("02012006", newStr)
	if err != nil {
		return time.Time{}, err
	}
	return t, nil
}

func String(data string) (string, error) {
	if data == "" {
		return "", fmt.Errorf("invalid string: %s", data)
	}
	return data, nil
}

func Int(data string) (int, error) {
	var newStr string
	for _, b := range data {
		if b >= '0' && b <= '9' {
			newStr += string(b)
		}
	}
	if newStr == "" {
		return 0, fmt.Errorf("invalid int: %s", data)
	}
	count, err := strconv.Atoi(newStr)
	if err != nil {
		return 0, err
	}
	return count, nil
}

func Year(data string) (int, error) {
	var newStr string
	for _, b := range data {
		if b >= '0' && b <= '9' {
			newStr += string(b)
		}
	}
	if len(newStr) != 4 {
		return 0, fmt.Errorf("invalid year: %s", data)
	}
	year, err := strconv.Atoi(newStr)
	if err != nil {
		return 0, err
	}
	return year, nil
}

func Semester(data string) (int, error) {
	var newStr string
	for _, b := range data {
		if b >= '1' && b <= '2' {
			newStr += string(b)
		}
	}
	if newStr == "" {
		return 0, fmt.Errorf("invalid semester: %s", data)
	}
	semester, err := strconv.Atoi(newStr)
	if err != nil {
		return 0, err
	}
	return semester, nil
}
