package store

import (
	"context"
	"encoding/json"
	"fmt"
	"io"

	"github.com/qri-io/jsonschema"
)

const resolutionPath = "/requires_resolution"

type resolutionPair struct {
	ID
	Type ID
}

type PointerTo string

func (ptv PointerTo) String() string { return string(ptv) }
func (ptv PointerTo) Validate(propPath string, data interface{}, errs *[]jsonschema.ValError) {
	if id, ok := data.(string); ok {
		pair := &resolutionPair{
			ID:   ID(id),
			Type: ID(ptv),
		}

		// error with rule /resolution and the pair
		*errs = append(*errs, jsonschema.ValError{
			PropertyPath: propPath,
			RulePath:     resolutionPath,
			InvalidValue: pair,
			Message:      "",
		})
	} else {
		// error as data should be string
		*errs = append(*errs, jsonschema.ValError{
			PropertyPath: propPath,
			RulePath:     "",
			InvalidValue: data,
			Message:      "pointer_to should be a string",
		})
	}

}

func NewPointerTo() jsonschema.Validator {
	return new(PointerTo)
}

func init() {
	jsonschema.RegisterValidator("pointer_to", NewPointerTo)
}

func TypeProperties(ctx context.Context, s Store, typ *Type) (jsonschema.Properties, jsonschema.Required, error) {
	props := jsonschema.Properties{}
	requiredSet := map[string]struct{}{}

	for currType := typ; ; {
		if required, ok := currType.Properties["required"].([]string); ok {
			for _, f := range required {
				requiredSet[f] = struct{}{}
			}
		}

		if spec, ok := currType.Properties["spec"]; ok {
			reader, writer := io.Pipe()

			encoder := json.NewEncoder(writer)
			decoder := json.NewDecoder(reader)

			var currProps jsonschema.Properties

			go func() { encoder.Encode(spec) }()
			err := decoder.Decode(&currProps)
			if err != nil {
				return nil, nil, err
			}

			for k, schema := range currProps {
				if _, ok := props[k]; !ok {
					props[k] = schema
				}
			}
		}

		if nextTypeID, ok := currType.Properties["parent"]; !ok {
			break
		} else {
			nextType, err := s.GetTypeByID(ctx, ID(nextTypeID.(string)))
			if err != nil {
				return nil, nil, err
			}

			currType = nextType
		}
	}

	required := make([]string, len(requiredSet))
	idx := 0
	for f, _ := range requiredSet {
		required[idx] = f
		idx = idx + 1
	}

	return props, jsonschema.Required(required), nil
}

func Validate(ctx context.Context, s Store, thingable Thingable, options ValidateOptions) ([]ValidationError, error) {
	thing := thingable.Thing()

	typ, err := s.GetTypeByID(ctx, thing.Metadata.Type)
	if err != nil {
		return nil, err
	}

	propsSchema, requiredProps, err := TypeProperties(ctx, s, typ)
	if err != nil {
		return nil, err
	}

	props, err := json.Marshal(thing.Properties)
	if err != nil {
		return nil, err
	}

	rootSchema := jsonschema.RootSchema{
		jsonschema.Schema{
			Validators: map[string]jsonschema.Validator{
				"properties": propsSchema,
				"required":   requiredProps,
			},
		},
		"https://github.com/uswitch/ontology",
	}

	errors, err := rootSchema.ValidateBytes(props)
	if err != nil {
		return nil, err
	}

	validationErrors := []ValidationError{}
	for _, error := range errors {
		// this is a bit of a hack as we won't have the store available when
		// we are running PointerTo.Validate. This lets us defer the resolution
		if error.RulePath == resolutionPath {
			if options.Pointers == IgnoreAllPointers {
				continue
			}

			pair, ok := error.InvalidValue.(*resolutionPair)
			if !ok {
				return nil, fmt.Errorf("Failed to get a resolution pair from: %v", error.InvalidValue)
			}

			typ, err := s.GetTypeByID(ctx, pair.Type)
			if err == ErrNotFound {
				validationErrors = append(
					validationErrors,
					ValidationError(fmt.Sprintf("could not find type %s", string(pair.ID))),
				)
				continue
			} else if err != nil {
				return nil, err
			}

			thing, err := s.GetByID(ctx, pair.ID)
			if err == ErrNotFound {
				if options.Pointers != IgnoreMissingPointers {
					validationErrors = append(
						validationErrors,
						ValidationError(fmt.Sprintf("could not find thing %s", string(pair.ID))),
					)
				}
				continue
			} else if err != nil {
				return nil, err
			}

			typeMatches, err := s.IsA(ctx, thing, typ)
			if err != nil {
				return nil, err
			}

			if !typeMatches {
				validationErrors = append(
					validationErrors,
					ValidationError(fmt.Sprintf("%s does not match type %s", string(pair.ID), string(pair.Type))),
				)
			}
		} else {
			validationErrors = append(validationErrors, ValidationError(error.Message))
		}
	}

	return validationErrors, nil
}
