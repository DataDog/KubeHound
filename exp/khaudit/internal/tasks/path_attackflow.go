package tasks

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"slices"
	"strings"

	"github.com/DataDog/KubeHound/exp/khaudit/internal/domain/kubehound"
	"github.com/DataDog/KubeHound/exp/khaudit/internal/domain/mitre"
	"github.com/DataDog/KubeHound/exp/khaudit/internal/domain/mitre/stixv2"
	"github.com/google/uuid"
	"github.com/spf13/cobra"
	"golang.org/x/sync/errgroup"
)

func pathAttackFlowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "attackflow",
		Aliases: []string{"af"},
		Short:   "Convert an HexTuples path to MITRE Attack Flow",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runPathAttackFlow(cmd.Context(), args)
		},
	}

	return cmd
}

func runPathAttackFlow(ctx context.Context, args []string) error {
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

	// Prepare the tuple channel.
	tupleChan := make(chan kubehound.AttackPath, 10)
	chainChan := make(chan []map[string]string, 10)

	// Prepare the error group.
	eg, egCtx := errgroup.WithContext(ctx)

	// Convert to attack flow.
	eg.Go(func() error {
		for {
			select {
			case <-egCtx.Done():
				return nil
			case path, ok := <-tupleChan:
				if !ok {
					// Channel is closed
					close(chainChan)
					return nil
				}

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

				// Send the chain to the chain channel.
				select {
				case <-egCtx.Done():
					return nil
				case chainChan <- chain:
				}
			}
		}
	})

	// Convert to attack flow.
	eg.Go(func() error {
		// Create the bundle.
		bundle := stixv2.NewBundle()

		// Create the attack flow.
		flow := stixv2.NewAttackFlow()

		// Prepare the UUID namespace to generate deterministic UUIDs from JanusGraph identities.
		assetNS := uuid.NewSHA1(uuid.NameSpaceURL, []byte("https://center-for-threat-informed-defense.github.io/attack-flow/language/#attack-asset"))
		actionNS := uuid.NewSHA1(uuid.NameSpaceURL, []byte("https://center-for-threat-informed-defense.github.io/attack-flow/language/#attack-action"))

		// Add the assume breach action as initial action.
		assumeBreachAction := stixv2.NewAttackAction("Assume Breach", "The attacker gains access to the target environment by assuming the identity of a specific user or service account.")
		assumeBreachAction.TechniqueID = "T1609"
		assumeBreachAction.TechniqueRef = "attack-pattern--7b50a1d3-4ca7-45d1-989d-a6503f04bfe1"
		assumeBreachAction.TacticID = "TA0002"
		assumeBreachAction.TacticRef = "x-mitre-tactic--4ca45d45-df4d-4613-8980-bac22d278fa5"
		bundle.AddObject(assumeBreachAction)

		// Add the assume breach action as start ref.
		flow.StartRefs = []string{assumeBreachAction.ID}
		bundle.AddObject(flow)

		// Iterate over the received chains.
		// This will merge all chains into a single attack flow bundle.
	LOOP:
		for {
			select {
			case <-egCtx.Done():
				return nil
			case chain, ok := <-chainChan:
				if !ok {
					// Channel is closed
					break LOOP
				}

				// Prepare the last action ref, we assume linear flow.
				// Re-attach the last action ref to the assume breach action for all chains.
				lastActionRef := assumeBreachAction.ID

				// Iterate over the chain.
				for _, items := range chain {
					// Get the object id.
					oid, ok := items["@id"]
					if !ok {
						slog.Warn("missing @id", "items", items)
						continue
					}

					// Get the object type.
					oc, ok := items["@context"]
					if !ok {
						slog.Warn("missing @context", "items", items)
						continue
					}

					switch {
					case strings.HasPrefix(oc, "https://kubehound.io/schemas/v1/vertices#"):
						assetUUID := fmt.Sprintf("attack-asset--%s", uuid.NewSHA1(assetNS, []byte(oid)))

						// Set the asset to the assume breach action.
						if len(assumeBreachAction.AssetRefs) == 0 {
							assumeBreachAction.AssetRefs = append(assumeBreachAction.AssetRefs, assetUUID)
							assumeBreachAction.AssetRefs = slices.Compact(assumeBreachAction.AssetRefs)
							// Update the last action.
							bundle.AddObject(assumeBreachAction)
						}

						// Check if the asset already exists.
						if _, ok := bundle.Objects[assetUUID]; ok {
							continue
						}

						// Create the asset.
						asset := stixv2.NewAttackAsset(strings.TrimPrefix(oc, "https://kubehound.io/schemas/v1/vertices#"), getAttackAssetDescription(items))
						// Override the ID with deterministic UUID to simplify the lookup.
						asset.ID = assetUUID
						// Register the asset.
						bundle.AddObject(asset)

					case strings.HasPrefix(oc, "https://kubehound.io/schemas/v1/edges#"):
						actionUUID := fmt.Sprintf("attack-action--%s", uuid.NewSHA1(actionNS, []byte(oid)))

						var action *stixv2.AttackAction

						// Check if the action already exists.
						existing, ok := bundle.Objects[actionUUID]
						if !ok {
							// Create the action.
							action = stixv2.NewAttackAction(strings.TrimPrefix(oc, "https://kubehound.io/schemas/v1/edges#"), getAttackActionDescription(items))
							// Override the ID with deterministic UUID to simplify the lookup.
							action.ID = actionUUID
							// Set the technique.
							action.TechniqueID = items["attckTechniqueID"]
							action.TechniqueRef = mitre.TechniqueMapping[items["attckTechniqueID"]]
							// Set the tactic.
							action.TacticID = items["attckTacticID"]
							action.TacticRef = mitre.TacticMapping[items["attckTacticID"]]
						} else {
							action, ok = existing.(*stixv2.AttackAction)
							if !ok {
								slog.Warn("existing object is not an attack action", "object", existing)
								continue
							}
						}

						// Compute effect/asset references.
						assetRefs := make([]string, 0)
						if v, ok := items["@in"]; ok {
							assetRefs = append(assetRefs, fmt.Sprintf("attack-asset--%s", uuid.NewSHA1(assetNS, []byte(v))))
						}
						if v, ok := items["@out"]; ok {
							assetRefs = append(assetRefs, fmt.Sprintf("attack-asset--%s", uuid.NewSHA1(assetNS, []byte(v))))
						}

						// Set the asset refs.
						action.AssetRefs = slices.Compact(assetRefs)

						// Update the effect refs of the last action.
						if lastActionRef != "" {
							lastAction, ok := bundle.Objects[lastActionRef].(*stixv2.AttackAction)
							if ok {
								lastAction.EffectRefs = append(lastAction.EffectRefs, actionUUID)
								lastAction.EffectRefs = slices.Compact(lastAction.EffectRefs)
								// Update the last action.
								bundle.AddObject(lastAction)
							} else {
								slog.Warn("last action not found", "lastActionRef", lastActionRef)
							}
						}

						// Set the last action ref.
						lastActionRef = actionUUID

						// Register/update the action.
						bundle.AddObject(action)
					default:
						slog.Warn("unknown context", "context", oc)
						continue
					}
				}
			}
		}

		// Encode the bundle.
		encoder := json.NewEncoder(os.Stdout)
		if err := encoder.Encode(bundle); err != nil {
			return fmt.Errorf("error encoding attack flow: %w", err)
		}

		return nil
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

func getAttackActionDescription(items map[string]string) string {
	oc, ok := items["@context"]
	if !ok {
		return ""
	}

	attackAction := strings.TrimPrefix(oc, "https://kubehound.io/schemas/v1/edges#")
	switch attackAction {
	case "CE_MODULE_LOAD":
		return "Load a kernel module from within an overprivileged container to breakout into the node."
	case "CE_NSENTER":
		return "Container escape via the nsenter built-in linux program that allows executing a binary into another namespace."
	case "CE_PRIV_MOUNT":
		return "Mount the host disk and gain access to the host via arbitrary filesystem write."
	case "CE_SYS_PTRACE":
		return "Given the requisite capabilities, abuse the legitimate OS debugging mechanisms to escape the container via attaching to a node process."
	case "CE_UMH_CORE_PATTERN":
		return "Arbitrary file writes on the host from a node via the core pattern feature of the kernel."
	case "CE_VAR_LOG_SYMLINK":
		return "Arbitrary file reads on the host from a node via an exposed /var/log mount."
	case "CONTAINER_ATTACH":
		return "The attacker attaches to a running container to gain access."
	case "ENDPOINT_EXPLOIT":
		return "The attacker exploits a known vulnerability in the endpoint to gain access."
	case "EXPLOIT_HOST_READ":
		return "The attacker reads sensitive information from the node."
	case "EXPLOIT_HOST_WRITE":
		return "The attacker alters node filesystem to gain access."
	case "EXPLOIT_HOST_TRAVERSE":
		return "This attack represents the ability to steal a K8s API token from a container via access to a mounted parent volume of the /var/lib/kubelet/pods directory."
	case "IDENTITY_ASSUME":
		return "The attacker assumes the identity of a specific user or service account."
	case "IDENTITY_IMPERSONATE":
		return "The attacker impersonates a specific user or service account."
	case "PERMISSION_DISCOVER":
		return "The attacker discovers the permissions of a specific user or service account."
	case "POD_ATTACH":
		return "The attacker attaches to a running pod to gain access."
	case "POD_CREATE":
		return "The attacker creates a new pod to gain access to the node."
	case "POD_EXEC":
		return "With the correct privileges an attacker can use the Kubernetes API to obtain a shell on a running pod."
	case "POD_PATCH":
		return "With the correct privileges an attacker can use the Kubernetes API to modify certain properties of an existing pod and achieve code execution within the pod."
	case "ROLE_BIND":
		return "A role that grants permission to create or modify (Cluster)RoleBindings can allow an attacker to escalate privileges on a compromised user."
	case "SHARE_PS_NAMESPACE":
		return "Represents a relationship between containers within the same pod that share a process namespace."
	case "TOKEN_BRUTEFORCE":
		return "An attacker can brute force the token of a compromised user."
	case "TOKEN_LIST":
		return "An identity with a role that allows listing secrets can potentially view all the secrets in a specific namespace or in the whole cluster (with ClusterRole)."
	case "TOKEN_STEAL":
		return "This attack represents the ability to steal a K8s API token from an accessible volume."
	case "VOLUME_ACCESS":
		return "This attack represents the ability to access a volume from a pod."
	case "VOLUME_DISCOVER":
		return "The attacker discovers a volume attached to the pod."
	}

	return ""
}

func getAttackAssetDescription(items map[string]string) string {
	oc, ok := items["@context"]
	if !ok {
		return ""
	}

	attackAsset := strings.TrimPrefix(oc, "https://kubehound.io/schemas/v1/vertices#")
	switch attackAsset {
	case "Node":
		return fmt.Sprintf("The node %s is compromised.", items["name"])
	case "Pod":
		return fmt.Sprintf("The pod %s in namespace %s is compromised.", items["name"], items["namespace"])
	case "Container":
		return fmt.Sprintf("The container %s in pod %s in namespace %s is compromised.", items["image"], items["pod"], items["namespace"])
	case "Volume":
		return fmt.Sprintf("The volume %s in namespace %s (%s => %s) is exposed.", items["name"], items["namespace"], items["sourcePath"], items["mountPath"])
	case "PermissionSet":
		return fmt.Sprintf("The permission set %s in namespace %s is exposed.", items["name"], items["namespace"])
	case "Identity":
		return fmt.Sprintf("The identity %s is exposed.", items["name"])
	case "Endpoint":
		return fmt.Sprintf("The endpoint %s is compromised.", items["name"])
	}

	return ""
}
