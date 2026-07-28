package fixtures

func computeTotals(rows []Row) Summary {
	var s Summary
	for _, r := range rows {
		s.count++
		s.amount += r.value
		if r.value > s.max {
			s.max = r.value
		}
		if r.value < s.min || s.count == 1 {
			s.min = r.value
		}
	}
	s.average = s.amount / float64(s.count)
	return s
}
