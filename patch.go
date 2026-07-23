package comparator

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
)

// ApplyJSONPatch applies an RFC 6902 JSON Patch to a document and returns the
// modified document. It supports the "add", "remove", "replace", "move", "copy",
// and "test" operations.
//
// The document is first normalized to JSON values (objects become
// map[string]any and arrays become []any) by round-tripping through
// encoding/json, so doc must be JSON-serializable. The returned value uses those
// same JSON types.
//
// Patches produced by GetJSONPatch use Go field names by default; generate them
// with WithFieldNaming(JSONTagNaming) when you intend to apply them to
// JSON-shaped documents so the pointers line up with the object keys.
//
// A malformed operation, an unreachable path, or an unsupported op returns an
// error wrapping ErrInvalidPatch. A failed "test" operation also returns an
// error wrapping ErrInvalidPatch.
func ApplyJSONPatch(doc any, patch []JSONPatchOperation) (any, error) {
	root, err := toJSONValue(doc)
	if err != nil {
		return nil, fmt.Errorf("%w: normalize document: %v", ErrInvalidPatch, err)
	}

	for i, op := range patch {
		root, err = applyOp(root, op)
		if err != nil {
			return nil, fmt.Errorf("operation %d (%s %s): %w", i, op.Op, op.Path, err)
		}
	}
	return root, nil
}

func toJSONValue(v any) (any, error) {
	data, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}
	var out any
	if err := json.Unmarshal(data, &out); err != nil {
		return nil, err
	}
	return out, nil
}

func applyOp(root any, op JSONPatchOperation) (any, error) {
	switch op.Op {
	case "add":
		return setAt(root, parsePointer(op.Path), op.Value, true)
	case "replace":
		return setAt(root, parsePointer(op.Path), op.Value, false)
	case "remove":
		v, err := removeAt(root, parsePointer(op.Path))
		return v, err
	case "test":
		got, err := getAt(root, parsePointer(op.Path))
		if err != nil {
			return nil, err
		}
		want, err := toJSONValue(op.Value)
		if err != nil {
			return nil, fmt.Errorf("%w: normalize test value: %v", ErrInvalidPatch, err)
		}
		if !Equal(got, want) {
			return nil, fmt.Errorf("%w: test failed at %q", ErrInvalidPatch, op.Path)
		}
		return root, nil
	case "move":
		return moveOrCopy(root, op, true)
	case "copy":
		return moveOrCopy(root, op, false)
	default:
		return nil, fmt.Errorf("%w: unsupported op %q", ErrInvalidPatch, op.Op)
	}
}

func moveOrCopy(root any, op JSONPatchOperation, remove bool) (any, error) {
	from := parsePointer(op.From)
	val, err := getAt(root, from)
	if err != nil {
		return nil, err
	}
	if remove {
		root, err = removeAt(root, from)
		if err != nil {
			return nil, err
		}
	}
	return setAt(root, parsePointer(op.Path), val, true)
}

// parsePointer splits an RFC 6901 JSON Pointer into unescaped reference tokens.
// The empty pointer ("") yields no tokens and refers to the whole document.
func parsePointer(pointer string) []string {
	if pointer == "" {
		return nil
	}
	pointer = strings.TrimPrefix(pointer, "/")
	tokens := strings.Split(pointer, "/")
	for i, t := range tokens {
		t = strings.ReplaceAll(t, "~1", "/")
		t = strings.ReplaceAll(t, "~0", "~")
		tokens[i] = t
	}
	return tokens
}

func getAt(root any, tokens []string) (any, error) {
	cur := root
	for _, tok := range tokens {
		switch node := cur.(type) {
		case map[string]any:
			v, ok := node[tok]
			if !ok {
				return nil, fmt.Errorf("%w: missing key %q", ErrInvalidPatch, tok)
			}
			cur = v
		case []any:
			idx, err := arrayIndex(tok, len(node), false)
			if err != nil {
				return nil, err
			}
			cur = node[idx]
		default:
			return nil, fmt.Errorf("%w: cannot descend into %q", ErrInvalidPatch, tok)
		}
	}
	return cur, nil
}

func setAt(root any, tokens []string, value any, add bool) (any, error) {
	value, err := toJSONValue(value)
	if err != nil {
		return nil, fmt.Errorf("%w: normalize value: %v", ErrInvalidPatch, err)
	}
	if len(tokens) == 0 {
		return value, nil
	}

	parent, err := getAt(root, tokens[:len(tokens)-1])
	if err != nil {
		return nil, err
	}
	last := tokens[len(tokens)-1]

	switch node := parent.(type) {
	case map[string]any:
		if !add {
			if _, ok := node[last]; !ok {
				return nil, fmt.Errorf("%w: cannot replace missing key %q", ErrInvalidPatch, last)
			}
		}
		node[last] = value
		return root, nil
	case []any:
		return setInArray(root, tokens, node, last, value, add)
	default:
		return nil, fmt.Errorf("%w: cannot set %q on %T", ErrInvalidPatch, last, parent)
	}
}

func setInArray(root any, tokens []string, arr []any, last string, value any, add bool) (any, error) {
	if add {
		idx := len(arr)
		if last != "-" {
			var err error
			idx, err = arrayIndex(last, len(arr)+1, true)
			if err != nil {
				return nil, err
			}
		}
		arr = append(arr, nil)
		copy(arr[idx+1:], arr[idx:])
		arr[idx] = value
		return replaceArray(root, tokens[:len(tokens)-1], arr)
	}

	idx, err := arrayIndex(last, len(arr), false)
	if err != nil {
		return nil, err
	}
	arr[idx] = value
	return root, nil
}

func removeAt(root any, tokens []string) (any, error) {
	if len(tokens) == 0 {
		return nil, fmt.Errorf("%w: cannot remove whole document", ErrInvalidPatch)
	}

	parent, err := getAt(root, tokens[:len(tokens)-1])
	if err != nil {
		return nil, err
	}
	last := tokens[len(tokens)-1]

	switch node := parent.(type) {
	case map[string]any:
		if _, ok := node[last]; !ok {
			return nil, fmt.Errorf("%w: cannot remove missing key %q", ErrInvalidPatch, last)
		}
		delete(node, last)
		return root, nil
	case []any:
		idx, err := arrayIndex(last, len(node), false)
		if err != nil {
			return nil, err
		}
		node = append(node[:idx], node[idx+1:]...)
		return replaceArray(root, tokens[:len(tokens)-1], node)
	default:
		return nil, fmt.Errorf("%w: cannot remove %q from %T", ErrInvalidPatch, last, parent)
	}
}

// replaceArray writes a rebuilt slice back into its parent, since appending to
// or shrinking a []any produces a new header that the parent must reference.
func replaceArray(root any, parentTokens []string, arr []any) (any, error) {
	if len(parentTokens) == 0 {
		return arr, nil
	}
	grand, err := getAt(root, parentTokens[:len(parentTokens)-1])
	if err != nil {
		return nil, err
	}
	key := parentTokens[len(parentTokens)-1]
	switch node := grand.(type) {
	case map[string]any:
		node[key] = arr
	case []any:
		idx, err := arrayIndex(key, len(node), false)
		if err != nil {
			return nil, err
		}
		node[idx] = arr
	default:
		return nil, fmt.Errorf("%w: cannot attach array at %q", ErrInvalidPatch, key)
	}
	return root, nil
}

func arrayIndex(tok string, length int, allowEnd bool) (int, error) {
	idx, err := strconv.Atoi(tok)
	if err != nil {
		return 0, fmt.Errorf("%w: invalid array index %q", ErrInvalidPatch, tok)
	}
	limit := length
	if !allowEnd {
		limit = length - 1
	}
	if idx < 0 || idx > limit {
		return 0, fmt.Errorf("%w: array index %d out of range", ErrInvalidPatch, idx)
	}
	return idx, nil
}
