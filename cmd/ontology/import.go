package main

import (
	"bufio"
	"context"
	"log"
	"os"

	"github.com/spf13/cobra"

	"github.com/uswitch/ontology/pkg/store/gremlin"
	"github.com/uswitch/ontology/pkg/types"
	_ "github.com/uswitch/ontology/pkg/types/entity"
	_ "github.com/uswitch/ontology/pkg/types/entity/v1"
	_ "github.com/uswitch/ontology/pkg/types/relation"
	_ "github.com/uswitch/ontology/pkg/types/relation/v1"
)

var importCmd = &cobra.Command{
	Use:   "import",
	Short: "Load JSON lines data",
	Run: func(cmd *cobra.Command, args []string) {
		s, err := gremlin.NewLocalServer(serverURL)
		if err != nil {
			log.Fatalf("failed to connect to store: %v", err)
		}

		instances := []types.Instance{}

		scanner := bufio.NewScanner(os.Stdin)
		for scanner.Scan() {
			text := scanner.Text()

			if text == "" {
				continue
			}

			in, err := types.Parse(text)
			if err != nil {
				log.Fatalf("error parsing: %v", err)
				continue
			}

			instances = append(instances, in)
			//log.Printf("%T %+v", out, out)
		}
		if err := scanner.Err(); err != nil {
			log.Fatalf("reading standard input: %v", err)
		}

		if len(instances) > 0 {
			if err = s.Add(context.Background(), instances...); err != nil {
				log.Fatalf("error adding: %v", err)
			}
		}
	},
}

func init() {
	rootCmd.AddCommand(importCmd)
}
