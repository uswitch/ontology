package graphql

import (
	"context"
	"fmt"
	"log"

	"github.com/graphql-go/graphql"
	"github.com/graphql-go/graphql/gqlerrors"
	"github.com/graphql-go/graphql/language/ast"
	"github.com/graphql-go/graphql/language/parser"
	"github.com/graphql-go/graphql/language/source"

	"github.com/uswitch/ontology/pkg/audit"
	graphqlws "github.com/uswitch/ontology/pkg/graphql/ws"
)

func errorsToStrings(errors []gqlerrors.FormattedError) []string {
	out := make([]string, len(errors))

	for idx, err := range errors {
		out[idx] = err.Error()
	}

	return out
}

func sendAndReturn(ctx context.Context, ch chan *graphqlws.OperationResult, data interface{}, errors []gqlerrors.FormattedError) (chan *graphqlws.OperationResult, error) {
	gerr := &graphqlws.OperationResult{
		Data:   data,
		Errors: errorsToStrings(errors),
	}

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case ch <- gerr:
	}

	close(ch)

	return ch, nil
}

func operationFrom(doc *ast.Document) (*ast.OperationDefinition, error) {
	var operation *ast.OperationDefinition

	for _, definition := range doc.Definitions {
		switch definition := definition.(type) {
		case *ast.OperationDefinition:
			if operation != nil {
				return nil, fmt.Errorf("We only support a single operation at the moment")
			}
			operation = definition
		}
	}

	if operation == nil {
		return nil, fmt.Errorf("didn't find an operation")
	}

	return operation, nil
}

func WSHandler(p *provider, auditLogger audit.Logger) graphqlws.OnOperationFunc {
	return func(ctx context.Context, op graphqlws.OperationParams) (chan *graphqlws.OperationResult, error) {
		ch := make(chan *graphqlws.OperationResult, 1)

		auditLogger.Log(ctx, audit.AuditData{
			"query":          op.Query,
			"variables":      op.Variables,
			"operation_name": op.OperationName,
		})

		p.SyncOnce(ctx)

		schema, streamFields, _ := p.Schema()

		source := source.NewSource(&source.Source{
			Body: []byte(op.Query),
			Name: "GraphQL request",
		})

		// parse the source
		AST, err := parser.Parse(parser.ParseParams{Source: source})
		if err != nil {
			return sendAndReturn(ctx, ch, nil, gqlerrors.FormatErrors(err))
		}

		validationResult := graphql.ValidateDocument(&schema, AST, nil)

		if !validationResult.IsValid {
			return sendAndReturn(ctx, ch, nil, validationResult.Errors)
		}

		opDef, err := operationFrom(AST)
		if err != nil {
			return sendAndReturn(ctx, ch, nil, gqlerrors.FormatErrors(err))
		}

		switch opDef.GetOperation() {
		case ast.OperationTypeQuery:
			result := graphql.Execute(graphql.ExecuteParams{
				Schema:        schema,
				AST:           AST,
				OperationName: op.OperationName,
				Args:          op.Variables,
				Context:       p.AddValuesTo(ctx),
			})

			return sendAndReturn(ctx, ch, result.Data, result.Errors)
		case ast.OperationTypeSubscription:
			fieldNames := []string{}

			for _, selection := range opDef.SelectionSet.Selections {
				switch selection := selection.(type) {
				case *ast.Field:
					fieldNames = append(fieldNames, selection.Name.Value)
				default:
					log.Printf("unknown selectino: %v", selection)
				}

			}

			if len(fieldNames) > 1 {
				return sendAndReturn(ctx, ch, nil, gqlerrors.FormatErrors(fmt.Errorf("can only have one field")))
			}

			streamField, ok := streamFields[fieldNames[0]]
			if !ok {
				return sendAndReturn(ctx, ch, nil, gqlerrors.FormatErrors(fmt.Errorf("no stream field for %s", fieldNames[0])))
			}

			stream, err := streamField.Stream(graphql.ResolveParams{
				Source:  nil,
				Context: p.AddValuesTo(ctx),
				Args:    op.Variables,
			})
			if err != nil {
				return sendAndReturn(ctx, ch, nil, gqlerrors.FormatErrors(err))
			}

			go func() {
				for {
					select {
					case <-ctx.Done():
						close(ch)
						return
					case fieldValue := <-stream:
						result := graphql.Execute(graphql.ExecuteParams{
							Root: map[string]interface{}{
								fieldNames[0]: fieldValue,
							},
							Schema:        schema,
							AST:           AST,
							OperationName: op.OperationName,
							Args:          op.Variables,
							Context:       p.AddValuesTo(ctx),
						})

						opResult := &graphqlws.OperationResult{
							Data:   result.Data,
							Errors: errorsToStrings(result.Errors),
						}

						select {
						case <-ctx.Done():
							close(ch)
							return
						case ch <- opResult:
						}
					}
				}
			}()

			return ch, nil
		default:
			return sendAndReturn(ctx, ch, nil, gqlerrors.FormatErrors(fmt.Errorf("implemented operation type: %s", opDef.GetOperation())))
		}
	}
}
