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
	"golang.org/x/sync/errgroup"
)

type pathFilterParams struct {
	profiles      []string
	not           bool
	dedupProfiles bool
}

func pathFilterCmd() *cobra.Command {
	var params pathFilterParams

	cmd := &cobra.Command{
		Use:     "filter",
		Aliases: []string{"f"},
		Short:   "Filter an HexTuples path",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runPathFilter(cmd.Context(), args, params)
		},
	}

	cmd.Flags().StringSliceVarP(&params.profiles, "profile", "p", []string{}, "Profiles to filter the path")
	cmd.Flags().BoolVarP(&params.not, "not", "n", false, "Negate the filter")
	cmd.Flags().BoolVarP(&params.dedupProfiles, "dedup-profiles", "d", false, "Deduplicate profiles - only keep one path per profile")

	return cmd
}

func runPathFilter(ctx context.Context, args []string, params pathFilterParams) error {
	// Check arguments.
	if len(args) != 1 {
		return fmt.Errorf("expected 1 argument, got %d", len(args))
	}

	// Prepare the profile matchers.
	profileMatchers := [][]string{}
	for _, profile := range params.profiles {
		profileMatchers = append(profileMatchers, strings.Split(profile, PROFILE_STEP_SEPARATOR))
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

	// Prepare the tuple channel.
	tupleChan := make(chan kubehound.AttackPath, 10)

	// Prepare the error group.
	eg, egCtx := errgroup.WithContext(ctx)

	// Convert hextuples to chains.
	eg.Go(func() error {
		// Prepare the encoder.
		encoder := json.NewEncoder(os.Stdout)

		// Store the seen attack profiles
		attackProfilesSeen := [][]string{}

		for {
			select {
			case <-egCtx.Done():
				return nil
			case path, ok := <-tupleChan:
				if !ok {
					// Channel is closed
					return nil
				}

				// Check if the path is acceptable.
				if isAcceptablePath(path, profileMatchers) == params.not {
					// If the path is not acceptable, skip it.
					continue
				}

				// Deduplicate profiles.
				if params.dedupProfiles {
					// Extract the profile from the path.
					profile, err := buildProfileFromPath(path)
					if err != nil {
						slog.Warn("error building profile from path - skipping", "error", err)
						continue
					}
					// If the path has already been seen, skip it
					if isProfileInList(profile, attackProfilesSeen) {
						continue
					}
					// If not seen yet, add it to the list and continue execution.
					attackProfilesSeen = append(attackProfilesSeen, profile)
				}

				// Encode the path.
				if err := encoder.Encode(path); err != nil {
					return fmt.Errorf("error encoding path: %w", err)
				}
			}
		}
	})

	// Parse the attack paths.
	eg.Go(func() error {
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
			select {
			case <-egCtx.Done():
				return fmt.Errorf("context cancelled")
			case tupleChan <- path:
			}
		}

		// Close the channel.
		close(tupleChan)

		return nil
	})

	// Wait for the conversion to complete.
	if err := eg.Wait(); err != nil {
		return fmt.Errorf("error converting to attack flow: %w", err)
	}

	return nil
}

// Extract the profile from the path, as a list of strings. Each element of the array is a profile
// step, in order of execution.
func buildProfileFromPath(path kubehound.AttackPath) ([]string, error) {
	// Initialize the current subject and properties.
	chain := []string{}

	// Iterate over the tuples to extract path profile.
	for _, tuple := range path {
		// Decode the tuple.
		switch {
		case strings.HasPrefix(tuple.Predicate, "urn:property:"):
			property := strings.TrimPrefix(tuple.Predicate, "urn:property:")
			switch {
			case property == "label":
				switch {
				case strings.HasPrefix(tuple.Subject, "urn:vertex:"):
					chain = append(chain, tuple.Value)
				case strings.HasPrefix(tuple.Subject, "urn:edge:"):
					chain = append(chain, tuple.Value)
				default:
					slog.Warn("unknown class", "class", tuple.Value)
					continue
				}
			default:
			}
		default:
			slog.Warn("unknown predicate", "predicate", tuple.Predicate)
			continue
		}
	}

	// If the chain is empty, the path is not acceptable.
	if len(chain) == 0 {
		return nil, fmt.Errorf("empty path")
	}

	return chain, nil
}

// Return true if the profile is in the list of profiles, else false.
// The matching is case-insensitive.
func isProfileInList(profile []string, list [][]string) bool {
Candidate:
	for _, candidate := range list {
		// Check length.
		if len(profile) != len(candidate) {
			continue
		}

		// Check if the profile matched the candidate.
		for i := range candidate {
			if !strings.EqualFold(profile[i], candidate[i]) {
				// If the profile does not match the candidate, skip it. Do not return early, as it
				// would prevent checking the rest of the list.
				continue Candidate
			}
		}

		// If we get here, the path is acceptable.
		return true
	}
	// If the list is empty, the path does not match any profile.
	return false
}

// Return true if the path is acceptable for the given profile matchers, meaning it matches any of them.
func isAcceptablePath(path kubehound.AttackPath, profileMatchers [][]string) bool {
	// Build the profile from the path.
	chain, err := buildProfileFromPath(path)
	if err != nil {
		slog.Info("error building profile from path - skipping", "error", err)
		return false
	}

	// If there are no profiles, the path is acceptable. It is important to check this before checking
	// with profileMatchers because isProfileInList will return false if the list is empty.
	if len(profileMatchers) == 0 {
		return true
	}

	// Check if the chain is acceptable for any profile.
	return isProfileInList(chain, profileMatchers)
}
