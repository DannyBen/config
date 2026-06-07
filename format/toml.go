package format

import "github.com/dannyben/config/format/tomldoc"

type tomlDocument struct{}

func (tomlDocument) ArrayAdd(source, key string, values []string) (string, error) {
	return tomldoc.ArrayAdd(source, key, values)
}

func (tomlDocument) ArrayDel(source, key string, values []string) (string, error) {
	return tomldoc.ArrayDel(source, key, values)
}

func (tomlDocument) Delete(source, key string, selectors []string) (string, error) {
	return tomldoc.Delete(source, key, selectors)
}

func (tomlDocument) DeleteIfEmpty(source, key string) (string, error) {
	return tomldoc.DeleteIfEmpty(source, key)
}

func (tomlDocument) Get(source, key string) (string, error) {
	return tomldoc.Get(source, key)
}

func (tomlDocument) GetIn(source, collection string, selectors []string, key string) (string, error) {
	return tomldoc.GetIn(source, collection, selectors, key)
}

func (tomlDocument) Dump(source, key string) (string, error) {
	return tomldoc.Dump(source, key)
}

func (tomlDocument) Set(source, key, value string) (string, error) {
	return tomldoc.Set(source, key, value)
}

func (tomlDocument) SetArray(source, key string, values []string) (string, error) {
	return tomldoc.SetArray(source, key, values)
}

func (tomlDocument) SetIn(source, collection, on, key, value string) (string, error) {
	return tomldoc.SetIn(source, collection, on, key, value)
}

func (tomlDocument) SetInArray(source, collection, on, key string, values []string) (string, error) {
	return tomldoc.SetInArray(source, collection, on, key, values)
}

func (tomlDocument) SetInString(source, collection, on, key, value string) (string, error) {
	return tomldoc.SetInString(source, collection, on, key, value)
}

func (tomlDocument) SetString(source, key, value string) (string, error) {
	return tomldoc.SetString(source, key, value)
}

func (tomlDocument) Unset(source, key string) (string, error) {
	return tomldoc.Unset(source, key)
}

func (tomlDocument) UnsetIn(source, collection string, selectors []string, key string) (string, error) {
	return tomldoc.UnsetIn(source, collection, selectors, key)
}

func (tomlDocument) List(source, key string) ([]Entry, error) {
	entries, err := tomldoc.List(source, key)
	if err != nil {
		return nil, err
	}

	out := make([]Entry, 0, len(entries))
	for _, entry := range entries {
		out = append(out, Entry{Key: entry.Key, Value: entry.Value})
	}
	return out, nil
}
