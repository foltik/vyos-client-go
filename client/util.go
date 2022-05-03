package client

import "fmt"

func flatten(result *[][]string, value any, path string) error {
	switch value.(type) {
	case map[string]any:
		tree := value.(map[string]any)

		if len(tree) == 0 {
			*result = append(*result, []string{path, ""})
		}

		for k, v := range tree {
			subpath := path
			if len(subpath) > 0 {
				subpath += " "
			}
			subpath += k

			err := flatten(result, v, subpath)
			if err != nil {
				return err
			}
		}

	case map[string]string:
		tree := value.(map[string]string)

		if len(tree) == 0 {
			*result = append(*result, []string{path, ""})
		}

		for k, v := range tree {
			subpath := path
			if len(subpath) > 0 {
				subpath += " "
			}
			subpath += k

			err := flatten(result, v, subpath)
			if err != nil {
				return err
			}
		}

	case []any:
		array := value.([]any)

		for _, v := range array {
			err := flatten(result, v, path)
			if err != nil {
				return err
			}
		}

	case []string:
		array := value.([]string)

		for _, v := range array {
			err := flatten(result, v, path)
			if err != nil {
				return err
			}
		}

	case string:
		*result = append(*result, []string{path, value.(string)})

	default:
		return fmt.Errorf("%s: invalid type %T", path, value)
	}

	return nil
}

// Flatten a multi level object into a flat list of {key, value} pairs
func Flatten(tree map[string]any) ([][]string, error) {
	res := [][]string{}
	err := flatten(&res, tree, "")
	return res, err
}
