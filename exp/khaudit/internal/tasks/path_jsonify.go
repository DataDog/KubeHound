package tasks

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"strings"

	"github.com/DataDog/KubeHound/exp/khaudit/internal/domain/kubehound"
	"github.com/spf13/cobra"
)

func pathJsonifyCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "jsonify",
		Aliases: []string{"j", "json"},
		Short:   "Convert an HexTuples path to a JSON object",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runPathJsonify(cmd.Context(), args)
		},
	}

	return cmd
}

type pathElement struct {
	Label string
}

func runPathJsonify(_ context.Context, args []string) error {
	// Check arguments.
	if len(args) != 1 {
		return fmt.Errorf("expected 1 argument, got %d", len(args))
	}

	// Initialize the scanner.
	var scanner *bufio.Scanner
	if args[0] == "-" {
		scanner = bufio.NewScanner(os.Stdin)
	} else {
		file, err := os.Open(args[0])
		if err != nil {
			return fmt.Errorf("error opening file: %w", err)
		}
		defer file.Close()

		scanner = bufio.NewScanner(file)
	}

	// Parse the attack paths.
	attackPaths := make([]kubehound.AttackPath, 0)
	for scanner.Scan() {
		// Read the line.
		line := scanner.Text()
		if err := scanner.Err(); err != nil {
			return fmt.Errorf("error reading stdin: %w", err)
		}
		if !strings.HasPrefix(line, "[") {
			return fmt.Errorf("invalid line: %s", line)
		}

		// Decode the hex tuple.
		var path kubehound.AttackPath
		if err := json.Unmarshal([]byte(line), &path); err != nil {
			return fmt.Errorf("error unmarshalling path: %w", err)
		}

		// Append the path to the list.
		attackPaths = append(attackPaths, path)
	}

	// Prepare the JSON encoder.
	encoder := json.NewEncoder(os.Stdout)

	// Iterate over the attack paths.
	for _, path := range attackPaths {
		// Initialize the current subject and properties.
		chain := []map[string]string{}
		currentSubject := ""
		properties := make(map[string]string)

		// Iterate over the tuples.
		for _, tuple := range path {
			// If the subject is not the current subject, append the properties to the chain.
			if currentSubject != tuple.Subject {
				// If the current subject is not empty, append the properties to the chain.
				if currentSubject != "" {
					chain = append(chain, properties)
				}

				// Set the current subject.
				currentSubject = tuple.Subject
				// Reset the properties for the next subject.
				properties = map[string]string{
					"@id": tuple.Subject,
				}
			}

			// Decode the tuple.
			switch {
			case strings.HasPrefix(tuple.Predicate, "urn:property:"):
				property := strings.TrimPrefix(tuple.Predicate, "urn:property:")
				switch {
				case property == "class", property == "label":
					switch {
					case strings.HasPrefix(tuple.Subject, "urn:vertex:"):
						properties["@context"] = "https://kubehound.io/schemas/v1/vertices#" + tuple.Value
					case strings.HasPrefix(tuple.Subject, "urn:edge:"):
						properties["@context"] = "https://kubehound.io/schemas/v1/edges#" + tuple.Value
					default:
						slog.Warn("unknown class", "class", tuple.Value)
						continue
					}
				case strings.HasPrefix(tuple.Subject, "urn:edge:"):
					// Edge properties.
					switch property {
					case "in":
						properties["@in"] = tuple.Value
					case "out":
						properties["@out"] = tuple.Value
					default:
						addIfNotEmpty(properties, property, tuple.Value)
					}
				default:
					addIfNotEmpty(properties, property, tuple.Value)
				}
			default:
				slog.Warn("unknown predicate", "predicate", tuple.Predicate)
				continue
			}
		}

		// Add the last properties to the chain.
		if currentSubject != "" {
			chain = append(chain, properties)
		}

		// Encode the chain.
		if err := encoder.Encode(chain); err != nil {
			return fmt.Errorf("error encoding exploitation chain: %w", err)
		}
	}

	return nil
}

func addIfNotEmpty(properties map[string]string, key string, value string) {
	if value != "" {
		properties[key] = value
	}
}
