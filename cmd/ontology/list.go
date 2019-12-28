package main

import (
	"context"
	"fmt"
	"log"

	"github.com/spf13/cobra"

	"github.com/uswitch/ontology/pkg/store"
	"github.com/uswitch/ontology/pkg/store/gremlin"
	"github.com/uswitch/ontology/pkg/types"
	"github.com/uswitch/ontology/pkg/types/entity"
	_ "github.com/uswitch/ontology/pkg/types/entity/v1"
	_ "github.com/uswitch/ontology/pkg/types/relation"
	_ "github.com/uswitch/ontology/pkg/types/relation/v1"
)

/*
Defined in dot.go
var (
	searchDepth       int
	traverseDirection string
	constrain         []string
)*/

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List entities by traversing the graph",
	Args:  cobra.ArbitraryArgs,
	Run: func(cmd *cobra.Command, args []string) {
		s, err := gremlin.NewLocalServer(serverURL)
		if err != nil {
			log.Fatalf("failed to connect to store: %v", err)
		}

		typesToList := []types.ID{entity.ID}

		if len(args) > 0 {
			typesToList = make([]types.ID, len(args))
			for idx, arg := range args {
				typesToList[idx] = types.ID(arg)
			}
		}

		allInsts := []types.Instance{}

		if rootID == "" {
			if allInsts, err = s.ListByType(context.Background(), typesToList, store.ListByTypeOptions{IncludeSubclasses: true}); err != nil {
				log.Fatalf("error getting relations: %v", err)
			}
		} else {
			constrainIDs := make([]types.ID, len(constrain))

			for idx, str := range constrain {
				constrainIDs[idx] = types.ID(str)
			}

			if allInsts, err = s.ListFromByType(
				context.Background(), types.ID(rootID), typesToList,
				store.ListFromByTypeOptions{
					ListByTypeOptions: store.ListByTypeOptions{IncludeSubclasses: true},

					MaxDepth:        searchDepth,
					Direction:       store.TraverseDirectionFrom(traverseDirection),
					ConstrainByType: constrainIDs,
				}); err != nil {
				log.Fatalf("error getting relations: %v", err)
			}
		}

		for _, inst := range allInsts {
			fmt.Println(inst.ID())
		}

	},
}

func init() {
	listCmd.Flags().StringVar(
		&rootID, "root-id",
		"", "id of a vertex to start with",
	)
	listCmd.Flags().IntVar(&searchDepth, "search-depth", 2, "How many relations to traverse before we stop")
	listCmd.Flags().StringVar(
		&traverseDirection, "traverse-direction",
		"out", "What direction to traverse the graph when rooting on an entity",
	)
	listCmd.Flags().StringSliceVar(
		&constrain, "constrain",
		[]string{}, "Constraint the relations that will be traversed to only these types",
	)

	rootCmd.AddCommand(listCmd)
}
