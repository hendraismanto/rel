package rel

type Changer interface {
	Build(changes *Changes)
}

func BuildChanges(changers ...Changer) Changes {
	changes := Changes{
		Fields: make(map[string]int),
		Assoc:  make(map[string]int),
	}

	for i := range changers {
		changers[i].Build(&changes)
	}

	return changes
}

// TODO: assoc changes
// Use Assoc fields in Changes?
// Table name not stored here, but handled by repo logic.
// TODO: handle deleteion
//	- Answer: Changes should be forward only operation, no delete change is supported (use changeset instead).
// Implement iterator to be used by adapter api?
// Not safe to be used multiple time. some operation my alter changes data.
type Changes struct {
	Fields       map[string]int // TODO: not copy friendly
	Changes      []Change
	Assoc        map[string]int
	AssocChanges []AssocChanges
	constraints  constraints
}

type AssocChanges struct {
	Changes []Changes
	// if nil, has many associations will be cleared.
	StaleIDs []interface{}
}

func (c Changes) Empty() bool {
	return len(c.Changes) == 0
}

func (c Changes) Get(field string) (Change, bool) {
	if index, ok := c.Fields[field]; ok {
		return c.Changes[index], true
	}

	return Change{}, false
}

func (c *Changes) Set(ch Change) {
	if index, exist := c.Fields[ch.Field]; exist {
		c.Changes[index] = ch
	} else {
		c.Fields[ch.Field] = len(c.Changes)
		c.Changes = append(c.Changes, ch)
	}
}

func (c Changes) GetValue(field string) (interface{}, bool) {
	var (
		ch, ok = c.Get(field)
	)

	return ch.Value, ok
}

func (c *Changes) SetValue(field string, value interface{}) {
	c.Set(Set(field, value))
}

func (c Changes) GetAssoc(field string) (AssocChanges, bool) {
	if index, ok := c.Assoc[field]; ok {
		return c.AssocChanges[index], true
	}

	return AssocChanges{}, false
}

func (c *Changes) SetAssoc(field string, chs ...Changes) {
	if index, exist := c.Assoc[field]; exist {
		c.AssocChanges[index].Changes = chs
	} else {
		c.appendAssoc(field, AssocChanges{
			Changes: chs,
		})
	}
}

func (c *Changes) SetStaleAssoc(field string, ids []interface{}) {
	if index, exist := c.Assoc[field]; exist {
		c.AssocChanges[index].StaleIDs = ids
	} else {
		c.appendAssoc(field, AssocChanges{
			StaleIDs: ids,
		})
	}
}

func (c *Changes) appendAssoc(field string, ac AssocChanges) {
	c.Assoc[field] = len(c.AssocChanges)
	c.AssocChanges = append(c.AssocChanges, ac)
}

type ChangeOp int

const (
	ChangeSetOp ChangeOp = iota
	ChangeIncOp
	ChangeDecOp
	ChangeFragmentOp
)

type Change struct {
	Type  ChangeOp
	Field string
	Value interface{}
}

func (c Change) Build(changes *Changes) {
	changes.Set(c)
}

func Set(field string, value interface{}) Change {
	return Change{
		Type:  ChangeSetOp,
		Field: field,
		Value: value,
	}
}

func Inc(field string) Change {
	return IncBy(field, 1)
}

func IncBy(field string, n int) Change {
	return Change{
		Type:  ChangeIncOp,
		Field: field,
		Value: n,
	}
}

func Dec(field string) Change {
	return DecBy(field, 1)
}

func DecBy(field string, n int) Change {
	return Change{
		Type:  ChangeDecOp,
		Field: field,
		Value: n,
	}
}

func ChangeFragment(raw string, args ...interface{}) Change {
	return Change{
		Type:  ChangeFragmentOp,
		Field: raw,
		Value: args,
	}
}
