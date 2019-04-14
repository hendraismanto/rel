package grimoire

type primaryKey interface {
	PrimaryKey() (string, interface{})
}

func getPrimaryKey(record interface{}, withValue bool) (string, interface{}) {
	if pk, ok := record.(primaryKey); ok {
		return pk.PrimaryKey()
	}

	value := 0
	if withValue {
		// TODO
	}

	return "id", value
}
