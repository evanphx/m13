package value

type mapEntry struct {
	hash  uint64
	key   Value
	value Value
}

type mapEntries struct {
	used    int
	fill    int
	entries []*mapEntry
}

/*
type Map struct {
	Object

	entries *mapEntries
}
*/

func NewMap(env Env) *Map {
	m := &Map{
		entries: &mapEntries{
			entries: make([]*mapEntry, 8),
		},
	}

	m.SetClass(env.MapClass())

	return m
}

func nextIndex(i, perturb uint64) (uint64, uint64) {
	return (i << 2) + i + perturb + 1, perturb >> 5
}

var deletedEntry = &mapEntry{}

func (m *mapEntries) findEntry(k Value, h uint64) (*mapEntry, int, bool) {
	var (
		mask    = uint64(len(m.entries) - 1)
		idx     = h & mask
		i       = idx
		perturb = h
		entry   = m.entries[idx]
	)

	for {
		// New key
		if entry == nil {
			return nil, int(idx), false
		}

		if entry.hash == h && Compare(k, entry.key) {
			return entry, int(idx), true
		}

		i, perturb = nextIndex(i, perturb)
		idx = i & mask

		entry = m.entries[idx]
	}
}

func (m *mapEntries) set(up *mapEntry) {
	var (
		mask    = uint64(len(m.entries) - 1)
		idx     = up.hash & mask
		i       = idx
		perturb = up.hash
		entry   = m.entries[idx]
	)

	for {
		// New key
		if entry == nil {
			m.fill++
			m.entries[idx] = up
			return
		}

		// Replace key
		if entry.hash == up.hash && Compare(up.key, entry.key) {
			m.entries[idx] = up
			return
		}

		i, perturb = nextIndex(i, perturb)
		idx = i & mask

		entry = m.entries[idx]
	}
}

func (m *mapEntries) add(up *mapEntry) {
	var (
		mask    = uint64(len(m.entries) - 1)
		idx     = up.hash & mask
		i       = idx
		perturb = up.hash
		entry   = m.entries[idx]
	)

	for {
		// New key
		if entry == nil {
			m.fill++
			m.entries[idx] = up
			return
		}

		i, perturb = nextIndex(i, perturb)
		idx = i & mask

		entry = m.entries[idx]
	}
}

func (m *mapEntries) filled() bool {
	return (m.fill+1)*3 > len(m.entries)*2
}

func (m *Map) Get(k Value) (Value, bool) {
	h := k.Hash()

	ent, _, ok := m.entries.findEntry(k, h)
	if ok {
		return ent.value, true
	}

	return nil, false
}

func (m *Map) Set(k, v Value) {
	h := k.Hash()

	ent, index, ok := m.entries.findEntry(k, h)
	if ok {
		if ent == deletedEntry {
			m.entries.used++
		}

		// Replace!
		m.entries.entries[index] = &mapEntry{h, k, v}
		return
	}

	if !m.entries.filled() {
		m.entries.fill++
		m.entries.used++
		m.entries.entries[index] = &mapEntry{h, k, v}
		return
	}

	// Grow the table.
	n := m.entries.used * 2

	newTable := &mapEntries{
		entries: make([]*mapEntry, n),
	}

	for _, old := range m.entries.entries {
		if old == nil {
			continue
		}

		newTable.add(old)
	}

	newTable.add(&mapEntry{h, k, v})

	m.entries = newTable
}

func (m *Map) Del(k Value) (Value, bool) {
	h := k.Hash()

	ent, index, ok := m.entries.findEntry(k, h)
	if ok {
		m.entries.used--
		m.entries.entries[index] = deletedEntry
		return ent.value, true
	}

	return nil, false
}
