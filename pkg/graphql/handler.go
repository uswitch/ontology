package graphql

import (
	"context"
	"fmt"

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

func operationType(doc *ast.Document) (string, error) {
	var operation *ast.OperationDefinition

	for _, definition := range doc.Definitions {
		switch definition := definition.(type) {
		case *ast.OperationDefinition:
			if operation != nil {
				return "", fmt.Errorf("We only support a single operation at the moment")
			}
			operation = definition
		}
	}

	if operation == nil {
		return "", fmt.Errorf("didn't find an operation")
	}

	return operation.GetOperation(), nil
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

		schema, _ := p.Schema()

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

		opType, err := operationType(AST)
		if err != nil {
			return sendAndReturn(ctx, ch, nil, gqlerrors.FormatErrors(err))
		}

		switch opType {
		case ast.OperationTypeQuery:
			result := graphql.Execute(graphql.ExecuteParams{
				Schema:        schema,
				AST:           AST,
				OperationName: op.OperationName,
				Args:          op.Variables,
				Context:       p.AddValuesTo(ctx),
			})

			return sendAndReturn(ctx, ch, result.Data, result.Errors)
		default:
			return sendAndReturn(ctx, ch, nil, gqlerrors.FormatErrors(fmt.Errorf("implemented operation type: %s", opType)))
		}
	}
}
