package filter

import "strings"

type Filters []*Filter

func (filters Filters) SingleParams() (params []string) {
	for _, filter := range filters {
		params = append(params, filter.String())
	}
	return
}

func (filters Filters) Params() (params []string) {
	txt := filters.String()
	if txt == "" {
		return
	}
	params = append(params, "-filter_complex", txt)
	return
}

func (filters Filters) String() string {
	return strings.Join(filters.SingleParams(), ";")
}
