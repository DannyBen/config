package format

import "config/format/yamldoc"

type yamlDocument struct{}

func (yamlDocument) Delete(source, key string, selectors []string) (string, error) {
	return yamldoc.Delete(source, key, selectors)
}

func (yamlDocument) Get(source, key string) (string, error) {
	return yamldoc.Get(source, key)
}

func (yamlDocument) GetIn(source, collection string, selectors []string, key string) (string, error) {
	return yamldoc.GetIn(source, collection, selectors, key)
}

func (yamlDocument) Set(source, key, value string) (string, error) {
	return yamldoc.Set(source, key, value)
}

func (yamlDocument) SetArray(source, key string, values []string) (string, error) {
	return yamldoc.SetArray(source, key, values)
}

func (yamlDocument) SetIn(source, collection, on, key, value string) (string, error) {
	return yamldoc.SetIn(source, collection, on, key, value)
}

func (yamlDocument) SetInArray(source, collection, on, key string, values []string) (string, error) {
	return yamldoc.SetInArray(source, collection, on, key, values)
}

func (yamlDocument) SetInString(source, collection, on, key, value string) (string, error) {
	return yamldoc.SetInString(source, collection, on, key, value)
}

func (yamlDocument) SetString(source, key, value string) (string, error) {
	return yamldoc.SetString(source, key, value)
}

func (yamlDocument) Unset(source, key string) (string, error) {
	return yamldoc.Unset(source, key)
}

func (yamlDocument) UnsetIn(source, collection string, selectors []string, key string) (string, error) {
	return yamldoc.UnsetIn(source, collection, selectors, key)
}

func (yamlDocument) List(source, key string) ([]Entry, error) {
	entries, err := yamldoc.List(source, key)
	if err != nil {
		return nil, err
	}

	out := make([]Entry, 0, len(entries))
	for _, entry := range entries {
		out = append(out, Entry{Key: entry.Key, Value: entry.Value})
	}
	return out, nil
}
