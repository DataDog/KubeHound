package tasks

import (
	"context"
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

type csvMarkdownParams struct {
	headers []string
}

// csvMarkdownCmd returns a cobra command that converts a CSV file to a Markdown table.
func csvMarkdownCmd() *cobra.Command {
	var params csvMarkdownParams

	cmd := &cobra.Command{
		Use:     "markdown",
		Aliases: []string{"md"},
		Short:   "Convert a CSV file to a Markdown table",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runCsvMarkdown(cmd.Context(), args, params)
		},
	}

	cmd.Flags().StringSliceVar(&params.headers, "headers", []string{}, "the headers to print")

	return cmd
}

func runCsvMarkdown(_ context.Context, args []string, params csvMarkdownParams) error {
	// Check arguments.
	if len(args) == 0 {
		return fmt.Errorf("missing csv file argument")
	}

	// Initialize the CSV reader.
	var csvReader *csv.Reader
	switch args[0] {
	case "-":
		csvReader = csv.NewReader(os.Stdin)
	default:
		file, err := os.Open(args[0])
		if err != nil {
			return fmt.Errorf("unable to open file: %w", err)
		}
		defer file.Close()

		csvReader = csv.NewReader(file)
	}

	// Read some records.
	csvHeaders, err := csvReader.Read()
	if err != nil {
		return fmt.Errorf("unable to read headers: %w", err)
	}

	// Index the headers.
	headerIndex := make(map[string]int)
	for i, header := range csvHeaders {
		headerIndex[header] = i
	}

	// Index the columns.
	columnIndexes := []int{}

	// Determine which headers to use
	headersToUse := csvHeaders
	if len(params.headers) > 0 {
		// Validate custom headers
		for _, header := range params.headers {
			if _, ok := headerIndex[header]; !ok {
				return fmt.Errorf("header not found: %s", header)
			}
		}
		headersToUse = params.headers
	}

	// Index the columns
	for _, header := range headersToUse {
		columnIndexes = append(columnIndexes, headerIndex[header])
	}

	// Print the markdown table header
	fmt.Println(strings.Join(headersToUse, "|"))

	// Print the markdown separator row
	separators := make([]string, len(headersToUse))
	for i := range headersToUse {
		separators[i] = "---"
	}
	fmt.Println(strings.Join(separators, "|"))

	// Read the rest of the records.
	for {
		record, err := csvReader.Read()
		if err != nil {
			if errors.Is(err, io.EOF) {
				// End of file.
				break
			}

			return fmt.Errorf("unable to read record: %w", err)
		}

		// Print the record.
		var row []string
		for _, columnIndex := range columnIndexes {
			row = append(row, record[columnIndex])
		}
		fmt.Println(strings.Join(row, "|"))
	}

	return nil
}
