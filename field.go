package exceltable

import (
	"reflect"
	"strings"
	"sync"
)

var defaultFieldCache = newFieldCache(defaultRules, defaultPredicates)

type field struct {
	l, r       int
	fieldIndex int
	typ        reflect.Type
	baseType   reflect.Type
	tag        *tag
	rules      map[string]*predicateList // rule name -> predicate list
	children   []*field
}

func newField(typ reflect.Type, rules *ruleList, predMap *predicateMap) (*field, error) {
	if typ.Kind() != reflect.Struct {
		return nil, ErrNotStructType
	}

	return createStructFields(&field{
		l:        0,
		r:        0,
		typ:      reflect.PointerTo(typ),
		baseType: typ,
		tag:      &tag{},
	}, typ, rules, predMap)
}

func createStructFields(fi *field, typ reflect.Type, rules *ruleList, predicates *predicateMap) (*field, error) {
	numField := typ.NumField()
	for i := range numField {
		child, err := createChildFields(fi, i, rules, predicates)
		if err != nil {
			return nil, err
		}
		if child == nil { // skip
			continue
		}

		fi.r += child.r - child.l
		fi.children = append(fi.children, child)
	}

	return fi, nil
}

func createChildFields(parent *field, fieldIndex int, rules *ruleList, predicates *predicateMap) (*field, error) {
	sf := parent.baseType.Field(fieldIndex)
	baseType := walkType(sf.Type)
	if !sf.IsExported() && !(baseType.Kind() == reflect.Struct && sf.Anonymous) {
		return nil, nil // skip
	}

	// Parse the basic tag element.
	tag := parseTag(sf)
	if tag.ignore {
		return nil, nil // skip
	}
	tag.columnName = parent.tag.prefix + tag.columnName
	tag.prefix = parent.tag.prefix + tag.prefix
	if parent.tag.omitZero {
		tag.omitZero = true
	}
	if baseType.Kind() == reflect.Struct && sf.Anonymous {
		tag.inline = true
	}

	// Parse the rule tag elements.
	fieldRules := make(map[string]*predicateList)
	for r := range rules.iter() {
		preds, err := parseRuleTag(sf, r.name, reflect.PointerTo(parent.baseType), predicates)
		if err != nil {
			return nil, err
		}
		fieldRules[r.name] = preds
	}

	if tag.inline {
		return createStructFields(&field{
			l:          parent.r,
			r:          parent.r,
			fieldIndex: fieldIndex,
			typ:        sf.Type,
			baseType:   baseType,
			tag:        tag,
			rules:      fieldRules,
		}, baseType, rules, predicates)
	}

	return &field{
		l:          parent.r,
		r:          parent.r + 1,
		fieldIndex: fieldIndex,
		typ:        sf.Type,
		baseType:   baseType,
		tag:        tag,
		rules:      fieldRules,
	}, nil
}

func parseRuleTag(sf reflect.StructField, tagName string, ptrType reflect.Type, predicates *predicateMap) (*predicateList, error) {
	preds := &predicateList{}

	for key := range strings.SplitSeq(sf.Tag.Get(tagName), ",") {
		switch key {
		case "", "-":
			// ignore
		default:
			if method, ok := ptrType.MethodByName(key); ok {
				pred, err := newPredicate(method.Func, true)
				if err != nil {
					return nil, err
				}

				if !pred.isAcceptableType(sf.Type) {
					return nil, ErrInvalidPredicate
				}

				preds.append(pred)
				break
			}

			if pred, ok := predicates.load(key); ok {
				if !pred.isAcceptableType(sf.Type) {
					return nil, ErrInvalidPredicate
				}

				preds.append(pred)
				break
			}

			return nil, ErrUnknownPredicateKey
		}
	}

	return preds, nil
}

type fieldCache struct {
	sync.Mutex
	v          map[reflect.Type]*field
	rules      *ruleList
	predicates *predicateMap
}

func newFieldCache(rules *ruleList, predicates *predicateMap) *fieldCache {
	return &fieldCache{
		v:          make(map[reflect.Type]*field),
		rules:      rules,
		predicates: predicates,
	}
}

func (c *fieldCache) cache(typ reflect.Type) (*field, error) {
	c.Lock()
	defer c.Unlock()

	cachedField, ok := c.v[typ]
	if ok {
		return cachedField, nil
	}

	field, err := newField(typ, c.rules, c.predicates)
	if err != nil {
		return nil, err
	}
	c.v[typ] = field

	return field, nil
}

// CountByRule counts the number of fields in obj that satisfy the predicates associated with the rule tag name.
func CountByRule[M any](obj *M, tagName string) (int, error) {
	field, err := defaultFieldCache.cache(reflect.TypeFor[M]())
	if err != nil {
		return 0, err
	}

	return countByRuleInternal(reflect.ValueOf(obj), field, tagName)
}

func countByRuleInternal(ptrV reflect.Value, field *field, tagName string) (int, error) {
	v := ptrV.Elem()

	cnt := 0
	for _, child := range field.children {
		preds := child.rules[tagName]
		fieldV := v.Field(child.fieldIndex)
		if preds.callWithReceiver(ptrV, fieldV) {
			cnt += child.r - child.l
			continue
		}

		if child.tag.inline {
			c, err := countByRuleInternal(walkValue(fieldV).Addr(), child, tagName)
			if err != nil {
				return 0, err
			}
			cnt += c
		}
	}

	return cnt, nil
}
