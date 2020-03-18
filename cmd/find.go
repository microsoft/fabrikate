package cmd

import (
	"context"
	"errors"
	"fmt"
	"github.com/google/go-github/v28/github"
	"github.com/spf13/cobra"
	"strings"

	log "github.com/sirupsen/logrus"
)

// FindComponent finds fabrikate components in the fabrikate-definitions repository that are related to the given keyword.
func FindComponent(keyword string) error {

	client := github.NewClient(nil)
	ctx := context.Background()
	query := keyword + "+repo:microsoft/fabrikate-definitions"

	results, _, err := client.Search.Code(ctx, query, nil)

	if err != nil || results.CodeResults == nil {
		return err
	}

	components := GetFabrikateComponents(results.CodeResults)

	fmt.Println(fmt.Sprintf("Search results for '%s':", keyword))
	if len(components) == 0 {
		log.Info(fmt.Sprintf("No components were found for '%s'", keyword))
	} else {
		for _, component := range components {
			fmt.Println(component)
		}
	}

	return nil
}

// GetFabrikateComponents returns a unique list of fabrikate components from a github search result
func GetFabrikateComponents(codeResults []github.CodeResult) []string {

	if codeResults == nil {
		return []string{}
	}

	components := []string{}
	uniqueComponents := map[string]bool{}

	for _, result := range codeResults {

		path := *result.Path
		if !strings.HasPrefix(path, "definitions") {
			continue
		}

		pathComponents := strings.Split(path, "/")
		componentName := pathComponents[1]

		if _, ok := uniqueComponents[componentName]; !ok {
			uniqueComponents[componentName] = true
			components = append(components, componentName)
		}
	}

	return components
}

var findCmd = &cobra.Command{
	Use:   "find <keyword>",
	Short: "Find fabrikate components related to the given keyword",
	Long: `Find fabrikate components related to the given keyword
Eg.
$ fab find prometheus
Finds fabrikate components that are related to 'prometheus'.
`,
	RunE: func(cmd *cobra.Command, args []string) (err error) {

		if len(args) != 1 {
			return errors.New("'find' takes one argument")
		}

		return FindComponent(args[0])
	},
}

func init() {
	rootCmd.AddCommand(findCmd)
}
