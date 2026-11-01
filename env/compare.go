package env

// DiffKind says how a key differs between two files.
type DiffKind int

const (
	DiffSame DiffKind = iota
	DiffChanged
	DiffOnlyLeft
	DiffOnlyRight
)

// DiffRow is one key in a comparison of two files.
type DiffRow struct {
	Key      string
	Left     string
	Right    string
	Kind     DiffKind
	IsSecret bool
}

// Compare lists every key of left and right: keys of left in file order,
// followed by keys that only exist in right. For duplicate keys the last
// definition wins, as with most dotenv loaders.
func Compare(left, right *EnvFile) []DiffRow {
	lv, lOrder := lastValues(left)
	rv, rOrder := lastValues(right)

	var rows []DiffRow
	for _, k := range lOrder {
		row := DiffRow{Key: k, Left: lv[k], IsSecret: IsSecret(k)}
		if r, ok := rv[k]; ok {
			row.Right = r
			if r == row.Left {
				row.Kind = DiffSame
			} else {
				row.Kind = DiffChanged
			}
		} else {
			row.Kind = DiffOnlyLeft
		}
		rows = append(rows, row)
	}
	for _, k := range rOrder {
		if _, ok := lv[k]; !ok {
			rows = append(rows, DiffRow{Key: k, Right: rv[k], Kind: DiffOnlyRight, IsSecret: IsSecret(k)})
		}
	}
	return rows
}

func lastValues(f *EnvFile) (map[string]string, []string) {
	values := make(map[string]string)
	var order []string
	if f == nil {
		return values, order
	}
	for _, v := range f.Variables {
		if v.Disabled {
			continue // commented out: treated as not set
		}
		if _, ok := values[v.Key]; !ok {
			order = append(order, v.Key)
		}
		values[v.Key] = v.Value
	}
	return values, order
}
