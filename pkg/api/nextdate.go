package api

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"
)

const maxDateIterations = 5000

// NextDate рассчитывает следующую дату выполнения задачи.
// Формат дат: 20060102. Возвращаемая дата всегда строго больше now.
func NextDate(now time.Time, dstart string, repeat string) (string, error) {
	const layout = "20060102"

	start, err := time.Parse(layout, dstart)
	if err != nil {
		return "", fmt.Errorf("invalid start date: %w", err)
	}

	repeat = strings.TrimSpace(repeat)
	if repeat == "" {
		return "", nil
	}

	parts := strings.Fields(repeat)
	if len(parts) == 0 {
		return "", fmt.Errorf("empty repeat")
	}

	switch parts[0] {
	case "y":
		if len(parts) != 1 {
			return "", fmt.Errorf("invalid yearly repeat format")
		}
		current := start
		for i := 0; i < maxDateIterations; i++ {
			// всегда делаем хотя бы один шаг по правилу
			current = current.AddDate(1, 0, 0)

			// специальный случай 29 февраля -> 1 марта в невисокосном году
			if start.Month() == time.February && start.Day() == 29 &&
				current.Month() == time.February && current.Day() == 28 {
				current = current.AddDate(0, 0, 1)
			}

			if current.After(now) {
				return current.Format(layout), nil
			}
		}
		return "", fmt.Errorf("cannot find next yearly date")

	case "d":
		if len(parts) != 2 {
			return "", fmt.Errorf("invalid daily repeat format")
		}
		n, err := strconv.Atoi(parts[1])
		if err != nil || n <= 0 || n > 400 {
			return "", fmt.Errorf("invalid daily repeat number")
		}
		current := start
		for i := 0; i < maxDateIterations; i++ {
			// всегда делаем хотя бы один шаг по правилу
			current = current.AddDate(0, 0, n)
			if current.After(now) {
				return current.Format(layout), nil
			}
		}
		return "", fmt.Errorf("cannot find next daily date")

	case "w":
		if len(parts) != 2 {
			return "", fmt.Errorf("invalid weekly repeat format")
		}
		days, err := parseCSVInts(parts[1])
		if err != nil {
			return "", fmt.Errorf("invalid weekly days: %w", err)
		}
		allowed := map[int]struct{}{}
		for _, d := range days {
			if d < 1 || d > 7 {
				return "", fmt.Errorf("weekday out of range: %d", d)
			}
			allowed[d] = struct{}{}
		}

		startScan := start
		minStart := now.AddDate(0, 0, 1)
		if startScan.Before(minStart) {
			startScan = minStart
		}
		current := startScan

		for i := 0; i < maxDateIterations; i++ {
			wd := int(current.Weekday())
			if wd == 0 {
				wd = 7
			}
			if _, ok := allowed[wd]; ok && current.After(now) {
				return current.Format(layout), nil
			}
			current = current.AddDate(0, 0, 1)
		}
		return "", fmt.Errorf("cannot find next weekly date")

	case "m":
		if len(parts) < 2 || len(parts) > 3 {
			return "", fmt.Errorf("invalid monthly repeat format")
		}

		daysSpec, err := parseCSVInts(parts[1])
		if err != nil {
			return "", fmt.Errorf("invalid monthly days: %w", err)
		}
		for _, d := range daysSpec {
			if !((d >= 1 && d <= 31) || d == -1 || d == -2) {
				return "", fmt.Errorf("monthly day out of range: %d", d)
			}
		}

		var monthsSpec []int
		if len(parts) == 3 {
			monthsSpec, err = parseCSVInts(parts[2])
			if err != nil {
				return "", fmt.Errorf("invalid monthly months: %w", err)
			}
			for _, m := range monthsSpec {
				if m < 1 || m > 12 {
					return "", fmt.Errorf("month out of range: %d", m)
				}
			}
		}

		allowedMonths := map[int]struct{}{}
		if len(monthsSpec) > 0 {
			for _, m := range monthsSpec {
				allowedMonths[m] = struct{}{}
			}
		}

		startScan := start
		minStart := now.AddDate(0, 0, 1)
		if startScan.Before(minStart) {
			startScan = minStart
		}
		current := startScan

		for i := 0; i < maxDateIterations; i++ {
			month := int(current.Month())
			if len(allowedMonths) > 0 {
				if _, ok := allowedMonths[month]; !ok {
					current = current.AddDate(0, 0, 1)
					continue
				}
			}

			if matchesMonthlyDay(current, daysSpec) && current.After(now) {
				return current.Format(layout), nil
			}
			current = current.AddDate(0, 0, 1)
		}
		return "", fmt.Errorf("cannot find next monthly date")

	default:
		return "", fmt.Errorf("unknown repeat rule: %s", parts[0])
	}
}

func parseCSVInts(s string) ([]int, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil, fmt.Errorf("empty list")
	}
	parts := strings.Split(s, ",")
	res := make([]int, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			return nil, fmt.Errorf("empty item")
		}
		v, err := strconv.Atoi(p)
		if err != nil {
			return nil, err
		}
		res = append(res, v)
	}
	return res, nil
}

func matchesMonthlyDay(d time.Time, daysSpec []int) bool {
	year, month, day := d.Date()
	firstNextMonth := time.Date(year, month+1, 1, 0, 0, 0, 0, d.Location())
	lastOfMonth := firstNextMonth.AddDate(0, 0, -1).Day()

	for _, spec := range daysSpec {
		switch {
		case spec > 0:
			if day == spec {
				return true
			}
		case spec == -1:
			if day == lastOfMonth {
				return true
			}
		case spec == -2:
			if day == lastOfMonth-1 {
				return true
			}
		}
	}
	return false
}

// nextDateHandler обрабатывает запросы к /api/nextdate.
func nextDateHandler(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()

	nowStr := strings.TrimSpace(query.Get("now"))
	date := strings.TrimSpace(query.Get("date"))
	repeat := strings.TrimSpace(query.Get("repeat"))

	var now time.Time
	const layout = "20060102"

	if nowStr == "" {
		now = time.Now()
	} else {
		var err error
		now, err = time.Parse(layout, nowStr)
		if err != nil {
			// Неверный now — возвращаем пустой ответ.
			return
		}
	}

	next, err := NextDate(now, date, repeat)
	if err != nil {
		// При ошибке форматирования или правила просто ничего не возвращаем.
		return
	}

	if next != "" {
		_, _ = w.Write([]byte(next))
	}
}
