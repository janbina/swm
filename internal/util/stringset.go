package util

type StringSet map[string]bool

func (m StringSet) SetAll(vals []string) {
	for _, val := range vals {
		m[val] = true
	}
}

func (m StringSet) UnSetAll(vals []string) {
	for _, val := range vals {
		m[val] = false
	}
}

func (m StringSet) Any(vals ...string) bool {
	for _, val := range vals {
		if m[val] {
			return true
		}
	}
	return false
}

func (m StringSet) All(vals ...string) bool {
	for _, val := range vals {
		if !m[val] {
			return false
		}
	}
	return true
}

func (m StringSet) GetActive() []string {
	s := make([]string, 0, len(m))
	for k, v := range m {
		if v {
			s = append(s, k)
		}
	}
	return s
}
